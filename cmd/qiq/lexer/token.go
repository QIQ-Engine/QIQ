package lexer

import (
	"QIQ/cmd/qiq/position"
	"fmt"
	"slices"
)

type Token struct {
	TokenType TokenType
	Value     string
	Position  *position.Position
}

func NewToken(tokenType TokenType, value string, position *position.Position) *Token {
	return &Token{TokenType: tokenType, Value: value, Position: position}
}

func (token *Token) TokenTypeString() string {
	return TokenTypeToString(token.TokenType)
}

func (token *Token) String() string {
	return fmt.Sprintf(`&{Token - type: %s, value: "%s", position: %s}`, token.TokenTypeString(), token.Value, token.Position)
}

func (token *Token) GetPosition() *position.Position {
	if token.Position == nil {
		return &position.Position{}
	}
	return token.Position
}

func (token *Token) GetPosString() string { return token.GetPosition().ToPosString() }

type TokenType uint8

const (
	EndOfFileToken TokenType = iota
	// Spec: https://phplang.org/spec/04-basic-concepts.html#grammar-start-tag
	TextToken
	StartTagToken
	EndTagToken
	// Spec: https://phplang.org/spec/09-lexical-structure.html#general-1
	VariableNameToken
	NameToken
	KeywordToken
	IntegerLiteralToken
	FloatingLiteralToken
	StringLiteralToken
	OpOrPuncToken
)

func TokenTypeToString(tokenType TokenType) string {
	switch tokenType {
	case EndOfFileToken:
		return "EndOfFileToken"
	// Spec: https://phplang.org/spec/04-basic-concepts.html#grammar-start-tag
	case TextToken:
		return "TextToken"
	case StartTagToken:
		return "StartTagToken"
	case EndTagToken:
		return "EndTagToken"
	// Spec: https://phplang.org/spec/09-lexical-structure.html#general-1
	case VariableNameToken:
		return "VariableNameToken"
	case NameToken:
		return "NameToken"
	case KeywordToken:
		return "KeywordToken"
	case IntegerLiteralToken:
		return "IntegerLiteralToken"
	case FloatingLiteralToken:
		return "FloatingLiteralToken"
	case StringLiteralToken:
		return "StringLiteralToken"
	case OpOrPuncToken:
		return "OpOrPuncToken"
	default:
		return "Unknown token type"
	}
}

func IsLiteral(token *Token) bool {
	return slices.Contains([]TokenType{IntegerLiteralToken, FloatingLiteralToken, StringLiteralToken}, token.TokenType)
}
