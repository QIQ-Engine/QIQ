package classes

import (
	"QIQ/cmd/qiq/ast"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/runtime"
)

func RegisterDefaultClasses(interpreter runtime.Interpreter) phpError.Error {
	// -------------------------------------- Exception -------------------------------------- MARK: Exception

	// Spec: https://www.php.net/manual/en/class.exception.php
	Exception := ast.NewClassDeclarationStmt(0, nil, "Exception", false, false)
	Exception.Interfaces = append(Exception.Interfaces, "Throwable")
	Exception.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$message", "protected", false, []string{"string"}, ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)))
	Exception.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$string", "private", false, []string{"string"}, ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)))
	Exception.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$code", "protected", false, []string{"int"}, nil))
	Exception.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$file", "protected", false, []string{"string"}, ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)))
	Exception.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$line", "protected", false, []string{"int"}, nil))
	Exception.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$trace", "private", false, []string{"array"}, ast.NewArrayLiteralExpr(0, nil)))
	Exception.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$previous", "private", false, []string{"null", "Throwable"}, ast.NewConstantAccessExpr(0, nil, "NULL")))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "__construct", []string{"public"}, []ast.FunctionParameter{{Name: "$message", Type: []string{"string"}, DefaultValue: ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)}, {Name: "$code", Type: []string{"int"}, DefaultValue: ast.NewIntegerLiteralExpr(0, nil, 0)}, {Name: "$previous", Type: []string{"null", "Throwable"}, DefaultValue: ast.NewConstantAccessExpr(0, nil, "NULL")}}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewExpressionStmt(0, ast.NewSimpleAssignmentExpr(0, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "message")), ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$message")))), ast.NewExpressionStmt(0, ast.NewSimpleAssignmentExpr(0, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "code")), ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$code")))), ast.NewExpressionStmt(0, ast.NewSimpleAssignmentExpr(0, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "previous")), ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$previous"))))}), []string{}))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getMessage", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "message")))}), []string{"string"}))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getPrevious", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "previous")))}), []string{"null", "Throwable"}))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getCode", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "code")))}), []string{"int"}))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getFile", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "file")))}), []string{"string"}))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getLine", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "line")))}), []string{"int"}))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getTrace", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "trace")))}), []string{"array"}))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getTraceAsString", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewFunctionCallExpr(0, nil, ast.NewStringLiteralExpr(0, nil, "implode", ast.DoubleQuotedString), []ast.IExpression{ast.NewConstantAccessExpr(0, nil, "PHP_EOL"), ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "trace"))}))}), []string{"string"}))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "__toString", []string{"public"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "string")))}), []string{"string"}))
	Exception.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "__clone", []string{"private"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{}), []string{"void"}))

	if err := interpreter.AddClass(Exception.Name, Exception); err != nil {
		return err
	}

	// -------------------------------------- ErrorException -------------------------------------- MARK: ErrorException

	// Spec: https://www.php.net/manual/en/class.errorexception.php
	ErrorException := ast.NewClassDeclarationStmt(0, nil, "ErrorException", false, false)
	ErrorException.BaseClass = "Exception"
	ErrorException.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$severity", "protected", false, []string{"int"}, ast.NewConstantAccessExpr(0, nil, "E_ERROR")))
	ErrorException.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "__construct", []string{"public"}, []ast.FunctionParameter{{Name: "$message", Type: []string{"string"}, DefaultValue: ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)}, {Name: "$code", Type: []string{"int"}, DefaultValue: ast.NewIntegerLiteralExpr(0, nil, 0)}, {Name: "$severity", Type: []string{"int"}, DefaultValue: ast.NewConstantAccessExpr(0, nil, "E_ERROR")}, {Name: "$filename", Type: []string{"null", "string"}, DefaultValue: ast.NewConstantAccessExpr(0, nil, "NULL")}, {Name: "$line", Type: []string{"null", "int"}, DefaultValue: ast.NewConstantAccessExpr(0, nil, "NULL")}, {Name: "$previous", Type: []string{"null", "Throwable"}, DefaultValue: ast.NewConstantAccessExpr(0, nil, "NULL")}}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewExpressionStmt(0, ast.NewMemberAccessExpr(0, nil, ast.NewConstantAccessExpr(0, nil, "parent"), ast.NewFunctionCallExpr(0, nil, ast.NewStringLiteralExpr(0, nil, "__construct", ast.DoubleQuotedString), []ast.IExpression{ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$message")), ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$code")), ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$previous"))}))), ast.NewExpressionStmt(0, ast.NewSimpleAssignmentExpr(0, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "severity")), ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$severity")))), ast.NewExpressionStmt(0, ast.NewSimpleAssignmentExpr(0, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "file")), ast.NewCoalesceExpr(0, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$filename")), ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)))), ast.NewExpressionStmt(0, ast.NewSimpleAssignmentExpr(0, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "line")), ast.NewCoalesceExpr(0, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$line")), ast.NewIntegerLiteralExpr(0, nil, 0))))}), []string{}))
	ErrorException.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getSeverity", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "severity")))}), []string{"int"}))

	if err := interpreter.AddClass(ErrorException.Name, ErrorException); err != nil {
		return err
	}

	// -------------------------------------- Error -------------------------------------- MARK: Error

	// Spec: https://www.php.net/manual/en/class.error.php
	Error := ast.NewClassDeclarationStmt(0, nil, "Error", false, false)
	Error.Interfaces = append(Error.Interfaces, "Throwable")
	Error.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$message", "protected", false, []string{"string"}, ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)))
	Error.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$string", "private", false, []string{"string"}, ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)))
	Error.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$code", "protected", false, []string{"int"}, nil))
	Error.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$file", "protected", false, []string{"string"}, ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)))
	Error.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$line", "protected", false, []string{"int"}, nil))
	Error.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$trace", "private", false, []string{"array"}, ast.NewArrayLiteralExpr(0, nil)))
	Error.AddProperty(ast.NewPropertyDeclarationStmt(0, nil, "$previous", "private", false, []string{"null", "Throwable"}, ast.NewConstantAccessExpr(0, nil, "NULL")))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "__construct", []string{"public"}, []ast.FunctionParameter{{Name: "$message", Type: []string{"string"}, DefaultValue: ast.NewStringLiteralExpr(0, nil, "", ast.DoubleQuotedString)}, {Name: "$code", Type: []string{"int"}, DefaultValue: ast.NewIntegerLiteralExpr(0, nil, 0)}, {Name: "$previous", Type: []string{"null", "Throwable"}, DefaultValue: ast.NewConstantAccessExpr(0, nil, "NULL")}}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewExpressionStmt(0, ast.NewSimpleAssignmentExpr(0, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "message")), ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$message")))), ast.NewExpressionStmt(0, ast.NewSimpleAssignmentExpr(0, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "code")), ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$code")))), ast.NewExpressionStmt(0, ast.NewSimpleAssignmentExpr(0, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "previous")), ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$previous"))))}), []string{}))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getMessage", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "message")))}), []string{"string"}))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getPrevious", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "previous")))}), []string{"null", "Throwable"}))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getCode", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "code")))}), []string{"int"}))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getFile", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "file")))}), []string{"string"}))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getLine", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "line")))}), []string{"int"}))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getTrace", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "trace")))}), []string{"array"}))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "getTraceAsString", []string{"public", "final"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewFunctionCallExpr(0, nil, ast.NewStringLiteralExpr(0, nil, "implode", ast.DoubleQuotedString), []ast.IExpression{ast.NewConstantAccessExpr(0, nil, "PHP_EOL"), ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "trace"))}))}), []string{"string"}))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "__toString", []string{"public"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{ast.NewReturnStmt(0, nil, ast.NewMemberAccessExpr(0, nil, ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$this")), ast.NewConstantAccessExpr(0, nil, "string")))}), []string{"string"}))
	Error.AddMethod(ast.NewMethodDefinitionStmt(0, nil, "__clone", []string{"private"}, []ast.FunctionParameter{}, ast.NewCompoundStmt(0, []ast.IStatement{}), []string{"void"}))

	if err := interpreter.AddClass(Error.Name, Error); err != nil {
		return err
	}

	// -------------------------------------- CompileError -------------------------------------- MARK: CompileError

	// Spec: https://www.php.net/manual/en/class.compileerror.php
	CompileError := ast.NewClassDeclarationStmt(0, nil, "CompileError", false, false)
	CompileError.BaseClass = "Error"

	if err := interpreter.AddClass(CompileError.Name, CompileError); err != nil {
		return err
	}

	// -------------------------------------- ParseError -------------------------------------- MARK: ParseError

	// Spec: https://www.php.net/manual/en/class.parseerror.php
	ParseError := ast.NewClassDeclarationStmt(0, nil, "ParseError", false, false)
	ParseError.BaseClass = "CompileError"

	if err := interpreter.AddClass(ParseError.Name, ParseError); err != nil {
		return err
	}

	// -------------------------------------- TypeError -------------------------------------- MARK: TypeError

	// Spec: https://www.php.net/manual/en/class.typeerror.php
	TypeError := ast.NewClassDeclarationStmt(0, nil, "TypeError", false, false)
	TypeError.BaseClass = "Error"

	if err := interpreter.AddClass(TypeError.Name, TypeError); err != nil {
		return err
	}

	// -------------------------------------- ArgumentCountError -------------------------------------- MARK: ArgumentCountError

	// Spec: https://www.php.net/manual/en/class.argumentcounterror.php
	ArgumentCountError := ast.NewClassDeclarationStmt(0, nil, "ArgumentCountError", false, false)
	ArgumentCountError.BaseClass = "TypeError"

	if err := interpreter.AddClass(ArgumentCountError.Name, ArgumentCountError); err != nil {
		return err
	}

	// -------------------------------------- ValueError -------------------------------------- MARK: ValueError

	// Spec: https://www.php.net/manual/en/class.valueerror.php
	ValueError := ast.NewClassDeclarationStmt(0, nil, "ValueError", false, false)
	ValueError.BaseClass = "Error"

	if err := interpreter.AddClass(ValueError.Name, ValueError); err != nil {
		return err
	}

	// -------------------------------------- ArithmeticError -------------------------------------- MARK: ArithmeticError

	// Spec: https://www.php.net/manual/en/class.arithmeticerror.php

	ArithmeticError := ast.NewClassDeclarationStmt(0, nil, "ArithmeticError", false, false)
	ArithmeticError.BaseClass = "Error"

	if err := interpreter.AddClass(ArithmeticError.Name, ArithmeticError); err != nil {
		return err
	}

	// -------------------------------------- DivisionByZeroError -------------------------------------- MARK: DivisionByZeroError

	// Spec: https://www.php.net/manual/en/class.divisionbyzeroerror.php
	DivisionByZeroError := ast.NewClassDeclarationStmt(0, nil, "DivisionByZeroError", false, false)
	DivisionByZeroError.BaseClass = "ArithmeticError"

	if err := interpreter.AddClass(DivisionByZeroError.Name, DivisionByZeroError); err != nil {
		return err
	}

	// -------------------------------------- UnhandledMatchError -------------------------------------- MARK: UnhandledMatchError

	// Spec: https://www.php.net/manual/en/class.unhandledmatcherror.php
	UnhandledMatchError := ast.NewClassDeclarationStmt(0, nil, "UnhandledMatchError", false, false)
	UnhandledMatchError.BaseClass = "Error"

	if err := interpreter.AddClass(UnhandledMatchError.Name, UnhandledMatchError); err != nil {
		return err
	}

	// -------------------------------------- RequestParseBodyException -------------------------------------- MARK: RequestParseBodyException

	// Spec: https://www.php.net/manual/en/class.requestparsebodyexception.php
	RequestParseBodyException := ast.NewClassDeclarationStmt(0, nil, "RequestParseBodyException", false, false)
	RequestParseBodyException.BaseClass = "Exception"

	if err := interpreter.AddClass(RequestParseBodyException.Name, RequestParseBodyException); err != nil {
		return err
	}

	// -------------------------------------- ClosedGeneratorException -------------------------------------- MARK: ClosedGeneratorException

	// Spec: https://www.php.net/manual/en/class.closedgeneratorexception.php
	ClosedGeneratorException := ast.NewClassDeclarationStmt(0, nil, "ClosedGeneratorException", false, false)
	ClosedGeneratorException.BaseClass = "Exception"

	if err := interpreter.AddClass(ClosedGeneratorException.Name, ClosedGeneratorException); err != nil {
		return err
	}

	// -------------------------------------- FiberError -------------------------------------- MARK: FiberError

	// Spec: https://www.php.net/manual/en/class.fibererror.php
	FiberError := ast.NewClassDeclarationStmt(0, nil, "FiberError", false, false)
	FiberError.BaseClass = "Error"

	if err := interpreter.AddClass(FiberError.Name, FiberError); err != nil {
		return err
	}

	// -------------------------------------- stdClass -------------------------------------- MARK: stdClass

	// Spec: https://www.php.net/manual/en/class.stdclass.php
	stdClass := ast.NewClassDeclarationStmt(0, nil, "stdClass", false, false)

	if err := interpreter.AddClass(stdClass.Name, stdClass); err != nil {
		return err
	}

	// -------------------------------------- JsonException -------------------------------------- MARK: JsonException

	// Spec: https://www.php.net/manual/en/class.jsonexception.php
	JsonException := ast.NewClassDeclarationStmt(0, nil, "JsonException", false, false)
	JsonException.BaseClass = "Exception"

	if err := interpreter.AddClass(JsonException.Name, JsonException); err != nil {
		return err
	}

	// -------------------------------------- ReflectionException -------------------------------------- MARK: ReflectionException

	// Spec: https://www.php.net/manual/en/class.reflectionexception.php
	ReflectionException := ast.NewClassDeclarationStmt(0, nil, "ReflectionException", false, false)
	ReflectionException.BaseClass = "Exception"

	if err := interpreter.AddClass(ReflectionException.Name, ReflectionException); err != nil {
		return err
	}

	// -------------------------------------- LogicException -------------------------------------- MARK: LogicException

	// Spec: https://www.php.net/manual/en/class.logicexception.php
	LogicException := ast.NewClassDeclarationStmt(0, nil, "LogicException", false, false)
	LogicException.BaseClass = "Exception"

	if err := interpreter.AddClass(LogicException.Name, LogicException); err != nil {
		return err
	}

	// -------------------------------------- BadFunctionCallException -------------------------------------- MARK: BadFunctionCallException

	// Spec: https://www.php.net/manual/en/class.badfunctioncallexception.php
	BadFunctionCallException := ast.NewClassDeclarationStmt(0, nil, "BadFunctionCallException", false, false)
	BadFunctionCallException.BaseClass = "LogicException"

	if err := interpreter.AddClass(BadFunctionCallException.Name, BadFunctionCallException); err != nil {
		return err
	}

	// -------------------------------------- BadMethodCallException -------------------------------------- MARK: BadMethodCallException

	// Spec: https://www.php.net/manual/en/class.badmethodcallexception.php
	BadMethodCallException := ast.NewClassDeclarationStmt(0, nil, "BadMethodCallException", false, false)
	BadMethodCallException.BaseClass = "BadFunctionCallException"

	if err := interpreter.AddClass(BadMethodCallException.Name, BadMethodCallException); err != nil {
		return err
	}

	// -------------------------------------- DomainException -------------------------------------- MARK: DomainException

	// Spec: https://www.php.net/manual/en/class.domainexception.php
	DomainException := ast.NewClassDeclarationStmt(0, nil, "DomainException", false, false)
	DomainException.BaseClass = "LogicException"

	if err := interpreter.AddClass(DomainException.Name, DomainException); err != nil {
		return err
	}

	// -------------------------------------- InvalidArgumentException -------------------------------------- MARK: InvalidArgumentException

	// Spec: https://www.php.net/manual/en/class.invalidargumentexception.php
	InvalidArgumentException := ast.NewClassDeclarationStmt(0, nil, "InvalidArgumentException", false, false)
	InvalidArgumentException.BaseClass = "LogicException"

	if err := interpreter.AddClass(InvalidArgumentException.Name, InvalidArgumentException); err != nil {
		return err
	}

	// -------------------------------------- LengthException -------------------------------------- MARK: LengthException

	// Spec: https://www.php.net/manual/en/class.lengthexception.php
	LengthException := ast.NewClassDeclarationStmt(0, nil, "LengthException", false, false)
	LengthException.BaseClass = "LogicException"

	if err := interpreter.AddClass(LengthException.Name, LengthException); err != nil {
		return err
	}

	// -------------------------------------- OutOfRangeException -------------------------------------- MARK: OutOfRangeException

	// Spec: https://www.php.net/manual/en/class.outofrangeexception.php
	OutOfRangeException := ast.NewClassDeclarationStmt(0, nil, "OutOfRangeException", false, false)
	OutOfRangeException.BaseClass = "LogicException"

	if err := interpreter.AddClass(OutOfRangeException.Name, OutOfRangeException); err != nil {
		return err
	}

	// -------------------------------------- RuntimeException -------------------------------------- MARK: RuntimeException

	// Spec: https://www.php.net/manual/en/class.runtimeexception.php
	RuntimeException := ast.NewClassDeclarationStmt(0, nil, "RuntimeException", false, false)
	RuntimeException.BaseClass = "Exception"

	if err := interpreter.AddClass(RuntimeException.Name, RuntimeException); err != nil {
		return err
	}

	// -------------------------------------- OutOfBoundsException -------------------------------------- MARK: OutOfBoundsException

	// Spec: https://www.php.net/manual/en/class.outofboundsexception.php
	OutOfBoundsException := ast.NewClassDeclarationStmt(0, nil, "OutOfBoundsException", false, false)
	OutOfBoundsException.BaseClass = "RuntimeException"

	if err := interpreter.AddClass(OutOfBoundsException.Name, OutOfBoundsException); err != nil {
		return err
	}

	// -------------------------------------- OverflowException -------------------------------------- MARK: OverflowException

	// Spec: https://www.php.net/manual/en/class.overflowexception.php
	OverflowException := ast.NewClassDeclarationStmt(0, nil, "OverflowException", false, false)
	OverflowException.BaseClass = "RuntimeException"

	if err := interpreter.AddClass(OverflowException.Name, OverflowException); err != nil {
		return err
	}

	// -------------------------------------- RangeException -------------------------------------- MARK: RangeException

	// Spec: https://www.php.net/manual/en/class.rangeexception.php
	RangeException := ast.NewClassDeclarationStmt(0, nil, "RangeException", false, false)
	RangeException.BaseClass = "RuntimeException"

	if err := interpreter.AddClass(RangeException.Name, RangeException); err != nil {
		return err
	}

	// -------------------------------------- UnderflowException -------------------------------------- MARK: UnderflowException

	// Spec: https://www.php.net/manual/en/class.underflowexception.php
	UnderflowException := ast.NewClassDeclarationStmt(0, nil, "UnderflowException", false, false)
	UnderflowException.BaseClass = "RuntimeException"

	if err := interpreter.AddClass(UnderflowException.Name, UnderflowException); err != nil {
		return err
	}

	// -------------------------------------- UnexpectedValueException -------------------------------------- MARK: UnexpectedValueException

	// Spec: https://www.php.net/manual/en/class.unexpectedvalueexception.php
	UnexpectedValueException := ast.NewClassDeclarationStmt(0, nil, "UnexpectedValueException", false, false)
	UnexpectedValueException.BaseClass = "RuntimeException"

	if err := interpreter.AddClass(UnexpectedValueException.Name, UnexpectedValueException); err != nil {
		return err
	}

	// -------------------------------------- AssertionError -------------------------------------- MARK: AssertionError

	// Spec: https://www.php.net/manual/en/class.assertionerror.php
	AssertionError := ast.NewClassDeclarationStmt(0, nil, "AssertionError", false, false)
	AssertionError.BaseClass = "Error"

	if err := interpreter.AddClass(AssertionError.Name, AssertionError); err != nil {
		return err
	}

	return nil
}
