package ctype

import (
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/funcParamValidator"
	"QIQ/cmd/qiq/runtime/values"
)

func Register(environment runtime.Environment) {
	// Category: Character type checking Functions
	environment.AddNativeFunction("ctype_alnum", nativeFn_ctype_alnum)
	environment.AddNativeFunction("ctype_alpha", nativeFn_ctype_alpha)
	environment.AddNativeFunction("ctype_cntrl", nativeFn_ctype_cntrl)
	environment.AddNativeFunction("ctype_digit", nativeFn_ctype_digit)
	environment.AddNativeFunction("ctype_graph", nativeFn_ctype_graph)
	environment.AddNativeFunction("ctype_lower", nativeFn_ctype_lower)
	environment.AddNativeFunction("ctype_print", nativeFn_ctype_print)
	environment.AddNativeFunction("ctype_punct", nativeFn_ctype_punct)
	environment.AddNativeFunction("ctype_space", nativeFn_ctype_space)
	environment.AddNativeFunction("ctype_upper", nativeFn_ctype_upper)
	environment.AddNativeFunction("ctype_xdigit", nativeFn_ctype_xdigit)
}

// -------------------------------------- ctype_alnum -------------------------------------- MARK: ctype_alnum

func nativeFn_ctype_alnum(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_alnum").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-alnum.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9')) {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_alpha -------------------------------------- MARK: ctype_alpha

func nativeFn_ctype_alpha(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_alpha").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-alpha.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z')) {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_cntrl -------------------------------------- MARK: ctype_cntrl

func nativeFn_ctype_cntrl(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_cntrl").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-cntrl.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if !(character <= 31 || character == 127) {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_digit -------------------------------------- MARK: ctype_digit

func nativeFn_ctype_digit(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_digit").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-digit.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if !(character >= '0' && character <= '9') {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_graph -------------------------------------- MARK: ctype_graph

func nativeFn_ctype_graph(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_graph").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-graph.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if character < 33 || character > 126 {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_lower -------------------------------------- MARK: ctype_lower

func nativeFn_ctype_lower(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_lower").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-lower.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if !(character >= 'a' && character <= 'z') {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_print -------------------------------------- MARK: ctype_print

func nativeFn_ctype_print(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_print").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-print.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if character < 32 || character > 126 {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_punct -------------------------------------- MARK: ctype_punct

func nativeFn_ctype_punct(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_punct").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-punct.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if character < 33 || character > 126 ||
			(character >= '0' && character <= '9') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= 'a' && character <= 'z') {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_space -------------------------------------- MARK: ctype_space

func nativeFn_ctype_space(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_space").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-space.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if character != ' ' && character != '\t' && character != '\v' &&
			character != '\n' && character != '\r' && character != '\f' {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_upper -------------------------------------- MARK: ctype_upper

func nativeFn_ctype_upper(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_upper").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-upper.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if !(character >= 'A' && character <= 'Z') {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}

// -------------------------------------- ctype_xdigit -------------------------------------- MARK: ctype_xdigit

func nativeFn_ctype_xdigit(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	_, err := funcParamValidator.NewValidator("ctype_xdigit").
		AddParam("$text", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ctype-xdigit.php

	text := args[0].(*values.Str).Value

	if len(text) == 0 {
		return values.NewBool(false), nil
	}

	for i := 0; i < len(text); i++ {
		character := text[i]
		if !((character >= 'a' && character <= 'f') ||
			(character >= 'A' && character <= 'F') ||
			(character >= '0' && character <= '9')) {
			return values.NewBool(false), nil
		}
	}

	return values.NewBool(true), nil
}
