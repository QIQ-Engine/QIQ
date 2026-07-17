package runtime

import (
	"QIQ/cmd/qiq/ast"
	"QIQ/cmd/qiq/ini"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/request"
	"QIQ/cmd/qiq/runtime/outputBuffer"
	"QIQ/cmd/qiq/runtime/values"
)

type Interpreter interface {
	ProcessStatement(stmt ast.IStatement, env any) (*values.Slot, phpError.Error)
	// Context
	GetIni() *ini.Ini
	GetExectionContext() *ExecutionContext
	GetFilename() string
	GetWorkingDir() string
	// Request
	GetRequest() *request.Request
	GetResponse() *request.Response
	// Class declarations
	AddClass(class string, classDecl *ast.ClassDeclarationStatement) phpError.Error
	GetClass(class string) (*ast.ClassDeclarationStatement, bool)
	GetClasses() []string
	// Interface declarations
	AddInterface(interfaceName string, interfaceDecl *ast.InterfaceDeclarationStatement) phpError.Error
	GetInterface(interfaceName string) (*ast.InterfaceDeclarationStatement, bool)
	GetInterfaces() []string
	// Output
	GetOutputBufferStack() *outputBuffer.Stack
	Print(str string)
	Println(str string)
	PrintError(err phpError.Error)
	WriteResult(str string)
}
