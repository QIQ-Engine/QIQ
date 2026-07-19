package parser

import (
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/config"
	"QIQ/cmd/qiq/lexer"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/position"
	"fmt"
	"strings"
)

func (parser *Parser) PrintParserCallstack(function string) {
	if !config.ShowParserCallStack {
		return
	}

	if parser != nil {
		fmt.Printf("%s%s (%s)\n", strings.Repeat("  ", parser.stackDepth), function, parser.at().GetPosString())
	} else {
		println(function)
	}

	parser.stackDepth++
}

func (parser *Parser) PopParserCallstack() {
	if parser.stackDepth > 0 {
		parser.stackDepth--
	}
}

func NewExpectedError(expected string, got *lexer.Token) phpError.Error {
	return phpError.NewParseError(`Expected "%s", got "%s" instead in %s`, expected, got.Value, got.GetPosString())
}

// -------------------------------------- Common -------------------------------------- MARK: Common

func (parser *Parser) isEof() bool { return parser.currPos > len(parser.tokens)-1 }

func (parser *Parser) at() *lexer.Token {
	if parser.isEof() {
		lastPos := parser.tokens[parser.currPos-1].Position
		eofPos := position.NewPosition(lastPos.File, lastPos.Line, lastPos.Column+1)
		return lexer.NewToken(lexer.EndOfFileToken, "EOF", eofPos)
	}

	return parser.tokens[parser.currPos]
}

func (parser *Parser) next(offset int) *lexer.Token {
	pos := parser.currPos + offset + 1
	if parser.isEof() || pos >= len(parser.tokens) {
		return lexer.NewToken(lexer.EndOfFileToken, "", nil)
	}

	return parser.tokens[pos]
}

func (parser *Parser) eat() *lexer.Token {
	if parser.isEof() {
		return lexer.NewToken(lexer.EndOfFileToken, "", nil)
	}

	result := parser.at()
	parser.currPos++
	return result
}

func (parser *Parser) eatN(length int) *lexer.Token {
	var result *lexer.Token
	for range length {
		result = parser.eat()
	}
	return result
}

func (parser *Parser) isTokenType(tokenType lexer.TokenType, eat bool) bool {
	result := parser.at().TokenType == tokenType
	if result && eat {
		parser.eat()
	}
	return result
}

func (parser *Parser) isToken(tokenType lexer.TokenType, value string, eat bool) bool {
	parserValue := parser.at().Value
	if common.IsKeyword(parserValue) {
		parserValue = strings.ToLower(parserValue)
	}
	// TODO Find a way to reduce "is..constant" to just one time
	if common.IsCorePredefinedConstant(parserValue) || common.IsContextDependentConstant(parserValue) {
		parserValue = strings.ToUpper(parserValue)
	}

	result := parser.at().TokenType == tokenType && parserValue == value
	if result && eat {
		parser.eat()
	}
	return result
}

func (parser *Parser) expectTokenType(tokenType lexer.TokenType, eat bool) phpError.Error {
	if parser.isTokenType(tokenType, eat) {
		return nil
	}
	return phpError.NewParseError("Unexpected token %s. Expected: %s", parser.at().TokenTypeString(), lexer.TokenTypeToString(tokenType))
}

func (parser *Parser) expect(tokenType lexer.TokenType, value string, eat bool) phpError.Error {
	if parser.isToken(tokenType, value, eat) {
		return nil
	}
	return phpError.NewParseError("Unexpected token %s. Expected: %s", parser.at(), lexer.NewToken(tokenType, value, nil))
}

// -------------------------------------- Tokens -------------------------------------- MARK: Tokens

func (parser *Parser) isTextExpression(eat bool) bool {
	isTextExpression :=
		// ; ?>Text
		(parser.isToken(lexer.OpOrPuncToken, ";", false) && parser.next(0).TokenType == lexer.EndTagToken &&
			parser.next(1).TokenType == lexer.TextToken) || (
		// ?>Text
		parser.isTokenType(lexer.EndTagToken, false) && parser.next(0).TokenType == lexer.TextToken) || (
		// ; ?><?php
		parser.isToken(lexer.OpOrPuncToken, ";", false) && parser.next(0).TokenType == lexer.EndTagToken &&
			parser.next(1).TokenType == lexer.StartTagToken) || (
		// ?><?php
		parser.isTokenType(lexer.EndTagToken, false) && parser.next(0).TokenType == lexer.StartTagToken)

	if isTextExpression && eat {
		parser.isToken(lexer.OpOrPuncToken, ";", true)
		parser.eat()
	}

	return isTextExpression
}

func (parser *Parser) isPhpType(token *lexer.Token) bool {
	return token.TokenType == lexer.OpOrPuncToken && token.Value == "?" ||
		token.TokenType == lexer.KeywordToken && common.IsReturnTypeKeyword(token.Value)
}

func (parser *Parser) getTypes(eat bool) ([]string, phpError.Error) {
	types, _, err := parser.getTypesWithOffset(eat, -1)
	return types, err
}

func (parser *Parser) getTypesWithOffset(eat bool, offset int) ([]string, int, phpError.Error) {
	types := []string{}

	token := func() *lexer.Token {
		return parser.next(offset)
	}

	if token().TokenType == lexer.OpOrPuncToken && token().Value == "?" {
		types = append(types, "null")
		offset++
	}

	if !parser.isPhpType(token()) {
		types = append(types, token().Value)
		offset++
	}

	for parser.isPhpType(token()) {
		types = append(types, strings.ToLower(token().Value))
		offset++

		if token().TokenType == lexer.OpOrPuncToken && token().Value == "|" {
			offset++
			continue
		}

		break
	}

	if eat {
		parser.eatN(offset + 1)
	}

	if len(types) == 0 {
		types = append(types, "mixed")
	}

	return types, offset, nil
}

func (parser *Parser) getQualifiedName(eat bool) (string, phpError.Error) {
	offset := 0
	var name strings.Builder

	token := func() *lexer.Token {
		return parser.next(offset - 1)
	}

	nextMustBeSeparator := false
	for {
		if token().TokenType == lexer.OpOrPuncToken {
			if token().Value == `\` {
				name.WriteString(token().Value)
				offset++
				nextMustBeSeparator = false
			}
		}

		if !nextMustBeSeparator && (token().TokenType == lexer.NameToken || token().TokenType == lexer.KeywordToken) {
			name.WriteString(token().Value)
			offset++
			nextMustBeSeparator = true
			continue
		}

		break
	}

	if eat {
		parser.eatN(offset)
	}

	return name.String(), nil
}
