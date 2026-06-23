package outputControl

import (
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/funcParamValidator"
	"QIQ/cmd/qiq/runtime/values"
)

func Register(environment runtime.Environment) {
	// Category: Output Control Functions
	environment.AddNativeFunction("ob_clean", nativeFn_ob_clean)
	environment.AddNativeFunction("ob_end_clean", nativeFn_ob_end_clean)
	environment.AddNativeFunction("ob_end_flush", nativeFn_ob_end_flush)
	environment.AddNativeFunction("ob_flush", nativeFn_ob_flush)
	environment.AddNativeFunction("ob_get_clean", nativeFn_ob_get_clean)
	environment.AddNativeFunction("ob_get_contents", nativeFn_ob_get_contents)
	environment.AddNativeFunction("ob_get_flush", nativeFn_ob_get_flush)
	environment.AddNativeFunction("ob_get_length", nativeFn_ob_get_length)
	environment.AddNativeFunction("ob_get_level", nativeFn_ob_get_level)
	environment.AddNativeFunction("ob_start", nativeFn_ob_start)
}

// -------------------------------------- ob_clean -------------------------------------- MARK: ob_clean

func nativeFn_ob_clean(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-clean.php

	_, err := funcParamValidator.NewValidator("ob_clean").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-clean.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_CLEAN flag),
	// discards it's return value and cleans (erases) the contents of the active output buffer.

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		context.Interpreter.PrintError(phpError.NewNotice("ob_clean(): Failed to delete buffer. No buffer to delete in %s", context.Stmt.GetPosString()))
		return values.NewBool(false), nil
	}

	context.Interpreter.GetOutputBufferStack().GetLast().Content = ""
	return values.NewBool(true), nil
}

// -------------------------------------- ob_end_clean -------------------------------------- MARK: ob_end_clean

func nativeFn_ob_end_clean(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-end-clean.php

	_, err := funcParamValidator.NewValidator("ob_end_clean").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-end-clean.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_CLEAN and PHP_OUTPUT_HANDLER_FINAL flags),
	// discards it's return value, discards the contents of the active output buffer and turns off the active output buffer.

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		context.Interpreter.PrintError(phpError.NewNotice("ob_end_clean(): Failed to delete buffer. No buffer to delete in %s", context.Stmt.GetPosString()))
		return values.NewBool(false), nil
	}

	context.Interpreter.GetOutputBufferStack().Pop()
	return values.NewBool(true), nil
}

// -------------------------------------- ob_end_flush -------------------------------------- MARK: ob_end_flush

func nativeFn_ob_end_flush(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-end-flush.php

	_, err := funcParamValidator.NewValidator("ob_end_flush").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return ObEndFlush(context)
}

func ObEndFlush(context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-end-flush.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_FINAL flag),
	// flushes (sends) it's return value, discards the contents of the active output buffer and turns off the active output buffer.

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		context.Interpreter.PrintError(phpError.NewNotice("ob_end_flush(): Failed to delete and flush buffer. No buffer to delete or flush in %s", context.Stmt.GetPosString()))
		return values.NewBool(false), nil
	}

	nativeFn_ob_flush([]values.RuntimeValue{}, context)
	context.Interpreter.GetOutputBufferStack().Pop()
	return values.NewBool(true), nil
}

// -------------------------------------- ob_flush -------------------------------------- MARK: ob_flush

func nativeFn_ob_flush(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-flush.php

	_, err := funcParamValidator.NewValidator("ob_flush").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-flush.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_FLUSH flag),
	// discards it's return value and flushs (erases) the contents of the active output buffer.

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		context.Interpreter.PrintError(phpError.NewNotice("ob_flush(): Failed to flush buffer. No buffer to flush in %s", context.Stmt.GetPosString()))
		return values.NewBool(false), nil
	}

	if context.Interpreter.GetOutputBufferStack().Len() == 1 {
		context.Interpreter.WriteResult(context.Interpreter.GetOutputBufferStack().GetLast().Content)
	} else {
		context.Interpreter.GetOutputBufferStack().Get(context.Interpreter.GetOutputBufferStack().Len() - 2).Content += context.Interpreter.GetOutputBufferStack().GetLast().Content
	}

	nativeFn_ob_clean(args, context)
	return values.NewBool(true), nil
}

// -------------------------------------- ob_get_clean -------------------------------------- MARK: ob_get_clean

func nativeFn_ob_get_clean(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-get-clean.php

	_, err := funcParamValidator.NewValidator("ob_get_clean").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-get-clean.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_CLEAN and PHP_OUTPUT_HANDLER_FINAL flags),
	// discards it's return value, returns the contents of the active output buffer and turns off the active output buffer.

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	content := context.Interpreter.GetOutputBufferStack().GetLast().Content
	context.Interpreter.GetOutputBufferStack().Pop()
	return values.NewStr(content), nil
}

// -------------------------------------- ob_get_contents -------------------------------------- MARK: ob_get_contents

func nativeFn_ob_get_contents(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-get-contents.php

	_, err := funcParamValidator.NewValidator("ob_get_contents").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	return values.NewStr(context.Interpreter.GetOutputBufferStack().GetLast().Content), nil
}

// -------------------------------------- ob_get_flush -------------------------------------- MARK: ob_get_flush

func nativeFn_ob_get_flush(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-get-flush.php

	_, err := funcParamValidator.NewValidator("ob_get_flush").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-get-flush.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_FINAL flag),
	// flushes (sends) it's return value, returns the contents of the active output buffer and turns off the active output buffer.

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		context.Interpreter.PrintError(phpError.NewNotice("ob_get_flush(): Failed to delete buffer. No buffer to delete in %s", context.Stmt.GetPosString()))
		return values.NewBool(false), nil
	}

	content := context.Interpreter.GetOutputBufferStack().GetLast().Content
	nativeFn_ob_end_flush(args, context)
	return values.NewStr(content), nil
}

// -------------------------------------- ob_get_length -------------------------------------- MARK: ob_get_length

func nativeFn_ob_get_length(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-get-length.php

	_, err := funcParamValidator.NewValidator("ob_get_length").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.ob-get-length.php
	// Returns ... false if no buffering is active.
	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	// Spec: https://www.php.net/manual/en/function.ob-get-length.php
	// Returns the length of the output buffer contents, in bytes
	return values.NewInt(int64(len(context.Interpreter.GetOutputBufferStack().GetLast().Content))), nil
}

// -------------------------------------- ob_get_level -------------------------------------- MARK: ob_get_level

func nativeFn_ob_get_level(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-get-level.php

	_, err := funcParamValidator.NewValidator("ob_get_level").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewInt(int64(context.Interpreter.GetOutputBufferStack().Len())), nil
}

// -------------------------------------- ob_start -------------------------------------- MARK: ob_start

func nativeFn_ob_start(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-start

	_, err := funcParamValidator.NewValidator("ob_start").Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO ob_start parameters
	//  ob_start(?callable $callback = null, int $chunk_size = 0, int $flags = PHP_OUTPUT_HANDLER_STDFLAGS): bool

	context.Interpreter.GetOutputBufferStack().Push()

	return values.NewBool(true), nil
}

// TODO flush
// TODO ob_​get_​status
// TODO ob_​implicit_​flush
// TODO ob_​list_​handlers
// TODO output_​add_​rewrite_​var
// TODO output_​reset_​rewrite_​vars
