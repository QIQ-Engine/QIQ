package ast

import (
	"QIQ/cmd/qiq/position"
	"slices"
	"strings"
)

type NodeId uint32

// -------------------------------------- Statement -------------------------------------- MARK: Statement

type IStatement interface {
	GetId() NodeId
	GetKind() NodeType
	GetKindString() string
	GetPosition() *position.Position
	GetPosString() string
	Process(visitor Visitor, context any) (any, error)
}

type Statement struct {
	id   NodeId
	kind NodeType
	pos  *position.Position
}

func NewEmptyStmt() *Statement { return &Statement{kind: EmptyNode} }

func NewStmt(id NodeId, kind NodeType, pos *position.Position) *Statement {
	return &Statement{id: id, kind: kind, pos: pos}
}

func (stmt *Statement) GetId() NodeId { return stmt.id }

func (stmt *Statement) GetKind() NodeType { return stmt.kind }

func (stmt *Statement) GetKindString() string { return NodeTypeToString(stmt.kind) }

func (stmt *Statement) GetPosition() *position.Position {
	if stmt.pos == nil {
		return &position.Position{}
	}
	return stmt.pos
}

func (stmt *Statement) GetPosString() string { return stmt.GetPosition().ToPosString() }

func (stmt *Statement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessStmt(stmt, context)
}

// -------------------------------------- MethodDefinitionStatement -------------------------------------- MARK: MethodDefinitionStatement

type MethodDefinitionStatement struct {
	*Statement
	Modifiers  []string
	Name       string
	Params     []FunctionParameter
	Body       *CompoundStatement
	ReturnType []string
	Class      *ClassDeclarationStatement
}

func NewMethodDefinitionStmt(id NodeId, pos *position.Position, name string, modifiers []string, params []FunctionParameter, body *CompoundStatement, returnType []string) *MethodDefinitionStatement {
	return &MethodDefinitionStatement{Statement: NewStmt(id, MethodDefinitionStmt, pos),
		Name:       name,
		Modifiers:  modifiers,
		Params:     params,
		Body:       body,
		ReturnType: returnType,
	}
}

func (stmt *MethodDefinitionStatement) Process(visitor Visitor, context any) (any, error) {
	panic("MethodDefinitionStatement.Process should not be called")
}

func (stmt *MethodDefinitionStatement) GetRequiredParamLen() int {
	for i, param := range stmt.Params {
		if param.DefaultValue != nil {
			return i + 1
		}
	}
	return len(stmt.Params)
}

func (stmt *MethodDefinitionStatement) IsStatic() bool {
	return slices.Contains(stmt.Modifiers, "static")
}

// -------------------------------------- FunctionDefinitionStatement -------------------------------------- MARK: FunctionDefinitionStatement

type FunctionParameter struct {
	Type         []string
	Name         string
	ByRef        bool
	DefaultValue IExpression
}

func NewFunctionParam(byRef bool, name string, paramType []string, defaultValue IExpression) FunctionParameter {
	return FunctionParameter{ByRef: byRef, Name: name, Type: paramType, DefaultValue: defaultValue}
}

type FunctionDefinitionStatement struct {
	*Statement
	FunctionName string
	Params       []FunctionParameter
	Body         *CompoundStatement
	ReturnType   []string
}

func NewFunctionDefinitionStmt(id NodeId, pos *position.Position, functionName string, params []FunctionParameter, body *CompoundStatement, returnType []string) *FunctionDefinitionStatement {
	return &FunctionDefinitionStatement{Statement: NewStmt(id, FunctionDefinitionStmt, pos),
		FunctionName: functionName, Params: params, Body: body, ReturnType: returnType,
	}
}

func (stmt *FunctionDefinitionStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessFunctionDefinitionStmt(stmt, context)
}

// -------------------------------------- ForStatement -------------------------------------- MARK: ForStatement

type ForStatement struct {
	*Statement
	Initializer *CompoundStatement
	Control     *CompoundStatement
	EndOfLoop   *CompoundStatement
	Block       IStatement
}

func NewForStmt(id NodeId, pos *position.Position, initializer *CompoundStatement, control *CompoundStatement, endOfLoop *CompoundStatement, block IStatement) *ForStatement {
	return &ForStatement{Statement: NewStmt(id, ForStmt, pos), Initializer: initializer, Control: control, EndOfLoop: endOfLoop, Block: block}
}

func (stmt *ForStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessForStmt(stmt, context)
}

// -------------------------------------- WhileStatement -------------------------------------- MARK: WhileStatement

type WhileStatement struct {
	*Statement
	Condition IExpression
	Block     IStatement
}

func NewWhileStmt(id NodeId, pos *position.Position, condition IExpression, block IStatement) *WhileStatement {
	return &WhileStatement{Statement: NewStmt(id, WhileStmt, pos), Condition: condition, Block: block}
}

func (stmt *WhileStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessWhileStmt(stmt, context)
}

// -------------------------------------- DoStatement -------------------------------------- MARK: DoStatement

type DoStatement struct {
	*WhileStatement
}

func NewDoStmt(id NodeId, pos *position.Position, condition IExpression, block IStatement) *DoStatement {
	return &DoStatement{&WhileStatement{Statement: NewStmt(id, DoStmt, pos), Condition: condition, Block: block}}
}

func (stmt *DoStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessDoStmt(stmt, context)
}

// -------------------------------------- IfStatement -------------------------------------- MARK: IfStatement

type IfStatement struct {
	*Statement
	Condition IExpression
	IfBlock   IStatement
	ElseIf    []*IfStatement
	ElseBlock IStatement
}

func NewIfStmt(id NodeId, pos *position.Position, condition IExpression, ifBlock IStatement, elseIf []*IfStatement, elseBlock IStatement) *IfStatement {
	return &IfStatement{Statement: NewStmt(id, IfStmt, pos), Condition: condition, IfBlock: ifBlock, ElseIf: elseIf, ElseBlock: elseBlock}
}

func (stmt *IfStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessIfStmt(stmt, context)
}

// -------------------------------------- CompoundStatement -------------------------------------- MARK: CompoundStatement

type CompoundStatement struct {
	*Statement
	Statements []IStatement
}

func NewCompoundStmt(id NodeId, statements []IStatement) *CompoundStatement {
	return &CompoundStatement{Statement: NewStmt(id, CompoundStmt, nil), Statements: statements}
}

func (stmt *CompoundStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessCompoundStmt(stmt, context)
}

// -------------------------------------- EchoStatement -------------------------------------- MARK: EchoStatement

type EchoStatement struct {
	*Statement
	Expressions []IExpression
}

func NewEchoStmt(id NodeId, pos *position.Position, expressions []IExpression) *EchoStatement {
	return &EchoStatement{Statement: NewStmt(id, EchoStmt, pos), Expressions: expressions}
}

func (stmt *EchoStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessEchoStmt(stmt, context)
}

// -------------------------------------- ClassConstDeclarationStatement -------------------------------------- MARK: ClassConstDeclarationStatement

type ClassConstDeclarationStatement struct {
	*Statement
	Name      string
	Value     IExpression
	Visiblity string
}

func NewClassConstDeclarationStmt(id NodeId, pos *position.Position, name string, value IExpression, visibility string) *ClassConstDeclarationStatement {
	return &ClassConstDeclarationStatement{Statement: NewStmt(id, ClassConstDeclarationStmt, pos), Name: name, Value: value, Visiblity: visibility}
}

func (stmt *ClassConstDeclarationStatement) Process(visitor Visitor, context any) (any, error) {
	panic("ClassConstDeclarationStatement.Process should not be called")
}

// -------------------------------------- TraitUseStatement -------------------------------------- MARK: TraitUseStatement

type TraitUseStatement struct {
	*Statement
	Name string
}

func NewTraitUseStmt(id NodeId, pos *position.Position, name string) *TraitUseStatement {
	return &TraitUseStatement{Statement: NewStmt(id, ClassConstDeclarationStmt, pos), Name: name}
}

func (stmt *TraitUseStatement) Process(visitor Visitor, context any) (any, error) {
	panic("TraitUseStatement.Process should not be called")
}

// -------------------------------------- ConstDeclarationStatement -------------------------------------- MARK: ConstDeclarationStatement

type ConstDeclarationStatement struct {
	*Statement
	Name  string
	Value IExpression
}

func NewConstDeclarationStmt(id NodeId, pos *position.Position, name string, value IExpression) *ConstDeclarationStatement {
	return &ConstDeclarationStatement{Statement: NewStmt(id, ConstDeclarationStmt, pos), Name: name, Value: value}
}

func (stmt *ConstDeclarationStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessConstDeclarationStmt(stmt, context)
}

// -------------------------------------- ExpressionStatement -------------------------------------- MARK: ExpressionStatement

type ExpressionStatement struct {
	*Statement
	Expr IExpression
}

func NewExpressionStmt(id NodeId, expr IExpression) *ExpressionStatement {
	return &ExpressionStatement{Statement: NewStmt(id, ExpressionStmt, expr.GetPosition()), Expr: expr}
}

func (stmt *ExpressionStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessExpressionStmt(stmt, context)
}

// -------------------------------------- BreakStatement -------------------------------------- MARK: BreakStatement

type BreakStatement struct {
	*ExpressionStatement
}

func NewBreakStmt(id NodeId, pos *position.Position, expr IExpression) *BreakStatement {
	return &BreakStatement{&ExpressionStatement{Statement: NewStmt(id, BreakStmt, pos), Expr: expr}}
}

func (stmt *BreakStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessBreakStmt(stmt, context)
}

// -------------------------------------- ContinueStatement -------------------------------------- MARK: ContinueStatement

type ContinueStatement struct {
	*ExpressionStatement
}

func NewContinueStmt(id NodeId, pos *position.Position, expr IExpression) *ContinueStatement {
	return &ContinueStatement{&ExpressionStatement{Statement: NewStmt(id, ContinueStmt, pos), Expr: expr}}
}

func (stmt *ContinueStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessContinueStmt(stmt, context)
}

// -------------------------------------- ReturnStatement -------------------------------------- MARK: ReturnStatement

type ReturnStatement struct {
	*ExpressionStatement
}

func NewReturnStmt(id NodeId, pos *position.Position, expr IExpression) *ReturnStatement {
	return &ReturnStatement{&ExpressionStatement{Statement: NewStmt(id, ReturnStmt, pos), Expr: expr}}
}

func (stmt *ReturnStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessReturnStmt(stmt, context)
}

// -------------------------------------- GlobalDeclarationStatement -------------------------------------- MARK: GlobalDeclarationStatement

type GlobalDeclarationStatement struct {
	*Statement
	Variables []IExpression
}

func NewGlobalDeclarationStmt(id NodeId, pos *position.Position, variables []IExpression) *GlobalDeclarationStatement {
	return &GlobalDeclarationStatement{Statement: NewStmt(id, GlobalDeclarationStmt, pos), Variables: variables}
}

func (stmt *GlobalDeclarationStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessGlobalDeclarationStmt(stmt, context)
}

// -------------------------------------- PropertyDeclarationStatement -------------------------------------- MARK: PropertyDeclarationStatement

type PropertyDeclarationStatement struct {
	*Statement
	Visibility   string
	IsStatic     bool
	Name         string
	Type         []string
	InitialValue IExpression
}

func NewPropertyDeclarationStmt(id NodeId, pos *position.Position, name, visibility string, isStatic bool, pType []string, initialValue IExpression) *PropertyDeclarationStatement {
	return &PropertyDeclarationStatement{
		Statement:    NewStmt(id, PropertyDeclarationStmt, pos),
		Name:         name,
		Visibility:   visibility,
		IsStatic:     isStatic,
		Type:         pType,
		InitialValue: initialValue,
	}
}

func (stmt *PropertyDeclarationStatement) Process(visitor Visitor, context any) (any, error) {
	panic("PropertyDeclarationStatement.Process should not be called")
}

// -------------------------------------- ClassDeclarationStatement -------------------------------------- MARK: ClassDeclarationStatement

type ClassDeclarationStatement struct {
	*Statement
	IsAbstract     bool
	IsFinal        bool
	Name           string
	BaseClass      string
	Interfaces     []string
	Constants      map[string]*ClassConstDeclarationStatement
	MethodNames    []string
	Methods        map[string]*MethodDefinitionStatement
	PropertieNames []string
	Properties     map[string]*PropertyDeclarationStatement
	Traits         []*TraitUseStatement
}

func NewClassDeclarationStmt(id NodeId, pos *position.Position, name string, isAbstract, isFinal bool) *ClassDeclarationStatement {
	return &ClassDeclarationStatement{
		Statement:      NewStmt(id, ClassDeclarationStmt, pos),
		Name:           name,
		IsAbstract:     isAbstract,
		IsFinal:        isFinal,
		Interfaces:     []string{},
		Constants:      map[string]*ClassConstDeclarationStatement{},
		MethodNames:    []string{},
		Methods:        map[string]*MethodDefinitionStatement{},
		PropertieNames: []string{},
		Properties:     map[string]*PropertyDeclarationStatement{},
		Traits:         []*TraitUseStatement{},
	}
}

func (stmt *ClassDeclarationStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessClassDeclarationStmt(stmt, context)
}

func (stmt *ClassDeclarationStatement) GetQualifiedName() string {
	if stmt.pos == nil || stmt.pos.File == nil {
		return stmt.Name
	}
	return stmt.pos.File.GetNamespaceStr() + stmt.Name
}

func (stmt *ClassDeclarationStatement) AddConst(constStmt *ClassConstDeclarationStatement) {
	stmt.Constants[constStmt.Name] = constStmt
}

func (stmt *ClassDeclarationStatement) GetConst(name string) (*ClassConstDeclarationStatement, bool) {
	classDecl, found := stmt.Constants[name]
	return classDecl, found
}

func (stmt *ClassDeclarationStatement) AddMethod(method *MethodDefinitionStatement) {
	method.Class = stmt
	if !slices.Contains(stmt.MethodNames, method.Name) {
		stmt.MethodNames = append(stmt.MethodNames, method.Name)
	}
	stmt.Methods[strings.ToLower(method.Name)] = method
}

func (stmt *ClassDeclarationStatement) GetMethod(name string) (*MethodDefinitionStatement, bool) {
	methodDecl, found := stmt.Methods[strings.ToLower(name)]
	return methodDecl, found
}

func (stmt *ClassDeclarationStatement) AddProperty(property *PropertyDeclarationStatement) {
	if !slices.Contains(stmt.PropertieNames, property.Name) {
		stmt.PropertieNames = append(stmt.PropertieNames, property.Name)
	}
	stmt.Properties[property.Name] = property
}

func (stmt *ClassDeclarationStatement) AddTrait(trait *TraitUseStatement) {
	stmt.Traits = append(stmt.Traits, trait)
}

// -------------------------------------- ThrowStatement -------------------------------------- MARK: ThrowStatement

type ThrowStatement struct {
	*Statement
	Expr IExpression
}

func NewThrowStmt(id NodeId, pos *position.Position, expr IExpression) *ThrowStatement {
	return &ThrowStatement{Statement: NewStmt(id, ThrowStmt, pos), Expr: expr}
}

func (stmt *ThrowStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessThrowStmt(stmt, context)
}

// -------------------------------------- DeclareStatement -------------------------------------- MARK: DeclareStatement

type DeclareStatement struct {
	*Statement
	Directive string
	Literal   IExpression
}

func NewDeclareStmt(id NodeId, pos *position.Position, directive string, literal IExpression) *DeclareStatement {
	return &DeclareStatement{Statement: NewStmt(id, DeclareStmt, pos), Directive: directive, Literal: literal}
}

func (stmt *DeclareStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessDeclareStmt(stmt, context)
}

// -------------------------------------- ForeachStatement -------------------------------------- MARK: ForeachStatement

type ForeachStatement struct {
	*Statement
	Collection IExpression
	Key        IExpression
	Value      IExpression
	ByRef      bool
	Block      IStatement
}

func NewForeachStmt(id NodeId, pos *position.Position, collection, key, value IExpression, byRef bool, block IStatement) *ForeachStatement {
	return &ForeachStatement{
		Statement:  NewStmt(id, ForeachStmt, pos),
		Collection: collection,
		Key:        key,
		Value:      value,
		ByRef:      byRef,
		Block:      block,
	}
}

func (stmt *ForeachStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessForeachStmt(stmt, context)
}

// -------------------------------------- InterfaceDeclarationStatement -------------------------------------- MARK: InterfaceDeclarationStatement

type InterfaceDeclarationStatement struct {
	*Statement
	Name        string
	Parents     []string
	Constants   map[string]*ClassConstDeclarationStatement
	MethodNames []string
	Methods     map[string]*MethodDefinitionStatement
}

func NewInterfaceDeclarationStmt(id NodeId, pos *position.Position, name string) *InterfaceDeclarationStatement {
	return &InterfaceDeclarationStatement{
		Statement:   NewStmt(id, InterfaceDeclarationStmt, pos),
		Name:        name,
		Parents:     []string{},
		Constants:   map[string]*ClassConstDeclarationStatement{},
		MethodNames: []string{},
		Methods:     map[string]*MethodDefinitionStatement{},
	}
}

func (stmt *InterfaceDeclarationStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessInterfaceDeclarationStmt(stmt, context)
}

func (stmt *InterfaceDeclarationStatement) GetQualifiedName() string {
	if stmt.pos == nil || stmt.pos.File == nil {
		return stmt.Name
	}
	return stmt.pos.File.GetNamespaceStr() + stmt.Name
}

func (stmt *InterfaceDeclarationStatement) AddConst(constStmt *ClassConstDeclarationStatement) {
	stmt.Constants[constStmt.Name] = constStmt
}

func (stmt *InterfaceDeclarationStatement) GetConst(name string) (*ClassConstDeclarationStatement, bool) {
	classDecl, found := stmt.Constants[name]
	return classDecl, found
}

func (stmt *InterfaceDeclarationStatement) AddMethod(method *MethodDefinitionStatement) {
	if !slices.Contains(stmt.MethodNames, method.Name) {
		stmt.MethodNames = append(stmt.MethodNames, method.Name)
	}
	stmt.Methods[strings.ToLower(method.Name)] = method
}

func (stmt *InterfaceDeclarationStatement) GetMethod(name string) (*MethodDefinitionStatement, bool) {
	methodDef, found := stmt.Methods[strings.ToLower(name)]
	if !found {
		return nil, false
	}
	return methodDef, true
}

// -------------------------------------- TryStatement -------------------------------------- MARK: TryStatement

type CatchStatement struct {
	ErrorType    []string
	VariableName string
	Body         *CompoundStatement
}

type TryStatement struct {
	*Statement
	Body    *CompoundStatement
	Catches []CatchStatement
	Finally *CompoundStatement
}

func NewTryStmt(id NodeId, pos *position.Position, body *CompoundStatement) *TryStatement {
	return &TryStatement{
		Statement: NewStmt(id, TryStmt, pos),
		Body:      body,
		Catches:   []CatchStatement{},
	}
}

func (stmt *TryStatement) Process(visitor Visitor, context any) (any, error) {
	return visitor.ProcessTryStmt(stmt, context)
}

func (stmt *TryStatement) AddCatch(catch CatchStatement) {
	stmt.Catches = append(stmt.Catches, catch)
}
