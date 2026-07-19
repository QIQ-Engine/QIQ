package ast

import (
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/config"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

func ToString(stmt IStatement) string { return NewDumpVisitor(false).toString(stmt) }

func PrintInterpreterCallstack(stmt IStatement) {
	if !config.ShowInterpreterCallStack {
		return
	}
	if stmt == nil {
		println("nil")
		return
	}
	println(NewDumpVisitor(true).toString(stmt))
}

var _ Visitor = &DumpVisitor{}

type DumpVisitor struct {
	withPos bool
}

func NewDumpVisitor(withPos bool) DumpVisitor { return DumpVisitor{withPos: withPos} }

func (visitor DumpVisitor) toString(stmt IStatement) string {
	if stmt == nil {
		return `"nil"`
	}
	// Check if the underlying value is nil
	val := reflect.ValueOf(stmt)
	if val.Kind() == reflect.Pointer && val.IsNil() {
		return `"nil"`
	}

	result, _ := stmt.Process(visitor, nil)
	return result.(string)
}

func (visitor DumpVisitor) dumpStatements(statements []IStatement) string {
	stmts := "["
	for _, statement := range statements {
		if len(stmts) > 1 {
			stmts += ", "
		}
		stmts += visitor.toString(statement)
	}
	stmts += "]"
	return stmts
}

func (visitor DumpVisitor) dumpExpressions(expressions []IExpression) string {
	exprs := "["
	for _, expression := range expressions {
		if len(exprs) > 1 {
			exprs += ", "
		}
		exprs += visitor.toString(expression)
	}
	exprs += "]"
	return exprs
}

func (visitor DumpVisitor) getKindAndPos(stmt IStatement) string {
	kind := fmt.Sprintf(`"kind": "%s"`, stmt.GetKind())
	if !visitor.withPos {
		return kind
	}
	return fmt.Sprintf(`%s, "pos": "%s"`, kind, stmt.GetPosString())
}

// ProcessAnonymousFunctionCreationExpr implements Visitor.
func (visitor DumpVisitor) ProcessAnonymousFunctionCreationExpr(stmt *AnonymousFunctionCreationExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "params": %s, "body": %s, "returnType": [%s] }`,
		visitor.getKindAndPos(stmt), visitor.ProcessFunctionParameterSlice(stmt.Params), visitor.toString(stmt.Body), common.ImplodeStrSlice(stmt.ReturnType),
	), nil
}

func (visitor DumpVisitor) ProcessFunctionParameterSlice(parameters []FunctionParameter) string {
	params := "["
	for _, param := range parameters {
		if len(params) > 1 {
			params += ", "
		}
		params += fmt.Sprintf(`{ "byRef": %v, "name": "%s", "type": [%s], "defaultValue": %s }`, param.ByRef, param.Name, common.ImplodeStrSlice(param.Type), visitor.toString(param.DefaultValue))
	}
	params += "]"
	return params
}

// ProcessArrayLiteralExpr implements Visitor.
func (visitor DumpVisitor) ProcessArrayLiteralExpr(stmt *ArrayLiteralExpression, _ any) (any, error) {
	elements := "["
	for _, key := range stmt.Keys {
		if len(elements) > 1 {
			elements += ", "
		}
		elements += fmt.Sprintf(`{ "key": %s, "value": %s }`, visitor.toString(key), visitor.toString(stmt.Elements[key]))
	}
	elements += "]"
	return fmt.Sprintf(`{ %s, "elements": %s }`, visitor.getKindAndPos(stmt), elements), nil
}

// ProcessArrayNextKeyExpr implements Visitor.
func (visitor DumpVisitor) ProcessArrayNextKeyExpr(stmt *ArrayNextKeyExpression, _ any) (any, error) {
	return fmt.Sprintf("{ %s }", visitor.getKindAndPos(stmt)), nil
}

// ProcessBinaryOpExpr implements Visitor.
func (visitor DumpVisitor) ProcessBinaryOpExpr(stmt *BinaryOpExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "lhs": %s, "operator": "%s", "rhs": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Lhs), stmt.Operator, visitor.toString(stmt.Rhs),
	), nil
}

// ProcessBreakStmt implements Visitor.
func (visitor DumpVisitor) ProcessBreakStmt(stmt *BreakStatement, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessCastExpr implements Visitor.
func (visitor DumpVisitor) ProcessCastExpr(stmt *CastExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "operator": "%s", "expr": %s }`,
		visitor.getKindAndPos(stmt), stmt.Operator, visitor.toString(stmt.Expr),
	), nil
}

// ProcessInterfaceDeclarationStmt implements Visitor.
func (visitor DumpVisitor) ProcessInterfaceDeclarationStmt(stmt *InterfaceDeclarationStatement, _ any) (any, error) {
	constants := "["
	constantsKeys := slices.Sorted(maps.Keys(stmt.Constants))
	for _, key := range constantsKeys {
		if len(constants) > 1 {
			constants += ", "
		}
		constant := stmt.Constants[key]
		constants += fmt.Sprintf(
			`{ "visibility": "%s", "name": "%s", %s }`,
			constant.Visiblity, constant.Name, visitor.toString(constant.Value),
		)
	}
	constants += "]"

	methods := "["
	methodsKeys := slices.Sorted(maps.Keys(stmt.Methods))
	for _, key := range methodsKeys {
		if len(methods) > 1 {
			methods += ", "
		}

		method := stmt.Methods[key]

		methods += fmt.Sprintf(`{ "name": "%s", "modifiers": [%s], "returnType": [%s], "parameters": %s }`,
			method.Name, common.ImplodeStrSlice(method.Modifiers), common.ImplodeStrSlice(method.ReturnType), visitor.ProcessFunctionParameterSlice(method.Params),
		)
	}
	methods += "]"

	return fmt.Sprintf(
		`{ %s, "name": "%s", "extends": "%s", "constants": %s, "methods": %s }`,
		visitor.getKindAndPos(stmt), stmt.Name, common.ImplodeStrSlice(stmt.Parents), constants, methods,
	), nil
}

// ProcessClassDeclarationStmt implements Visitor.
func (visitor DumpVisitor) ProcessClassDeclarationStmt(stmt *ClassDeclarationStatement, _ any) (any, error) {
	constants := "["
	constantsKeys := slices.Sorted(maps.Keys(stmt.Constants))
	for _, key := range constantsKeys {
		if len(constants) > 1 {
			constants += ", "
		}
		constant := stmt.Constants[key]
		constants += fmt.Sprintf(
			`{ "visibility": "%s", "name": "%s", %s }`,
			constant.Visiblity, constant.Name, visitor.toString(constant.Value),
		)
	}
	constants += "]"

	methods := "["
	methodsKeys := slices.Sorted(maps.Keys(stmt.Methods))
	for _, key := range methodsKeys {
		if len(methods) > 1 {
			methods += ", "
		}

		method := stmt.Methods[key]

		methods += fmt.Sprintf(`{ "name": "%s", "modifiers": [%s], "returnType": [%s], "parameters": %s, "body": %s }`,
			method.Name, common.ImplodeStrSlice(method.Modifiers), common.ImplodeStrSlice(method.ReturnType), visitor.ProcessFunctionParameterSlice(method.Params), visitor.toString(method.Body),
		)
	}
	methods += "]"

	traits := "["
	for _, trait := range stmt.Traits {
		if len(traits) > 1 {
			traits += ", "
		}
		traits += `"` + trait.Name + `"`
	}
	traits += "]"

	properties := "["
	propertiesKeys := slices.Sorted(maps.Keys(stmt.Properties))
	for _, key := range propertiesKeys {
		if len(properties) > 1 {
			properties += ", "
		}
		property := stmt.Properties[key]
		properties += fmt.Sprintf(`{ "name": "%s", "isStatic": %v, "visibility": "%s", "type": [%s], "initialValue": %s }`,
			property.Name, property.IsStatic, property.Visibility, common.ImplodeStrSlice(property.Type), visitor.toString(property.InitialValue),
		)
	}
	properties += "]"

	return fmt.Sprintf(
		`{ %s, "name": "%s", "isAbstract": %v, "isFinal": %v, "extends": "%s", "implements": %s, "constants": %s, "methods": %s, "traits": %s, "properties": %s }`,
		visitor.getKindAndPos(stmt), stmt.Name, stmt.IsAbstract, stmt.IsFinal, stmt.BaseClass, common.ImplodeStrSlice(stmt.Interfaces), constants, methods, traits, properties,
	), nil
}

// ProcessCoalesceExpr implements Visitor.
func (visitor DumpVisitor) ProcessCoalesceExpr(stmt *CoalesceExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "condition": %s, "elseExpr": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Cond), visitor.toString(stmt.ElseExpr),
	), nil
}

// ProcessCompoundAssignmentExpr implements Visitor.
func (visitor DumpVisitor) ProcessCompoundAssignmentExpr(stmt *CompoundAssignmentExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "variable": %s, "operator": "%s", "value": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Variable), stmt.Operator, visitor.toString(stmt.Value),
	), nil
}

// ProcessCompoundStmt implements Visitor.
func (visitor DumpVisitor) ProcessCompoundStmt(stmt *CompoundStatement, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "stmts": %s }`, visitor.getKindAndPos(stmt), visitor.dumpStatements(stmt.Statements)), nil
}

// ProcessConditionalExpr implements Visitor.
func (visitor DumpVisitor) ProcessConditionalExpr(stmt *ConditionalExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "condition": %s, "ifExpr": %s, "elseExpr": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Cond), visitor.toString(stmt.IfExpr), visitor.toString(stmt.ElseExpr),
	), nil
}

// ProcessConstDeclarationStmt implements Visitor.
func (visitor DumpVisitor) ProcessConstDeclarationStmt(stmt *ConstDeclarationStatement, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, name: "%s", "value": %s }`, visitor.getKindAndPos(stmt), stmt.Name, visitor.toString(stmt.Value)), nil
}

// ProcessConstantAccessExpr implements Visitor.
func (visitor DumpVisitor) ProcessConstantAccessExpr(stmt *ConstantAccessExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "constantName": "%s" }`, visitor.getKindAndPos(stmt), stmt.ConstantName), nil
}

// ProcessContinueStmt implements Visitor.
func (visitor DumpVisitor) ProcessContinueStmt(stmt *ContinueStatement, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessDeclareStmt implements Visitor.
func (visitor DumpVisitor) ProcessDeclareStmt(stmt *DeclareStatement, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "directive": "%s", "literal": %s }`,
		visitor.getKindAndPos(stmt), stmt.Directive, visitor.toString(stmt.Literal),
	), nil
}

// ProcessDoStmt implements Visitor.
func (visitor DumpVisitor) ProcessDoStmt(stmt *DoStatement, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "condition": %s, "block": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Condition), visitor.toString(stmt.Block),
	), nil
}

// ProcessEchoStmt implements Visitor.
func (visitor DumpVisitor) ProcessEchoStmt(stmt *EchoStatement, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.dumpExpressions(stmt.Expressions)), nil
}

// ProcessEmptyIntrinsicExpr implements Visitor.
func (visitor DumpVisitor) ProcessEmptyIntrinsicExpr(stmt *EmptyIntrinsicExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "functionName": "%s", "arguments": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.FunctionName), visitor.dumpExpressions(stmt.Arguments),
	), nil
}

// ProcessEqualityExpr implements Visitor.
func (visitor DumpVisitor) ProcessEqualityExpr(stmt *EqualityExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "lhs": %s, "operator": "%s", "rhs": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Lhs), stmt.Operator, visitor.toString(stmt.Rhs),
	), nil
}

// ProcessErrorControlExpr implements Visitor.
func (visitor DumpVisitor) ProcessErrorControlExpr(stmt *ErrorControlExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessEvalIntrinsicExpr implements Visitor.
func (visitor DumpVisitor) ProcessEvalIntrinsicExpr(stmt *EvalIntrinsicExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "functionName": "%s", "arguments": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.FunctionName), visitor.dumpExpressions(stmt.Arguments),
	), nil
}

// ProcessExitIntrinsicExpr implements Visitor.
func (visitor DumpVisitor) ProcessExitIntrinsicExpr(stmt *ExitIntrinsicExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, functionName: "%s", arguments: %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.FunctionName), visitor.dumpExpressions(stmt.Arguments),
	), nil
}

// ProcessExpr implements Visitor.
func (visitor DumpVisitor) ProcessExpr(stmt *Expression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "type": "Expression" }`, visitor.getKindAndPos(stmt)), nil
}

// ProcessExpressionStmt implements Visitor.
func (visitor DumpVisitor) ProcessExpressionStmt(stmt *ExpressionStatement, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessFloatingLiteralExpr implements Visitor.
func (visitor DumpVisitor) ProcessFloatingLiteralExpr(stmt *FloatingLiteralExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "value": %f }`, visitor.getKindAndPos(stmt), stmt.Value), nil
}

// ProcessForeachStmt implements Visitor.
func (visitor DumpVisitor) ProcessForeachStmt(stmt *ForeachStatement, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "collection": %s, "key": %s, "value": %s, "block": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Collection), visitor.toString(stmt.Key), visitor.toString(stmt.Value), visitor.toString(stmt.Block),
	), nil
}

// ProcessForStmt implements Visitor.
func (visitor DumpVisitor) ProcessForStmt(stmt *ForStatement, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "initializer": %s, "control": %s, "endOfLoop": %s, "block": %s}`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Initializer), visitor.toString(stmt.Control), visitor.toString(stmt.EndOfLoop), visitor.toString(stmt.Block),
	), nil
}

// ProcessFunctionCallExpr implements Visitor.
func (visitor DumpVisitor) ProcessFunctionCallExpr(stmt *FunctionCallExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "functionName": "%s", "arguments": %s}`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.FunctionName), visitor.dumpExpressions(stmt.Arguments),
	), nil
}

// ProcessFunctionDefinitionStmt implements Visitor.
func (visitor DumpVisitor) ProcessFunctionDefinitionStmt(stmt *FunctionDefinitionStatement, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "name": "%s", "params": %v, "body": %s, "returnType": [%s]}`,
		visitor.getKindAndPos(stmt), stmt.FunctionName, visitor.ProcessFunctionParameterSlice(stmt.Params), visitor.toString(stmt.Body), common.ImplodeStrSlice(stmt.ReturnType),
	), nil
}

// ProcessGlobalDeclarationStmt implements Visitor.
func (visitor DumpVisitor) ProcessGlobalDeclarationStmt(stmt *GlobalDeclarationStatement, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "variables": %s}`, visitor.getKindAndPos(stmt), visitor.dumpExpressions(stmt.Variables)), nil
}

// ProcessIfStmt implements Visitor.
func (visitor DumpVisitor) ProcessIfStmt(stmt *IfStatement, _ any) (any, error) {
	var elseIf strings.Builder
	elseIf.WriteString("{")
	for _, elseIfStmt := range stmt.ElseIf {
		elseIf.WriteString(visitor.toString(elseIfStmt) + ", ")
	}
	elseIf.WriteString("}")
	return fmt.Sprintf(
		`{ %s, "condition": %s, "ifBlock": %s, "elseIf": %s, "else": %s}`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Condition), visitor.toString(stmt.IfBlock), elseIf.String(), visitor.toString(stmt.ElseBlock),
	), nil
}

// ProcessIncludeExpr implements Visitor.
func (visitor DumpVisitor) ProcessIncludeExpr(stmt *IncludeExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessIncludeOnceExpr implements Visitor.
func (visitor DumpVisitor) ProcessIncludeOnceExpr(stmt *IncludeOnceExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessIntegerLiteralExpr implements Visitor.
func (visitor DumpVisitor) ProcessIntegerLiteralExpr(stmt *IntegerLiteralExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "value": %d }`, visitor.getKindAndPos(stmt), stmt.Value), nil
}

// ProcessIssetIntrinsicExpr implements Visitor.
func (visitor DumpVisitor) ProcessIssetIntrinsicExpr(stmt *IssetIntrinsicExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "functionName": %s, "arguments": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.FunctionName), visitor.dumpExpressions(stmt.Arguments),
	), nil
}

// ProcessLogicalExpr implements Visitor.
func (visitor DumpVisitor) ProcessLogicalExpr(stmt *LogicalExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "lhs": %s, "operator": "%s", "rhs": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Lhs), stmt.Operator, visitor.toString(stmt.Rhs),
	), nil
}

// ProcessLogicalNotExpr implements Visitor.
func (visitor DumpVisitor) ProcessLogicalNotExpr(stmt *LogicalNotExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "operator": "%s", "expr": %s }`, visitor.getKindAndPos(stmt), stmt.Operator, visitor.toString(stmt.Expr)), nil
}

// ProcessMemberAccessExpr implements Visitor.
func (visitor DumpVisitor) ProcessMemberAccessExpr(stmt *MemberAccessExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "object": %s, "member": %s, "isScoped": %t }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Object), visitor.toString(stmt.Member), stmt.IsScoped), nil
}

// ProcessObjectCreationExpr implements Visitor.
func (visitor DumpVisitor) ProcessObjectCreationExpr(stmt *ObjectCreationExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "designator": "%s", "args": %s }`, visitor.getKindAndPos(stmt), stmt.Designator, visitor.dumpExpressions(stmt.Args)), nil
}

// ProcessParenthesizedExpr implements Visitor.
func (visitor DumpVisitor) ProcessParenthesizedExpr(stmt *ParenthesizedExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s}`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessPostfixIncExpr implements Visitor.
func (visitor DumpVisitor) ProcessPostfixIncExpr(stmt *PostfixIncExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "operator": "%s", "expr": %s }`, visitor.getKindAndPos(stmt), stmt.Operator, visitor.toString(stmt.Expr)), nil
}

// ProcessPrefixIncExpr implements Visitor.
func (visitor DumpVisitor) ProcessPrefixIncExpr(stmt *PrefixIncExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "operator": "%s", "expr": %s }`, visitor.getKindAndPos(stmt), stmt.Operator, visitor.toString(stmt.Expr)), nil
}

// ProcessPrintExpr implements Visitor.
func (visitor DumpVisitor) ProcessPrintExpr(stmt *PrintExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessRelationalExpr implements Visitor.
func (visitor DumpVisitor) ProcessRelationalExpr(stmt *RelationalExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "lhs": %s, "operator": "%s", "rhs": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Lhs), stmt.Operator, visitor.toString(stmt.Rhs),
	), nil
}

// ProcessRequireExpr implements Visitor.
func (visitor DumpVisitor) ProcessRequireExpr(stmt *RequireExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessRequireOnceExpr implements Visitor.
func (visitor DumpVisitor) ProcessRequireOnceExpr(stmt *RequireOnceExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessReturnStmt implements Visitor.
func (visitor DumpVisitor) ProcessReturnStmt(stmt *ReturnStatement, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s}`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessSimpleAssignmentExpr implements Visitor.
func (visitor DumpVisitor) ProcessSimpleAssignmentExpr(stmt *SimpleAssignmentExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "variable": %s, "value": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Variable), visitor.toString(stmt.Value),
	), nil
}

// ProcessSimpleVariableExpr implements Visitor.
func (visitor DumpVisitor) ProcessSimpleVariableExpr(stmt *SimpleVariableExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "variableName": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.VariableName)), nil
}

// ProcessStmt implements Visitor.
func (visitor DumpVisitor) ProcessStmt(stmt *Statement, _ any) (any, error) {
	return fmt.Sprintf(`{%s, "type": "Statement"}`, visitor.getKindAndPos(stmt)), nil
}

// ProcessStringLiteralExpr implements Visitor.
func (visitor DumpVisitor) ProcessStringLiteralExpr(stmt *StringLiteralExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "type": "%s", "value": "%s" }`, visitor.getKindAndPos(stmt), stmt.StringType, stmt.Value), nil
}

// ProcessSubscriptExpr implements Visitor.
func (visitor DumpVisitor) ProcessSubscriptExpr(stmt *SubscriptExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "variable": %s, "index": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Variable), visitor.toString(stmt.Index),
	), nil
}

// ProcessTextExpr implements Visitor.
func (visitor DumpVisitor) ProcessTextExpr(stmt *TextExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "value": "%s" }`, visitor.getKindAndPos(stmt), stmt.Value), nil
}

// ProcessThrowStmt implements Visitor.
func (visitor DumpVisitor) ProcessThrowStmt(stmt *ThrowStatement, _ any) (any, error) {
	return fmt.Sprintf(`{ %s, "expr": %s }`, visitor.getKindAndPos(stmt), visitor.toString(stmt.Expr)), nil
}

// ProcessTryStmt implements Visitor.
func (visitor DumpVisitor) ProcessTryStmt(stmt *TryStatement, _ any) (any, error) {
	catches := "["
	for _, catch := range stmt.Catches {
		if len(catches) > 1 {
			catches += ", "
		}
		catches += fmt.Sprintf(
			`{ "errorTypes": [%s], "variableName": "%s", "body": %s }`,
			common.ImplodeStrSlice(catch.ErrorType), catch.VariableName, visitor.toString(catch.Body),
		)
	}
	catches += "]"

	return fmt.Sprintf(
		`{ %s, "body": %s, "catches": %s, "finally": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Body), catches, visitor.toString(stmt.Finally),
	), nil
}

// ProcessUnaryExpr implements Visitor.
func (visitor DumpVisitor) ProcessUnaryExpr(stmt *UnaryOpExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "operator": "%s", "expr": %s }`,
		visitor.getKindAndPos(stmt), stmt.Operator, visitor.toString(stmt.Expr),
	), nil
}

// ProcessUnsetIntrinsicExpr implements Visitor.
func (visitor DumpVisitor) ProcessUnsetIntrinsicExpr(stmt *UnsetIntrinsicExpression, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "functionName": %s, "arguments": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.FunctionName), visitor.dumpExpressions(stmt.Arguments),
	), nil
}

// ProcessWhileStmt implements Visitor.
func (visitor DumpVisitor) ProcessWhileStmt(stmt *WhileStatement, _ any) (any, error) {
	return fmt.Sprintf(
		`{ %s, "condition": %s, "block": %s }`,
		visitor.getKindAndPos(stmt), visitor.toString(stmt.Condition), visitor.toString(stmt.Block),
	), nil
}

// ProcessVariableNameExpr implements Visitor.
func (visitor DumpVisitor) ProcessVariableNameExpr(stmt *VariableNameExpression, _ any) (any, error) {
	return fmt.Sprintf(`{ %s,  "variableName": "%s" }`, visitor.getKindAndPos(stmt), stmt.VariableName), nil
}
