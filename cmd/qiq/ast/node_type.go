package ast

type NodeType uint8

const (
	EmptyNode NodeType = iota
	ProgramNode
	TextNode
	// Expressions
	AnonymousFunctionCreationExpr
	ArrayLiteralExpr
	ArrayNextKeyExpr
	BinaryOpExpr
	CastExpr
	CoalesceExpr
	CompoundAssignmentExpr
	ConditionalExpr
	ConstantAccessExpr
	EmptyIntrinsicExpr
	EqualityExpr
	ErrorControlExpr
	EvalIntrinsicExpr
	ExitIntrinsicExpr
	FloatingLiteralExpr
	FunctionCallExpr
	IncludeExpr
	IncludeOnceExpr
	IntegerLiteralExpr
	IssetIntrinsicExpr
	LogicalNotExpr
	MemberAccessExpr
	ObjectCreationExpr
	ParenthesizedExpr
	PostfixIncExpr
	PrefixIncExpr
	PrintExpr
	RelationalExpr
	RequireExpr
	RequireOnceExpr
	ShiftExpr
	SimpleAssignmentExpr
	SimpleVariableExpr
	StringLiteralExpr
	SubscriptExpr
	UnaryOpExpr
	UnsetIntrinsicExpr
	VariableNameExpr
	// Statements
	BreakStmt
	CompoundStmt
	ConstDeclarationStmt
	ContinueStmt
	DeclareStmt
	DoStmt
	EchoStmt
	ExpressionStmt
	ForeachStmt
	ForStmt
	FunctionDefinitionStmt
	GlobalDeclarationStmt
	IfStmt
	InterfaceDeclarationStmt
	ReturnStmt
	ThrowStmt
	TraitUseStmt
	TryStmt
	WhileStmt
	// Class
	ClassConstDeclarationStmt
	ClassDeclarationStmt
	MethodDefinitionStmt
	PropertyDeclarationStmt
)

func NodeTypeToString(nodeType NodeType) string {
	switch nodeType {
	case EmptyNode:
		return "EmptyNode"
	case ProgramNode:
		return "ProgramNode"
	case TextNode:
		return "TextNode"
	// Expressions
	case AnonymousFunctionCreationExpr:
		return "AnonymousFunctionCreationExpr"
	case ArrayLiteralExpr:
		return "ArrayLiteralExpr"
	case ArrayNextKeyExpr:
		return "ArrayNextKeyExpr"
	case BinaryOpExpr:
		return "BinaryOpExpr"
	case CastExpr:
		return "CastExpr"
	case CoalesceExpr:
		return "CoalesceExpr"
	case CompoundAssignmentExpr:
		return "CompoundAssignmentExpr"
	case ConditionalExpr:
		return "ConditionalExpr"
	case ConstantAccessExpr:
		return "ConstantAccessExpr"
	case EmptyIntrinsicExpr:
		return "EmptyIntrinsicExpr"
	case EqualityExpr:
		return "EqualityExpr"
	case ErrorControlExpr:
		return "ErrorControlExpr"
	case EvalIntrinsicExpr:
		return "EvalIntrinsicExpr"
	case ExitIntrinsicExpr:
		return "ExitIntrinsicExpr"
	case FloatingLiteralExpr:
		return "FloatingLiteralExpr"
	case FunctionCallExpr:
		return "FunctionCallExpr"
	case IncludeExpr:
		return "IncludeExpr"
	case IncludeOnceExpr:
		return "IncludeOnceExpr"
	case IntegerLiteralExpr:
		return "IntegerLiteralExpr"
	case IssetIntrinsicExpr:
		return "IssetIntrinsicExpr"
	case LogicalNotExpr:
		return "LogicalNotExpr"
	case MemberAccessExpr:
		return "MemberAccessExpr"
	case ObjectCreationExpr:
		return "ObjectCreationExpr"
	case ParenthesizedExpr:
		return "ParenthesizedExpr"
	case PostfixIncExpr:
		return "PostfixIncExpr"
	case PrefixIncExpr:
		return "PrefixIncExpr"
	case PrintExpr:
		return "PrintExpr"
	case RelationalExpr:
		return "RelationalExpr"
	case RequireExpr:
		return "RequireExpr"
	case RequireOnceExpr:
		return "RequireOnceExpr"
	case ShiftExpr:
		return "ShiftExpr"
	case SimpleAssignmentExpr:
		return "SimpleAssignmentExpr"
	case SimpleVariableExpr:
		return "SimpleVariableExpr"
	case StringLiteralExpr:
		return "StringLiteralExpr"
	case SubscriptExpr:
		return "SubscriptExpr"
	case UnaryOpExpr:
		return "UnaryOpExpr"
	case UnsetIntrinsicExpr:
		return "UnsetIntrinsicExpr"
	case VariableNameExpr:
		return "VariableNameExpr"
	// Statements
	case BreakStmt:
		return "BreakStmt"
	case CompoundStmt:
		return "CompoundStmt"
	case ConstDeclarationStmt:
		return "ConstDeclarationStmt"
	case ContinueStmt:
		return "ContinueStmt"
	case DeclareStmt:
		return "DeclareStmt"
	case DoStmt:
		return "DoStmt"
	case EchoStmt:
		return "EchoStmt"
	case ExpressionStmt:
		return "ExpressionStmt"
	case ForeachStmt:
		return "ForeachStmt"
	case ForStmt:
		return "ForStmt"
	case FunctionDefinitionStmt:
		return "FunctionDefinitionStmt"
	case GlobalDeclarationStmt:
		return "GlobalDeclarationStmt"
	case IfStmt:
		return "IfStmt"
	case InterfaceDeclarationStmt:
		return "InterfaceDeclarationStmt"
	case ReturnStmt:
		return "ReturnStmt"
	case ThrowStmt:
		return "ThrowStmt"
	case TraitUseStmt:
		return "TraitUseStmt"
	case TryStmt:
		return "TryStmt"
	case WhileStmt:
		return "WhileStmt"
	// Class
	case ClassConstDeclarationStmt:
		return "ClassConstDeclarationStmt"
	case ClassDeclarationStmt:
		return "ClassDeclarationStmt"
	case MethodDefinitionStmt:
		return "MethodDefinitionStmt"
	case PropertyDeclarationStmt:
		return "PropertyDeclarationStmt"
	default:
		return "Unknown node type"
	}
}
