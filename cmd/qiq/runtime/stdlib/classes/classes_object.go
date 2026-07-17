package classes

import (
	"QIQ/cmd/qiq/ast"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/funcParamValidator"
	"QIQ/cmd/qiq/runtime/values"
	"strings"
)

func Register(environment runtime.Environment) {
	// Category: Classes/Object Functions
	environment.AddNativeFunction("class_alias", nativeFn_class_alias)
	environment.AddNativeFunction("class_exists", nativeFn_class_exists)
	environment.AddNativeFunction("get_class", nativeFn_get_class)
	environment.AddNativeFunction("get_class_methods", nativeFn_get_class_methods)
	environment.AddNativeFunction("get_class_vars", nativeFn_get_class_vars)
	environment.AddNativeFunction("get_declared_classes", nativeFn_get_declared_classes)
	environment.AddNativeFunction("get_declared_interfaces", nativeFn_get_declared_interfaces)
	environment.AddNativeFunction("get_parent_class", nativeFn_get_parent_class)
	environment.AddNativeFunction("is_a", nativeFn_is_a)
	environment.AddNativeFunction("is_subclass_of", nativeFn_is_subclass_of)
	environment.AddNativeFunction("method_exists", nativeFn_method_exists)
	environment.AddNativeFunction("property_exists", nativeFn_property_exists)
}

// -------------------------------------- class_alias -------------------------------------- MARK: class_alias

func nativeFn_class_alias(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.class-alias.php

	args, err := funcParamValidator.NewValidator("class_alias").
		AddParam("$class", []string{"string"}, nil).
		AddParam("$alias", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO class_alias - Ad support for param $autoload

	className := args[0].(*values.Str).Value
	classDecl, found := context.Interpreter.GetClass(className)
	if !found {
		context.Interpreter.PrintError(phpError.NewWarning(`Class "%s" not found in %s`, className, context.Stmt.GetPosString()))
		return values.NewBool(false), nil
	}

	if err := context.Interpreter.AddClass(args[1].(*values.Str).Value, classDecl); err != nil {
		return values.NewVoid(), err
	}

	return values.NewBool(true), nil
}

// -------------------------------------- class_exists -------------------------------------- MARK: class_exists

func nativeFn_class_exists(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.class-exists.php

	args, err := funcParamValidator.NewValidator("class_exists").
		AddParam("$class", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	_, found := context.Interpreter.GetClass(args[0].(*values.Str).Value)

	return values.NewBool(found), nil
}

// -------------------------------------- get_class -------------------------------------- MARK: get_class

func nativeFn_get_class(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.get-class.php

	args, err := funcParamValidator.NewValidator("get_class").
		AddParam("$object", []string{"object"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	object := args[0].(*values.Object)

	// Spec: https://www.php.net/manual/en/function.get-class.php
	// If the object is an instance of a class which exists in a namespace,
	// the qualified namespaced name of that class is returned.
	return values.NewStr(object.Class.GetQualifiedName()), nil
}

// -------------------------------------- get_class_methods -------------------------------------- MARK: get_class_methods

func nativeFn_get_class_methods(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.get-class-methods.php

	args, err := funcParamValidator.NewValidator("get_class_methods").
		AddParam("$object_or_class", []string{"object", "string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	objectOrClass := args[0]

	var class *ast.ClassDeclarationStatement = nil

	if objectOrClass.GetType() == values.StrValue {
		className := objectOrClass.(*values.Str).Value
		var found bool
		class, found = context.Interpreter.GetClass(className)
		if !found {
			// TODO error handling - return a type error: https://www.php.net/manual/en/class.typeerror.php
			return values.NewVoid(), phpError.NewError("Uncaught Type Error: get_class_methods(): Argument #1 ($object_or_class) must be an object or a valid class name, string given in %s", context.Stmt.GetPosString())
		}
	}

	if objectOrClass.GetType() == values.ObjectValue {
		class = objectOrClass.(*values.Object).Class
	}

	array := values.NewArray()
	for _, methodName := range class.MethodNames {
		array.SetElement(nil, values.NewStr(methodName))
	}

	return array, nil
}

// -------------------------------------- get_class_vars -------------------------------------- MARK: get_class_vars

func nativeFn_get_class_vars(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.get-class-vars.php

	args, err := funcParamValidator.NewValidator("get_class_vars").
		AddParam("$class", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	className := args[0].(*values.Str).Value

	class, found := context.Interpreter.GetClass(className)
	if !found {
		// TODO error handling - return a type error: https://www.php.net/manual/en/class.typeerror.php
		return values.NewVoid(), phpError.NewError("Uncaught Type Error: get_class_vars(): Argument #1 ($class) must be a valid class name, %s given in %s", className, context.Stmt.GetPosString())
	}

	array := values.NewArray()

	for _, propertyName := range class.PropertieNames {
		if class.Properties[propertyName].Visibility != "public" {
			continue
		}

		if class.Properties[propertyName].InitialValue == nil {
			array.SetElement(values.NewStr(propertyName[1:]), values.NewNull())
			continue
		}

		slot, err := context.Interpreter.ProcessStatement(class.Properties[propertyName].InitialValue, context.Env)
		if err != nil {
			return array, err
		}
		array.SetElement(values.NewStr(propertyName[1:]), slot.Value)
	}

	return array, nil
}

// -------------------------------------- get_declared_classes -------------------------------------- MARK: get_declared_classes

func nativeFn_get_declared_classes(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.get-declared-classes.php

	_, err := funcParamValidator.NewValidator("get_declared_classes").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	classes := values.NewArray()
	for _, className := range context.Interpreter.GetClasses() {
		classes.SetElement(nil, values.NewStr(className))
	}

	return classes, nil
}

// -------------------------------------- get_declared_interfaces -------------------------------------- MARK: get_declared_interfaces

func nativeFn_get_declared_interfaces(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.get-declared-interfaces.php

	_, err := funcParamValidator.NewValidator("get_declared_interfaces").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	interfaces := values.NewArray()
	for _, interfaceName := range context.Interpreter.GetInterfaces() {
		interfaces.SetElement(nil, values.NewStr(interfaceName))
	}

	return interfaces, nil
}

// -------------------------------------- get_parent_class -------------------------------------- MARK: get_parent_class

func nativeFn_get_parent_class(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.get-parent-class.php

	args, err := funcParamValidator.NewValidator("get_parent_class").
		AddParam("$object_or_class", []string{"object", "string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	objectOrClass := args[0]

	var class *ast.ClassDeclarationStatement = nil

	if objectOrClass.GetType() == values.StrValue {
		className := objectOrClass.(*values.Str).Value
		var found bool
		class, found = context.Interpreter.GetClass(className)
		if !found {
			return values.NewBool(false), nil
		}
	}

	if objectOrClass.GetType() == values.ObjectValue {
		class = objectOrClass.(*values.Object).Class
	}

	if class.BaseClass == "" {
		return values.NewBool(false), nil
	}

	return values.NewStr(class.BaseClass), nil
}

// -------------------------------------- is_a -------------------------------------- MARK: is_a

func nativeFn_is_a(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.is-a.php

	args, err := funcParamValidator.NewValidator("is_a").
		AddParam("$object_or_class", []string{"object", "string"}, nil).
		AddParam("$class", []string{"string"}, nil).
		AddParam("$allow_string", []string{"bool"}, values.NewBool(false)).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	objectOrClass := args[0]
	class := args[1].(*values.Str).Value
	allowString := args[2].(*values.Bool).Value

	// Always return false, if a string is not allowed
	if !allowString && objectOrClass.GetType() == values.StrValue {
		return values.NewBool(false), nil
	}

	if objectOrClass.GetType() == values.StrValue {
		// Always return false, if the given string is not a valid class
		_, found := context.Interpreter.GetClass(objectOrClass.(*values.Str).Value)
		if !found {
			return values.NewBool(false), nil
		}

		return values.NewBool(strings.EqualFold(objectOrClass.(*values.Str).Value, class)), nil
	}

	return values.NewBool(strings.EqualFold(objectOrClass.(*values.Object).Class.GetQualifiedName(), class)), nil
}

// -------------------------------------- is_subclass_of -------------------------------------- MARK: is_subclass_of

func nativeFn_is_subclass_of(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.is-subclass-of.php

	args, err := funcParamValidator.NewValidator("is_subclass_of").
		AddParam("$object_or_class", []string{"object", "string"}, nil).
		AddParam("$class", []string{"string"}, nil).
		AddParam("$allow_string", []string{"bool"}, values.NewBool(true)).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	objectOrClass := args[0]
	class := args[1].(*values.Str).Value
	allowString := args[2].(*values.Bool).Value

	if !allowString && objectOrClass.GetType() == values.StrValue {
		return values.NewVoid(), phpError.NewError("is_subclass_of: $object_or_class must be an object")
	}

	var classDecl *ast.ClassDeclarationStatement = nil

	if objectOrClass.GetType() == values.StrValue {
		className := objectOrClass.(*values.Str).Value
		var found bool
		classDecl, found = context.Interpreter.GetClass(className)
		if !found {
			return values.NewBool(false), nil
		}
	}

	if objectOrClass.GetType() == values.ObjectValue {
		classDecl = objectOrClass.(*values.Object).Class
	}

	// TODO is_subclass_of - check if it or parent implements it
	// TODO is_subclass_of - check if some parent is this class

	return values.NewBool(strings.EqualFold(classDecl.BaseClass, class)), nil
}

// -------------------------------------- method_exists -------------------------------------- MARK: method_exists

func nativeFn_method_exists(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.method-exists.php

	args, err := funcParamValidator.NewValidator("method_exists").
		AddParam("$object_or_class", []string{"object", "string"}, nil).
		AddParam("$method", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	objectOrClass := args[0]
	methodName := args[1].(*values.Str).Value

	var classDecl *ast.ClassDeclarationStatement = nil

	if objectOrClass.GetType() == values.StrValue {
		className := objectOrClass.(*values.Str).Value

		// Always return false, if the given string is not a valid class
		var found bool
		classDecl, found = context.Interpreter.GetClass(className)
		if !found {
			return values.NewBool(false), nil
		}
	}

	if objectOrClass.GetType() == values.ObjectValue {
		classDecl = objectOrClass.(*values.Object).Class
	}

	_, found := classDecl.Methods[strings.ToLower(methodName)]
	return values.NewBool(found), nil
}

// -------------------------------------- property_exists -------------------------------------- MARK: property_exists

func nativeFn_property_exists(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.property-exists.php

	args, err := funcParamValidator.NewValidator("property_exists").
		AddParam("$object_or_class", []string{"object", "string"}, nil).
		AddParam("$property", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	objectOrClass := args[0]
	propertyName := args[1].(*values.Str).Value

	var classDecl *ast.ClassDeclarationStatement = nil

	if objectOrClass.GetType() == values.StrValue {
		className := objectOrClass.(*values.Str).Value

		// Always return false, if the given string is not a valid class
		var found bool
		classDecl, found = context.Interpreter.GetClass(className)
		if !found {
			return values.NewBool(false), nil
		}
	}

	if objectOrClass.GetType() == values.ObjectValue {
		classDecl = objectOrClass.(*values.Object).Class
	}

	_, found := classDecl.Properties["$"+propertyName]
	return values.NewBool(found), nil
}

// TODO enum_exists
// TODO get_called_class
// TODO get_declared_traits
// TODO get_mangled_object_vars
// TODO get_object_vars
// TODO interface_exists
// TODO trait_exists
