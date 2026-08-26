package interpreter

import (
	"QIQ/cmd/qiq/ast"
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/config"
	"QIQ/cmd/qiq/ini"
	"QIQ/cmd/qiq/parser"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/request"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/classes"
	"QIQ/cmd/qiq/runtime/interfaces"
	"QIQ/cmd/qiq/runtime/outputBuffer"
	"QIQ/cmd/qiq/runtime/values"
	"QIQ/cmd/qiq/stats"
	"path/filepath"
	"slices"
)

var _ ast.Visitor = &Interpreter{}

type Interpreter struct {
	executionContext   *runtime.ExecutionContext
	filename           string
	includedFiles      []string
	ini                *ini.Ini
	request            *request.Request
	response           *request.Response
	parser             *parser.Parser
	env                *Environment
	cache              map[ast.NodeId]values.RuntimeValue
	outputBufferStack  *outputBuffer.Stack
	result             string
	resultRuntimeValue values.RuntimeValue
	workingDir         string
	// Status
	suppressWarning bool
	exitCalled      bool
}

func NewInterpreter(executionContext *runtime.ExecutionContext, ini *ini.Ini, r *request.Request, filename string) (*Interpreter, phpError.Error) {
	interpreter := &Interpreter{
		executionContext:  executionContext,
		filename:          filename,
		includedFiles:     []string{},
		ini:               ini,
		request:           r,
		response:          request.NewResponse(),
		parser:            parser.NewParser(ini),
		cache:             map[ast.NodeId]values.RuntimeValue{},
		outputBufferStack: outputBuffer.NewStack(),
	}

	if filename != "" {
		interpreter.workingDir = filepath.Dir(filename)
	}

	var err phpError.Error
	interpreter.env, err = NewEnvironment(nil, r, interpreter)
	if err != nil {
		return interpreter, err
	}

	if !interpreter.executionContext.IsInitializationCompleted() {
		if err := interfaces.RegisterDefaultInterfaces(interpreter); err != nil {
			return interpreter, err
		}

		if err := classes.RegisterDefaultClasses(interpreter); err != nil {
			return interpreter, err
		}

		interpreter.executionContext.InitializationCompleted()
	}

	if ini.GetBool("register_argc_argv") {
		server := interpreter.env.predefinedVariables["$_SERVER"].Value.(*values.Array)
		if !server.Contains(values.NewStr("argc")) && !server.Contains(values.NewStr("argv")) {
			server.SetElement(values.NewStr("argv"), interpreter.env.predefinedVariables["$_GET"].Value)
			server.SetElement(values.NewStr("argc"), values.NewInt(int64(len(interpreter.env.predefinedVariables["$_GET"].Value.(*values.Array).Keys))))
		}
	}

	return interpreter, nil
}

func (interpreter *Interpreter) GetExectionContext() *runtime.ExecutionContext {
	return interpreter.executionContext
}

func (interpreter *Interpreter) GetFilename() string { return interpreter.filename }

func (interpreter *Interpreter) GetWorkingDir() string { return interpreter.workingDir }

func (interpreter *Interpreter) GetIni() *ini.Ini { return interpreter.ini }

func (interpreter *Interpreter) GetRequest() *request.Request { return interpreter.request }

func (interpreter *Interpreter) GetResponse() *request.Response { return interpreter.response }

func (interpreter *Interpreter) GetOutputBufferStack() *outputBuffer.Stack {
	return interpreter.outputBufferStack
}

func (interpreter *Interpreter) Process(sourceCode string) (string, phpError.Error) {
	return interpreter.process(sourceCode, interpreter.env, false)
}

func (interpreter *Interpreter) process(sourceCode string, env *Environment, resetResult bool) (string, phpError.Error) {
	if resetResult {
		interpreter.result = ""
	}

	program, err := interpreter.parser.ProduceAST(sourceCode, interpreter.filename)
	if err != nil {
		return interpreter.result, err
	}

	stat := stats.Start()
	defer stats.StopAndPrint(stat, "Interpreter")

	if config.ShowParserCallStack {
		println("")
		println("Interpreter callstack")
		println("---------------------")
	}

	interpreter.resultRuntimeValue = nil
	slot, err := interpreter.processProgram(program, env)
	interpreter.resultRuntimeValue = slot.Value

	return interpreter.result, err
}

func (interpreter *Interpreter) processProgram(program *ast.Program, env *Environment) (*values.Slot, phpError.Error) {
	err := interpreter.scanForFunctionDefinition(program.GetStatements(), env)
	if err != nil {
		return values.NewVoidSlot(), err
	}

	defer interpreter.flushOutputBuffers()

	slot := values.NewSlot(nil)
	for _, stmt := range program.GetStatements() {
		if slot, err = interpreter.processStmt(stmt, env); err != nil {
			// Handle exit event - Stop code execution
			if err.GetErrorType() == phpError.EventError && err.GetMessage() == phpError.ExitEvent {
				break
			}
			return slot, err
		}
	}

	if !interpreter.exitCalled {
		interpreter.ProcessExitIntrinsicExpr(ast.NewExitIntrinsic(0, nil, ast.NewIntegerLiteralExpr(0, nil, 0)), interpreter.env)
	}

	return slot, nil
}

func (interperter *Interpreter) ProcessStatement(stmt ast.IStatement, env any) (*values.Slot, phpError.Error) {
	return interperter.processStmt(stmt, env)
}

func (interpreter *Interpreter) processStmt(stmt ast.IStatement, env any) (slot *values.Slot, phpErr phpError.Error) {
	defer func() {
		if r := recover(); r != nil {
			slot = r.(SlotOrError).Slot
			phpErr = r.(SlotOrError).Error.(phpError.Error)
		}
	}()

	ast.PrintInterpreterCallstack(stmt)
	result, err := stmt.Process(interpreter, env)
	if err != nil {
		phpErr = err.(phpError.Error)
	}
	if _, ok := result.(*values.Slot); !ok {
		slot = values.NewVoidSlot()
		return
	}
	slot = result.(*values.Slot)

	// Destruct unused objects
	if stmt.GetKind() != ast.ObjectCreationExpr {
		for index, object := range slices.Backward(env.(*Environment).objects) {
			if object.IsUsed {
				continue
			}
			if err := interpreter.destructObject(object, env.(*Environment)); err != nil {
				interpreter.PrintError(err)
			}
			env.(*Environment).objects = common.RemoveIndex(env.(*Environment).objects, index)
		}
	}

	return
}

type SlotOrError struct {
	Slot  *values.Slot
	Error error
}

func must(slot *values.Slot, err error) *values.Slot {
	if err != nil {
		panic(SlotOrError{Slot: slot, Error: err})
	}
	return slot
}

// Return Void if error is not nil
func mustOrVoid[V any](value V, err error) V {
	if err != nil {
		panic(SlotOrError{Slot: values.NewVoidSlot(), Error: err})
	}
	return value
}
