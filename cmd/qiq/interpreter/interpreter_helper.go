package interpreter

import (
	"QIQ/cmd/qiq/ast"
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/common/os"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/stdlib/outputControl"
	"QIQ/cmd/qiq/runtime/stdlib/variableHandling"
	"QIQ/cmd/qiq/runtime/values"
	"math"
	GoOs "os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// -------------------------------------- Common -------------------------------------- MARK: Common

func (interpreter *Interpreter) Print(str string) {
	if interpreter.outputBufferStack.Len() > 0 {
		interpreter.outputBufferStack.Get(interpreter.outputBufferStack.Len() - 1).Content += str
	} else {
		interpreter.WriteResult(str)
	}
}

func (interpreter *Interpreter) Println(str string) { interpreter.Print(str + "\n") }

func (interpreter *Interpreter) WriteResult(str string) { interpreter.result += str }

func (interpreter *Interpreter) flushOutputBuffers() {
	if interpreter.outputBufferStack.Len() == 0 {
		return
	}

	for interpreter.outputBufferStack.Len() > 0 {
		outputControl.ObEndFlush(runtime.NewContext(interpreter, nil, nil))
	}
}

func (interpreter *Interpreter) processCondition(expr ast.IExpression, env *Environment) (values.RuntimeValue, bool, phpError.Error) {
	slot, err := interpreter.processStmt(expr, env)
	if err != nil {
		return slot.Value, false, err
	}

	boolean, err := variableHandling.BoolVal(slot.Value)
	return slot.Value, boolean, err
}

func (interpreter *Interpreter) lookupVariable(expr ast.IExpression, env *Environment) (string, *values.Slot, phpError.Error) {
	variableName, err := interpreter.varExprToVarName(expr, env)
	if err != nil {
		return "", values.NewVoidSlot(), err
	}

	slot, err := env.LookupVariable(variableName)
	if !interpreter.suppressWarning && err != nil {
		interpreter.PrintError(phpError.NewWarning("%s in %s", strings.TrimPrefix(err.Error(), "Warning: "), expr.GetPosString()))
	}
	return variableName, slot, nil
}

// Convert a variable expression into the interpreted variable name
func (interpreter *Interpreter) varExprToVarName(expr ast.IExpression, env *Environment) (string, phpError.Error) {
	switch expr.GetKind() {
	case ast.SimpleVariableExpr:
		variableNameExpr := expr.(*ast.SimpleVariableExpression).VariableName

		if variableNameExpr.GetKind() == ast.VariableNameExpr {
			return variableNameExpr.(*ast.VariableNameExpression).VariableName, nil
		}

		if variableNameExpr.GetKind() == ast.SimpleVariableExpr {
			variableName, err := interpreter.varExprToVarName(variableNameExpr, env)
			if err != nil {
				return "", err
			}
			slot, err := env.LookupVariable(variableName)
			if err != nil {
				interpreter.PrintError(err)
			}
			valueStr, err := variableHandling.StrVal(slot.Value)
			if err != nil {
				return "", err
			}
			return "$" + valueStr, nil
		}

		variableNameSlot, err := interpreter.processStmt(variableNameExpr, env)
		if err != nil {
			return "", err
		}
		valueStr, err := variableHandling.StrVal(variableNameSlot.Value)
		if err != nil {
			return "", err
		}
		return "$" + valueStr, nil
	case ast.SubscriptExpr:
		return interpreter.varExprToVarName(expr.(*ast.SubscriptExpression).Variable, env)
	case ast.MemberAccessExpr:
		return interpreter.varExprToVarName(expr.(*ast.MemberAccessExpression).Object, env)
	default:
		return "", phpError.NewError("varExprToVarName: Unsupported expression: %s", ast.ToString(expr))
	}
}

func (interpreter *Interpreter) ErrorToString(err phpError.Error) string {
	if (err.GetErrorType() == phpError.WarningPhpError && uint16(interpreter.ini.GetInt("error_reporting"))&phpError.E_WARNING == 0) ||
		(err.GetErrorType() == phpError.ErrorPhpError && uint16(interpreter.ini.GetInt("error_reporting"))&phpError.E_ERROR == 0) ||
		(err.GetErrorType() == phpError.NoticePhpError && uint16(interpreter.ini.GetInt("error_reporting"))&phpError.E_NOTICE == 0) ||
		(err.GetErrorType() == phpError.ParsePhpError && uint16(interpreter.ini.GetInt("error_reporting"))&phpError.E_PARSE == 0) {
		return ""
	}
	return err.GetMessage()
}

func (interpreter *Interpreter) PrintError(err phpError.Error) {
	if errStr := interpreter.ErrorToString(err); errStr == "" {
		return
	} else {
		interpreter.Println("")
		interpreter.Println(errStr)
	}
}

// Scan and process program for function definitions on root level and in compound statements
func (interpreter *Interpreter) scanForFunctionDefinition(statements []ast.IStatement, env *Environment) phpError.Error {
	for _, stmt := range statements {
		if stmt.GetKind() == ast.CompoundStmt {
			interpreter.scanForFunctionDefinition(stmt.(*ast.CompoundStatement).Statements, env)
			continue
		}

		if stmt.GetKind() != ast.FunctionDefinitionStmt {
			continue
		}

		_, err := interpreter.processStmt(stmt, env)
		if err != nil {
			return err
		}
	}
	return nil
}

var literalExprTypeRuntimeValue = map[ast.NodeType]string{
	ast.ArrayLiteralExpr:   "array",
	ast.IntegerLiteralExpr: "int",
	ast.StringLiteralExpr:  "string",
}

func literalExprTypeToRuntimeValue(expr ast.IExpression) (string, phpError.Error) {
	typeStr, found := literalExprTypeRuntimeValue[expr.GetKind()]
	if !found {
		return "", phpError.NewError("literalExprTypeToRuntimeValue: No mapping for type %s", expr.GetKindString())
	}
	return typeStr, nil
}

func checkParameterTypes(runtimeValue values.RuntimeValue, expectedTypes []string) phpError.Error {
	typeStr := values.ToPhpType(runtimeValue)
	if typeStr == "" {
		return phpError.NewError("checkParameterTypes: No mapping for type %s", runtimeValue.GetType())
	}

	for _, expectedType := range expectedTypes {
		if expectedType == "mixed" {
			return nil
		}

		if typeStr == expectedType {
			return nil
		}
	}
	return phpError.NewError("Types do not match")
}

func (interpreter *Interpreter) includeFile(filepathExpr ast.IExpression, env *Environment, include bool, once bool) (*values.Slot, phpError.Error) {
	slot, err := interpreter.processStmt(filepathExpr, env)
	if err != nil {
		return slot, err
	}
	if slot.GetType() == values.NullValue {
		return slot, phpError.NewError("Uncaught ValueError: Path cannot be empty in %s", filepathExpr.GetPosString())
	}

	filename, err := variableHandling.StrVal(slot.Value)
	if err != nil {
		return slot, err
	}

	// Spec: https://phplang.org/spec/10-expressions.html#the-require-operator
	// Once an include file has been included, a subsequent use of require_once on that include file
	// results in a return value of TRUE but nothing else happens.
	if once && slices.Contains(interpreter.includedFiles, filename) && !os.IS_WIN {
		return values.NewBoolSlot(true), nil
	}
	if once && slices.Contains(interpreter.includedFiles, strings.ToLower(filename)) && os.IS_WIN {
		return values.NewBoolSlot(true), nil
	}

	absFilename := filename
	if !common.IsAbsPath(filename) {
		absFilename = common.GetAbsPathForWorkingDir(common.ExtractPath(filepathExpr.GetPosition().File.Filename), filename)
	}

	var functionName string
	if include {
		functionName = "include"
	} else {
		functionName = "require"
	}

	// Spec: https://phplang.org/spec/10-expressions.html#the-require-operator
	// This operator is identical to operator include except that in the case of require,
	// failure to find/open the designated include file produces a fatal error.
	getError := func() (*values.Slot, phpError.Error) {
		if include {
			return values.NewVoidSlot(), phpError.NewWarning(
				"%s(): Failed opening '%s' for inclusion (include_path='%s') in %s",
				functionName, filename, common.ExtractPath(filepathExpr.GetPosition().File.Filename), filepathExpr.GetPosString(),
			)
		} else {
			return values.NewVoidSlot(), phpError.NewError(
				"Uncaught Error: Failed opening required '%s' (include_path='%s') in %s",
				filename, common.ExtractPath(filepathExpr.GetPosition().File.Filename), filepathExpr.GetPosString(),
			)
		}
	}

	if !common.PathExists(absFilename) {
		interpreter.PrintError(phpError.NewWarning(
			"%s(%s): Failed to open stream: No such file or directory in %s",
			functionName, filename, filepathExpr.GetPosString(),
		))
		return getError()
	}

	isCaseSensitiveInclude := interpreter.ini.GetBool("qiq.case_sensitive_include")

	if isCaseSensitiveInclude && os.IS_WIN {
		foundExact, err := common.FileExistsCaseSensitive(absFilename)
		if err != nil {
			return slot, phpError.NewError("%s", err.Error())
		}
		if !foundExact {
			interpreter.PrintError(phpError.NewWarning(
				"%s(%s): Failed to open stream: No such file or directory in %s",
				functionName, filename, filepathExpr.GetPosString(),
			))
			return getError()
		}
	}

	content, fileErr := GoOs.ReadFile(absFilename)
	if fileErr != nil {
		return getError()
	}
	program, parserErr := interpreter.parser.ProduceAST(string(content), absFilename)

	if os.IS_WIN {
		interpreter.includedFiles = append(interpreter.includedFiles, strings.ToLower(absFilename))
	} else {
		interpreter.includedFiles = append(interpreter.includedFiles, absFilename)
	}
	if parserErr != nil {
		return slot, parserErr
	}
	return interpreter.processProgram(program, env)
}

func GetExecutableCreationDate() time.Time {
	exePath, err := GoOs.Executable()
	if err != nil {
		return time.Now()
	}
	info, err := GoOs.Stat(filepath.Clean(exePath))
	if err != nil {
		// println(err.Error())

		return time.Now()
	}
	return info.ModTime()
}

// -------------------------------------- Methods -------------------------------------- MARK: Methods

func MethodDeclToSignature(methodDef *ast.MethodDefinitionStatement) string {
	var signature strings.Builder
	// Method name
	signature.WriteString(methodDef.Name)
	signature.WriteString("(")
	// Parameters
	for paramIndex, param := range methodDef.Params {
		if paramIndex > 0 {
			signature.WriteString(", ")
		}
		// Param types
		if len(param.Type) > 0 {
			signature.WriteString(ParamTypesToSignature(param.Type))
			signature.WriteString(" ")
		}
		// Param name
		signature.WriteString(param.Name)
	}
	signature.WriteString(")")
	// Return type
	if len(methodDef.ReturnType) > 0 {
		signature.WriteString(": ")
		signature.WriteString(ParamTypesToSignature(methodDef.ReturnType))
	}
	return signature.String()
	// F(?string $p): string|int|null
}

func ParamTypesToSignature(paramTypes []string) string {
	signature := ""
	if len(paramTypes) == 2 && slices.Contains(paramTypes, "null") {
		for _, paramType := range paramTypes {
			if paramType == "null" {
				continue
			}
			signature += "?" + paramType
		}
		return signature
	}

	for _, paramType := range paramTypes {
		if len(signature) > 0 {
			signature += "|"
		}
		signature += paramType
	}
	return signature
}

// -------------------------------------- Classes and Interfaces -------------------------------------- MARK: Classes and Interfaces

func (interpreter *Interpreter) AddClass(class string, classDecl *ast.ClassDeclarationStatement) phpError.Error {
	return interpreter.executionContext.AddClass(class, classDecl)
}

func (interpreter *Interpreter) GetClass(class string) (*ast.ClassDeclarationStatement, bool) {
	return interpreter.executionContext.GetClass(class)
}

func (interpreter *Interpreter) GetClasses() []string {
	return interpreter.executionContext.GetClasses()
}

func (interpreter *Interpreter) AddInterface(interfaceName string, interfaceDecl *ast.InterfaceDeclarationStatement) phpError.Error {
	return interpreter.executionContext.AddInterface(interfaceName, interfaceDecl)
}

func (interpreter *Interpreter) GetInterface(interfaceName string) (*ast.InterfaceDeclarationStatement, bool) {
	return interpreter.executionContext.GetInterface(interfaceName)
}

func (interpreter *Interpreter) GetInterfaces() []string {
	return interpreter.executionContext.GetInterfaces()
}

func (interpreter *Interpreter) validateClass(classDecl *ast.ClassDeclarationStatement) phpError.Error {

	// Check if interface exists
	if classDecl.BaseClass != "" {
		_, found := interpreter.GetClass(classDecl.BaseClass)
		if !found {
			return phpError.NewError(`Class "%s" not found in %s`, classDecl.BaseClass, classDecl.GetPosString())
		}
	}

	// Check if all interfaces are implemented
	if !classDecl.IsAbstract && len(classDecl.Interfaces) > 0 {
		for _, interfaceName := range classDecl.Interfaces {
			// Check if interface exists
			interfaceDecl, found := interpreter.GetInterface(classDecl.GetPosition().File.GetNamespaceStr() + interfaceName)
			if !found {
				return phpError.NewError(`Interface "%s" not found in %s`, interfaceName, classDecl.GetPosString())
			}

			// Check if all methods are implemented
			missingMethods := []string{}
			for _, methodName := range interfaceDecl.MethodNames {
				// Check if method exists
				methodDef, found := interpreter.getClassMethod(classDecl, methodName)
				if !found {
					missingMethods = append(missingMethods, interfaceName+"::"+methodName)
					continue
				}

				// Check if method signatures match
				classMethodSignature := MethodDeclToSignature(methodDef)

				interfaceMethodDef, _ := interfaceDecl.GetMethod(methodName)
				interfaceMethodSignature := MethodDeclToSignature(interfaceMethodDef)

				if classMethodSignature != interfaceMethodSignature {
					return phpError.NewError(
						"Declaration of %s::%s must be compatible with %s::%s in %s",
						classDecl.GetQualifiedName(), classMethodSignature,
						interfaceDecl.GetQualifiedName(), interfaceMethodSignature,
						classDecl.GetPosString(),
					)
				}
			}
			if len(missingMethods) > 0 {
				if len(missingMethods) > 1 {
					return phpError.NewError(
						"Class %s contains %d abstract methods and must therefore be declared abstract or implement the remaining methods (%s) in %s",
						classDecl.GetQualifiedName(), len(missingMethods), common.ImplodeSlice(missingMethods, ", "), classDecl.GetPosString(),
					)
				}
				return phpError.NewError(
					"Class %s contains %d abstract method and must therefore be declared abstract or implement the remaining method (%s) in %s",
					classDecl.GetQualifiedName(), len(missingMethods), common.ImplodeSlice(missingMethods, ", "), classDecl.GetPosString(),
				)

			}
		}
	}

	return nil
}

func (interpreter *Interpreter) destructObject(object *values.Object, env *Environment) phpError.Error {
	if object.IsDestructed {
		return nil
	}
	_, found := interpreter.getClassMethod(object.Class, "__destruct")
	if found {
		_, err := interpreter.CallMethod(object, "__destruct", []ast.IExpression{}, env)
		if err != nil {
			return err
		}
		object.IsDestructed = true
	}
	return nil
}

func (interpreter *Interpreter) getClassMethod(class *ast.ClassDeclarationStatement, methodName string) (*ast.MethodDefinitionStatement, bool) {
	classDecl := class
	for classDecl != nil {
		MethodDecl, found := classDecl.GetMethod(methodName)
		if found {
			return MethodDecl, true
		}
		if classDecl.BaseClass == "" {
			return nil, false
		}
		classDecl, _ = interpreter.GetClass(classDecl.BaseClass)
	}
	return nil, false
}

func (interpreter *Interpreter) destructAllObjects(env *Environment) {
	objects := env.getAllObjects()
	for _, object := range objects {
		if err := interpreter.destructObject(object, env); err != nil {
			interpreter.PrintError(err)
		}
	}
}

func (interpreter *Interpreter) CallStaticMethod(class *ast.ClassDeclarationStatement, method string, args []ast.IExpression, env *Environment) (*values.Slot, phpError.Error) {
	return interpreter.callMethod(nil, class, method, args, env)
}

func (interpreter *Interpreter) CallMethod(object *values.Object, method string, args []ast.IExpression, env *Environment) (*values.Slot, phpError.Error) {
	return interpreter.callMethod(object, nil, method, args, env)
}

func (interpreter *Interpreter) callMethod(object *values.Object, class *ast.ClassDeclarationStatement, method string, args []ast.IExpression, env *Environment) (*values.Slot, phpError.Error) {
	if object != nil {
		class = object.Class
	}
	methodDefinition, found := interpreter.getClassMethod(class, method)
	if !found {
		return values.NewNullSlot(), phpError.NewError(`Class %s does not have a function "%s"`, class.Name, method)
	}
	if object == nil && !methodDefinition.IsStatic() && !slices.Contains([]string{"__callStatic" /*"__construct"*/}, strings.ToLower(method)) {
		return values.NewNullSlot(), phpError.NewError(`%s::%s is not a static function`, class.Name, method)
	}

	methodEnv, err := NewEnvironment(env, nil, interpreter)
	if err != nil {
		return values.NewVoidSlot(), err
	}
	methodEnv.CurrentMethod = methodDefinition
	methodEnv.AddConstant("self", values.NewNull())
	methodEnv.AddConstant("parent", values.NewNull())
	if object != nil {
		methodEnv.CurrentObject = object
		methodEnv.variables["$this"] = values.NewSlot(object)
	}

	if methodDefinition.GetRequiredParamLen() > len(args) {
		return values.NewVoidSlot(), phpError.NewError(
			"Uncaught ArgumentCountError: %s::%s() expects exactly %d arguments, %d given",
			class.BaseClass, methodDefinition.Name, len(methodDefinition.Params), len(args),
		)
	}

	for index, param := range methodDefinition.Params {
		var slot *values.Slot
		if index+1 > len(args) {
			slot = must(interpreter.processStmt(param.DefaultValue, env))
		} else {
			slot = must(interpreter.processStmt(args[index], env))
		}

		// Check if the parameter types match
		err = checkParameterTypes(slot.Value, param.Type)
		if err != nil && err.GetMessage() == "Types do not match" {
			givenType, err := variableHandling.GetType(slot.Value)
			if err != nil {
				return values.NewVoidSlot(), err
			}
			return values.NewVoidSlot(), phpError.NewError(
				"Uncaught TypeError: %s::%s(): Argument #%d (%s) must be of type %s, %s given",
				class.BaseClass, methodDefinition.Name, index+1, param.Name, strings.Join(param.Type, "|"), givenType,
			)
		}
		// Declare parameter in method environment
		methodEnv.declareVariable(param.Name, values.DeepCopy(slot).Value)
	}

	slot, err := interpreter.processStmt(methodDefinition.Body, methodEnv)
	if err != nil && !(err.GetErrorType() == phpError.EventError && err.GetMessage() == phpError.ReturnEvent) {
		return slot, err
	}
	err = checkParameterTypes(slot.Value, methodDefinition.ReturnType)
	if err != nil && err.GetMessage() == "Types do not match" {
		givenType, err := variableHandling.GetType(slot.Value)
		if slot.GetType() == values.VoidValue {
			givenType = "void"
		}
		if err != nil {
			return slot, err
		}
		return slot, phpError.NewError(
			"Uncaught TypeError: %s::%s(): Return value must be of type %s, %s given",
			class.BaseClass, methodDefinition.Name, strings.Join(methodDefinition.ReturnType, "|"), givenType,
		)
	}

	return slot, nil
}

// -------------------------------------- Caching -------------------------------------- MARK: Caching

func (interpreter *Interpreter) isCached(stmt ast.IStatement) bool {
	_, found := interpreter.cache[stmt.GetId()]
	return found
}

func (interpreter *Interpreter) writeCache(stmt ast.IStatement, value values.RuntimeValue) values.RuntimeValue {
	interpreter.cache[stmt.GetId()] = value
	return value
}

// -------------------------------------- RuntimeValue -------------------------------------- MARK: RuntimeValue

func (interpreter *Interpreter) exprToRuntimeValue(expr ast.IExpression, env *Environment) (*values.Slot, phpError.Error) {
	switch expr.GetKind() {
	case ast.ArrayLiteralExpr:
		array := values.NewArray()
		for _, key := range expr.(*ast.ArrayLiteralExpression).Keys {
			keyValueSlot := values.NewSlot(nil)
			var err phpError.Error
			if key.GetKind() != ast.ArrayNextKeyExpr {
				keyValueSlot, err = interpreter.processStmt(key, env)
				if err != nil {
					return values.NewVoidSlot(), err
				}
			}
			elementValueSlot, err := interpreter.processStmt(expr.(*ast.ArrayLiteralExpression).Elements[key], env)
			if err != nil {
				return values.NewVoidSlot(), err
			}
			if err = array.SetElement(keyValueSlot.Value, elementValueSlot.Value); err != nil {
				return values.NewVoidSlot(), err
			}
		}
		return values.NewSlot(array), nil
	case ast.IntegerLiteralExpr:
		return values.NewIntSlot(expr.(*ast.IntegerLiteralExpression).Value), nil
	case ast.FloatingLiteralExpr:
		return values.NewFloatSlot(expr.(*ast.FloatingLiteralExpression).Value), nil
	case ast.StringLiteralExpr:
		str := expr.(*ast.StringLiteralExpression).Value
		strType := expr.(*ast.StringLiteralExpression).StringType
		if strType == ast.DoubleQuotedString || strType == ast.HeredocString {
			// Supported expression: variable substitution: `echo "{$a}";`
			// variable substitution
			// TODO improve variable substitution: Regex and replace will not work for every case here. A parser is required that searches for variables, subscriptExpr, ... and resolves them.
			// TODO improve variable substitution to detect if a $ is escaped. E.g. "\$i"
			// TODO improve variable substitution to accept nested arrays "$a[$b['c']][0]"
			r, _ := regexp.Compile(`({\$[A-Za-z_][A-Za-z0-9_]*['A-Za-z0-9\[\]]*[^}]*})|(\$[A-Za-z_][A-Za-z0-9_]*['A-Za-z0-9\[\]]*)`)
			matches := r.FindAllString(str, -1)
			for _, match := range matches {
				varExpr := match
				if match[0] == '{' {
					// Remove curly braces
					varExpr = match[1 : len(match)-1]
				}
				// TODO Find better solution for code evaluation
				exprStr := "<?= " + varExpr + ";"
				interp, err := NewInterpreter(interpreter.GetExectionContext(), interpreter.ini, interpreter.request, "__file_name__")
				if err != nil {
					return values.NewVoidSlot(), err
				}
				result, err := interp.process(exprStr, env, true)
				if err != nil {
					return values.NewVoidSlot(), err
				}
				// TODO Improve bad fix so that the warning is above the possible string output
				if strings.Contains(result, "Warning: Undefined variable") {
					filenameRegex := regexp.MustCompile(`__file_name__:\d+:\d+`)
					interpreter.PrintError(
						phpError.NewWarning(
							"%s", strings.TrimPrefix(
								strings.TrimSpace(
									filenameRegex.ReplaceAllString(result, expr.GetPosString())),
								"Warning: ")))
					result = ""
				}
				str = strings.Replace(str, match, result, 1)
			}

			// unicode escape sequence
			r, _ = regexp.Compile(`\\u\{[0-9a-fA-F]+\}`)
			matches = r.FindAllString(str, -1)
			for _, match := range matches {
				unicodeChar, err := strconv.ParseInt(match[3:len(match)-1], 16, 32)
				if err != nil {
					return values.NewVoidSlot(), phpError.NewError("%s", err.Error())
				}
				str = strings.Replace(str, match, string(rune(unicodeChar)), 1)
			}
		}
		return values.NewStrSlot(str), nil
	default:
		return values.NewVoidSlot(), phpError.NewError("exprToRuntimeValue: Unsupported expression: %s", expr)
	}
}

// -------------------------------------- inc-dec-calculation -------------------------------------- MARK: inc-dec-calculation

func calculateIncDec(operator string, operand values.RuntimeValue) (*values.Slot, phpError.Error) {
	switch operand.GetType() {
	case values.BoolValue:
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with a Boolean-valued operand, there is no side effect, and the result is the operand’s value.
		return values.NewSlot(operand), nil
	case values.FloatValue:
		return calculateIncDecFloating(operator, operand.(*values.Float))
	case values.IntValue:
		return calculateIncDecInteger(operator, operand.(*values.Int))
	case values.NullValue:
		return calculateIncDecNull(operator)
	case values.StrValue:
		return calculateIncDecString(operator, operand.(*values.Str))
	default:
		return values.NewVoidSlot(), phpError.NewError(`calculateIncDec: Type "%s" not implemented`, operand.GetType())
	}

	// TODO calculateIncDec - object
	// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
	// If the operand has an object type supporting the operation, then the object semantics defines the result. Otherwise, the operation has no effect and the result is the operand.
}

func calculateIncDecInteger(operator string, operand *values.Int) (*values.Slot, phpError.Error) {
	switch operator {
	case "++":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		//For a prefix "++" operator used with an arithmetic operand, the side effect of the operator is to increment the value of the operand by 1.
		// The result is the value of the operand after it has been incremented.
		// If an int operand’s value is the largest representable for that type, the operand is incremented as if it were float.

		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with an operand having the value INF, -INF, or NAN, there is no side effect, and the result is the operand’s value.
		return calculateInteger(operand, "+", values.NewInt(1))

	case "--":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix "--" operator used with an arithmetic operand, the side effect of the operator is to decrement the value of the operand by 1.
		// The result is the value of the operand after it has been decremented.
		// If an int operand’s value is the smallest representable for that type, the operand is decremented as if it were float.

		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with an operand having the value INF, -INF, or NAN, there is no side effect, and the result is the operand’s value.
		return calculateInteger(operand, "-", values.NewInt(1))

	default:
		return values.NewIntSlot(0), phpError.NewError(`calculateIncDecInteger: Operator "%s" not implemented`, operator)
	}
}

func calculateIncDecFloating(operator string, operand *values.Float) (*values.Slot, phpError.Error) {
	switch operator {
	case "++":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		//For a prefix "++" operator used with an arithmetic operand, the side effect of the operator is to increment the value of the operand by 1.
		// The result is the value of the operand after it has been incremented.
		// If an int operand’s value is the largest representable for that type, the operand is incremented as if it were float.

		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with an operand having the value INF, -INF, or NAN, there is no side effect, and the result is the operand’s value.
		return calculateFloating(operand, "+", values.NewFloat(1))

	case "--":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix "--" operator used with an arithmetic operand, the side effect of the operator is to decrement the value of the operand by 1.
		// The result is the value of the operand after it has been decremented.
		// If an int operand’s value is the smallest representable for that type, the operand is decremented as if it were float.

		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with an operand having the value INF, -INF, or NAN, there is no side effect, and the result is the operand’s value.
		return calculateFloating(operand, "-", values.NewFloat(1))

	default:
		return values.NewIntSlot(0), phpError.NewError(`calculateIncDecFloating: Operator "%s" not implemented`, operator)
	}
}

func calculateIncDecNull(operator string) (*values.Slot, phpError.Error) {
	switch operator {
	case "++":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ operator used with a NULL-valued operand, the side effect is that the operand’s type is changed to int,
		// the operand’s value is set to zero, and that value is incremented by 1.
		// The result is the value of the operand after it has been incremented.
		return values.NewIntSlot(1), nil

	case "--":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix – operator used with a NULL-valued operand, there is no side effect, and the result is the operand’s value.
		return values.NewNullSlot(), nil

	default:
		return values.NewIntSlot(0), phpError.NewError(`calculateIncDecNull: Operator "%s" not implemented`, operator)
	}
}

func calculateIncDecString(operator string, operand *values.Str) (*values.Slot, phpError.Error) {
	switch operator {
	case "++":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix "++" operator used with an operand whose value is an empty string,
		// the side effect is that the operand’s value is changed to the string “1”. The type of the operand is unchanged.
		// The result is the new value of the operand.
		if operand.Value == "" {
			return values.NewStrSlot("1"), nil
		}
		return values.NewVoidSlot(), phpError.NewError("TODO calculateIncDecString")

	case "--":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix "--" operator used with an operand whose value is an empty string,
		// the side effect is that the operand’s type is changed to int, the operand’s value is set to zero,
		// and that value is decremented by 1. The result is the value of the operand after it has been incremented.
		if operand.Value == "" {
			return values.NewIntSlot(-1), nil
		}
		return values.NewVoidSlot(), phpError.NewError("TODO calculateIncDecString")

	default:
		return values.NewIntSlot(0), phpError.NewError(`calculateIncDecNull: Operator "%s" not implemented`, operator)
	}

	// TODO calculateIncDecString
	// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
	/*
		String Operands

		For a prefix -- or ++ operator used with a numeric string, the numeric string is treated as the corresponding int or float value.

		For a prefix -- operator used with a non-numeric string-valued operand, there is no side effect, and the result is the operand’s value.

		For a non-numeric string-valued operand that contains only alphanumeric characters, for a prefix ++ operator, the operand is considered to be a representation of a base-36 number (i.e., with digits 0–9 followed by A–Z or a–z) in which letter case is ignored for value purposes. The right-most digit is incremented by 1. For the digits 0–8, that means going to 1–9. For the letters “A”–“Y” (or “a”–“y”), that means going to “B”–“Z” (or “b”–“z”). For the digit 9, the digit becomes 0, and the carry is added to the next left-most digit, and so on. For the digit “Z” (or “z”), the resulting string has an extra digit “A” (or “a”) appended. For example, when incrementing, “a” -> “b”, “Z” -> “AA”, “AA” -> “AB”, “F29” -> “F30”, “FZ9” -> “GA0”, and “ZZ9” -> “AAA0”. A digit position containing a number wraps modulo-10, while a digit position containing a letter wraps modulo-26.

		For a non-numeric string-valued operand that contains any non-alphanumeric characters, for a prefix ++ operator, all characters up to and including the right-most non-alphanumeric character is passed through to the resulting string, unchanged. Characters to the right of that right-most non-alphanumeric character are treated like a non-numeric string-valued operand that contains only alphanumeric characters, except that the resulting string will not be extended. Instead, a digit position containing a number wraps modulo-10, while a digit position containing a letter wraps modulo-26.
	*/
}

// -------------------------------------- unary-op-calculation -------------------------------------- MARK: unary-op-calculation

func calculateUnary(operator string, operand values.RuntimeValue) (*values.Slot, phpError.Error) {
	switch operand.GetType() {
	case values.BoolValue:
		return calculateUnaryBoolean(operator, operand.(*values.Bool))
	case values.IntValue:
		return calculateUnaryInteger(operator, operand.(*values.Int))
	case values.FloatValue:
		return calculateUnaryFloating(operator, operand.(*values.Float))
	case values.NullValue:
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary + or unary - operator used with a NULL-valued operand, the value of the result is zero and the type is int.
		return values.NewIntSlot(0), nil
	default:
		return values.NewVoidSlot(), phpError.NewError(`calculateUnary: Type "%s" not implemented`, operand.GetType())
	}

	// TODO calculateUnary - string
	// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
	// For a unary + or - operator used with a numeric string or a leading-numeric string, the string is first converted to an int or float, as appropriate, after which it is handled as an arithmetic operand. The trailing non-numeric characters in leading-numeric strings are ignored. With a non-numeric string, the result has type int and value 0. If the string was leading-numeric or non-numeric, a non-fatal error MUST be produced.
	// For a unary ~ operator used with a string, the result is the string with each byte being bitwise complement of the corresponding byte of the source string.

	// TODO calculateUnary - object
	// If the operand has an object type supporting the operation, then the object semantics defines the result. Otherwise, for ~ the fatal error is issued and for + and - the object is converted to int.
}

func calculateUnaryBoolean(operator string, operand *values.Bool) (*values.Slot, phpError.Error) {
	switch operator {
	case "+":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary "+" operator used with a TRUE-valued operand, the value of the result is 1 and the type is int.
		// When used with a FALSE-valued operand, the value of the result is zero and the type is int.
		if operand.Value {
			return values.NewIntSlot(1), nil
		}
		return values.NewIntSlot(0), nil

	case "-":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary "-" operator used with a TRUE-valued operand, the value of the result is -1 and the type is int.
		// When used with a FALSE-valued operand, the value of the result is zero and the type is int.
		if operand.Value {
			return values.NewIntSlot(-1), nil
		}
		return values.NewIntSlot(0), nil

	default:
		return values.NewIntSlot(0), phpError.NewError(`calculateUnaryBoolean: Operator "%s" not implemented`, operator)
	}
}

func calculateUnaryFloating(operator string, operand *values.Float) (*values.Slot, phpError.Error) {
	switch operator {
	case "+":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary "+" operator used with an arithmetic operand, the type and value of the result is the type and value of the operand.
		return values.NewSlot(operand), nil

	case "-":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary - operator used with an arithmetic operand, the value of the result is the negated value of the operand.
		// However, if an int operand’s original value is the smallest representable for that type,
		// the operand is treated as if it were float and the result will be float.
		return values.NewFloatSlot(-operand.Value), nil

	case "~":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary ~ operator used with a float operand, the value of the operand is first converted to int before the bitwise complement is computed.
		intRuntimeValue, err := variableHandling.ToValueType(values.IntValue, operand, false)
		if err != nil {
			return values.NewFloatSlot(0), err
		}
		return calculateUnaryInteger(operator, intRuntimeValue.(*values.Int))

	default:
		return values.NewFloatSlot(0), phpError.NewError(`calculateUnaryFloating: Operator "%s" not implemented`, operator)
	}
}

func calculateUnaryInteger(operator string, operand *values.Int) (*values.Slot, phpError.Error) {
	switch operator {
	case "+":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary "+" operator used with an arithmetic operand, the type and value of the result is the type and value of the operand.
		return values.NewSlot(operand), nil

	case "-":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary - operator used with an arithmetic operand, the value of the result is the negated value of the operand.
		// However, if an int operand’s original value is the smallest representable for that type,
		// the operand is treated as if it were float and the result will be float.
		return values.NewIntSlot(-operand.Value), nil

	case "~":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary ~ operator used with an int operand, the type of the result is int.
		// The value of the result is the bitwise complement of the value of the operand
		// (that is, each bit in the result is set if and only if the corresponding bit in the operand is clear).
		return values.NewIntSlot(^operand.Value), nil
	default:
		return values.NewIntSlot(0), phpError.NewError(`calculateUnaryInteger: Operator "%s" not implemented`, operator)
	}
}

// -------------------------------------- binary-op-calculation -------------------------------------- MARK: binary-op-calculation

func calculate(operand1 values.RuntimeValue, operator string, operand2 values.RuntimeValue) (*values.Slot, phpError.Error) {
	resultType := values.VoidValue
	if slices.Contains([]string{"."}, operator) {
		resultType = values.StrValue
	} else if slices.Contains([]string{"&", "|", "^", "<<", ">>"}, operator) {
		resultType = values.IntValue
	} else {
		resultType = values.IntValue
		if operand1.GetType() == values.FloatValue || operand2.GetType() == values.FloatValue {
			resultType = values.FloatValue
		}
	}

	var err phpError.Error
	operand1, err = variableHandling.ToValueType(resultType, operand1, false)
	if err != nil {
		return values.NewVoidSlot(), err
	}
	operand2, err = variableHandling.ToValueType(resultType, operand2, false)
	if err != nil {
		return values.NewVoidSlot(), err
	}
	// TODO testing how PHP behavious: var_dump(1.0 + 2); var_dump(1 + 2.0); var_dump("1" + 2);
	// var_dump("1" + "2"); => int
	// var_dump("1" . 2); => str
	// type order "string" - "int" - "float"

	// Testen
	//   true + 2
	//   true && 3

	switch resultType {
	case values.IntValue:
		return calculateInteger(operand1.(*values.Int), operator, operand2.(*values.Int))
	case values.FloatValue:
		return calculateFloating(operand1.(*values.Float), operator, operand2.(*values.Float))
	case values.StrValue:
		return calculateString(operand1.(*values.Str), operator, operand2.(*values.Str))
	default:
		return values.NewVoidSlot(), phpError.NewError(`calculate: Type "%s" not implemented`, resultType)
	}
}

func calculateFloating(operand1 *values.Float, operator string, operand2 *values.Float) (*values.Slot, phpError.Error) {
	switch operator {
	case "+":
		return values.NewFloatSlot(operand1.Value + operand2.Value), nil
	case "-":
		return values.NewFloatSlot(operand1.Value - operand2.Value), nil
	case "*":
		return values.NewFloatSlot(operand1.Value * operand2.Value), nil
	case "/":
		return values.NewFloatSlot(operand1.Value / operand2.Value), nil
	case "**":
		return values.NewFloatSlot(math.Pow(operand1.Value, operand2.Value)), nil
	case "%":
		op1, err := variableHandling.IntVal(operand1, false)
		if err != nil {
			return values.NewFloatSlot(0), err
		}
		op2, err := variableHandling.IntVal(operand2, false)
		if err != nil {
			return values.NewFloatSlot(0), err
		}
		return values.NewIntSlot(op1 % op2), nil
	default:
		return values.NewFloatSlot(0), phpError.NewError(`calculateFloating: Operator "%s" not implemented`, operator)
	}
}

func calculateInteger(operand1 *values.Int, operator string, operand2 *values.Int) (*values.Slot, phpError.Error) {
	switch operator {
	case "<<":
		if operand2.Value < 0 {
			return values.NewVoidSlot(), phpError.NewError("Bit shift by negative number")
		}
		return values.NewIntSlot(operand1.Value << operand2.Value), nil
	case ">>":
		if operand2.Value < 0 {
			return values.NewVoidSlot(), phpError.NewError("Bit shift by negative number")
		}
		return values.NewIntSlot(operand1.Value >> operand2.Value), nil
	case "^":
		return values.NewIntSlot(operand1.Value ^ operand2.Value), nil
	case "|":
		return values.NewIntSlot(operand1.Value | operand2.Value), nil
	case "&":
		return values.NewIntSlot(operand1.Value & operand2.Value), nil
	case "+":
		return values.NewIntSlot(operand1.Value + operand2.Value), nil
	case "-":
		return values.NewIntSlot(operand1.Value - operand2.Value), nil
	case "*":
		return values.NewIntSlot(operand1.Value * operand2.Value), nil
	case "/":
		if operand2.Value == 0 {
			// TODO Add position in output: Fatal error: Uncaught DivisionByZeroError: Division by zero in /home/user/scripts/code.php:3
			return values.NewIntSlot(0), phpError.NewError("Uncaught DivisionByZeroError: Division by zero")
		}
		return values.NewIntSlot(operand1.Value / operand2.Value), nil
	case "%":
		if operand2.Value == 0 {
			return values.NewVoidSlot(), phpError.NewError("Division by zero")
		}
		return values.NewIntSlot(operand1.Value % operand2.Value), nil
	case "**":
		return values.NewIntSlot(int64(math.Pow(float64(operand1.Value), float64(operand2.Value)))), nil
	default:
		return values.NewVoidSlot(), phpError.NewError(`calculateInteger: Operator "%s" not implemented`, operator)
	}
}

func calculateString(operand1 *values.Str, operator string, operand2 *values.Str) (*values.Slot, phpError.Error) {
	switch operator {
	case ".":
		return values.NewStrSlot(operand1.Value + operand2.Value), nil
	default:
		return values.NewStrSlot(""), phpError.NewError(`calculateString: Operator "%s" not implemented`, operator)
	}
}
