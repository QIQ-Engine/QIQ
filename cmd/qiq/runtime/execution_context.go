package runtime

import (
	"QIQ/cmd/qiq/ast"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/position"
	"QIQ/cmd/qiq/runtime/values"
	"strings"
)

type ExecutionContext struct {
	// State
	initializationCompleted bool
	// Classes
	classNames        []string
	classDeclarations map[string]*ast.ClassDeclarationStatement
	// Interfaces
	interfaceNames        []string
	interfaceDeclarations map[string]*ast.InterfaceDeclarationStatement
	// Objects
	objects map[string][]*values.Object
}

func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		// State
		initializationCompleted: false,
		// Classes
		classNames:        []string{},
		classDeclarations: map[string]*ast.ClassDeclarationStatement{},
		// Interfaces
		interfaceNames:        []string{},
		interfaceDeclarations: map[string]*ast.InterfaceDeclarationStatement{},
		// Objects
		objects: map[string][]*values.Object{},
	}
}

// -------------------------------------- State -------------------------------------- MARK: State

func (executionContext *ExecutionContext) InitializationCompleted() {
	executionContext.initializationCompleted = true
}

func (executionContext *ExecutionContext) IsInitializationCompleted() bool {
	return executionContext.initializationCompleted
}

// -------------------------------------- Classes & Interfaces -------------------------------------- MARK: Classes & Interfaces

func (executionContext *ExecutionContext) HasClassOrInterface(classOrInterface string, pos *position.Position) phpError.Error {
	interfaceDecl, found := executionContext.GetInterface(classOrInterface)
	if found {
		if interfaceDecl.GetPosition().File == nil {
			return phpError.NewError(
				"Cannot redeclare interface %s in %s",
				interfaceDecl.GetQualifiedName(), pos.ToPosString(),
			)
		}

		return phpError.NewError(
			"Cannot redeclare interface %s (previously declared in %s) in %s",
			interfaceDecl.GetQualifiedName(), interfaceDecl.GetPosString(), pos.ToPosString(),
		)
	}

	classDecl, found := executionContext.GetClass(classOrInterface)
	if found {
		if classDecl.GetPosition().File == nil {
			return phpError.NewError(
				"Cannot redeclare class %s in %s",
				classDecl.GetQualifiedName(), pos.ToPosString(),
			)
		}

		return phpError.NewError(
			"Cannot redeclare class %s (previously declared in %s) in %s",
			classDecl.GetQualifiedName(), classDecl.GetPosString(), pos.ToPosString(),
		)
	}

	return nil
}

// -------------------------------------- Classes -------------------------------------- MARK: Classes

func (executionContext *ExecutionContext) AddClass(class string, classDecl *ast.ClassDeclarationStatement) phpError.Error {
	if err := executionContext.HasClassOrInterface(class, classDecl.GetPosition()); err != nil {
		return err
	}

	executionContext.classNames = append(executionContext.classNames, class)
	executionContext.classDeclarations[strings.ToLower(class)] = classDecl

	return nil
}

func (executionContext *ExecutionContext) GetClass(class string) (*ast.ClassDeclarationStatement, bool) {
	classDeclaration, found := executionContext.classDeclarations[strings.ToLower(class)]
	if !found {
		return nil, false
	}
	return classDeclaration, true
}

func (executionContext *ExecutionContext) GetClasses() []string { return executionContext.classNames }

// -------------------------------------- Interfaces -------------------------------------- MARK: Interfaces

func (executionContext *ExecutionContext) AddInterface(interfaceName string, interfaceDecl *ast.InterfaceDeclarationStatement) phpError.Error {
	if err := executionContext.HasClassOrInterface(interfaceName, interfaceDecl.GetPosition()); err != nil {
		return err
	}

	executionContext.interfaceNames = append(executionContext.interfaceNames, interfaceName)
	executionContext.interfaceDeclarations[strings.ToLower(interfaceName)] = interfaceDecl

	return nil
}

func (executionContext *ExecutionContext) GetInterface(interfaceName string) (*ast.InterfaceDeclarationStatement, bool) {
	interfaceDecl, found := executionContext.interfaceDeclarations[strings.ToLower(interfaceName)]
	if !found {
		return nil, false
	}
	return interfaceDecl, true
}

func (executionContext *ExecutionContext) GetInterfaces() []string {
	return executionContext.interfaceNames
}

// -------------------------------------- Objects -------------------------------------- MARK: Objects

func (executionContext *ExecutionContext) AddObject(className string, object *values.Object) {
	_, found := executionContext.objects[className]
	if !found {
		executionContext.objects[className] = []*values.Object{object}
		return
	}
	executionContext.objects[className] = append(executionContext.objects[className], object)
}

func (executionContext *ExecutionContext) CountObjects(className string) int {
	_, found := executionContext.objects[className]
	if !found {
		return 0
	}
	return len(executionContext.objects[className])
}
