package lexer

import (
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/ini"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/position"
	"QIQ/cmd/qiq/stats"
	"fmt"
	"slices"
	"strings"
)

type Lexer struct {
	ini    *ini.Ini
	input  string
	tokens []*Token
	// Position info
	file              *position.File
	currPos           PositionSnapshot
	positionSnapShots []PositionSnapshot
}

type PositionSnapshot struct {
	CurrPos          int
	CurrLine         int
	CurrCol          int
	CurrTokenLine    int
	CurrTokenCol     int
	SearchTokenStart bool
}

func NewLexer(ini *ini.Ini) *Lexer {
	return &Lexer{ini: ini}
}

func (lexer *Lexer) init(input string, filename string) {
	lexer.input = input
	lexer.tokens = make([]*Token, 0)
	// Position info
	lexer.file = position.NewFile(filename)
	lexer.currPos = PositionSnapshot{
		CurrPos: 0, CurrLine: 1, CurrCol: 1, CurrTokenLine: 1, CurrTokenCol: 1, SearchTokenStart: false,
	}
	lexer.positionSnapShots = []PositionSnapshot{}
}

func (lexer *Lexer) Tokenize(sourceCode string, filename string) ([]*Token, error) {
	stat := stats.Start()
	defer stats.StopAndPrint(stat, "Lexer")

	lexer.init(sourceCode, filename)

	err := lexer.tokenizeScript()

	return lexer.tokens, err
}

func (lexer *Lexer) tokenizeScript() error {
	// Spec: https://phplang.org/spec/04-basic-concepts.html

	// script:
	//    script-section
	//    script   script-section

	// script-section:
	//    text(opt)   start-tag   statement-list(opt)   end-tag(opt)   text(opt)

	text := ""
	pushTextToken := func() {
		if text != "" {
			lexer.pushToken(TextToken, text)
			text = ""
		}
	}

	for !lexer.isEof() {
		// Push optional text token if a start-tag is detected
		if lexer.nextN(3) == "<?=" || strings.ToLower(lexer.nextN(5)) == "<?php" ||
			(lexer.ini.GetBool("short_open_tag") && lexer.nextN(2) == "<?") {
			pushTextToken()
		}

		if strings.ToLower(lexer.nextN(5)) == "<?php" {
			lexer.eatN(5)
			lexer.pushToken(StartTagToken, "")
			if err := lexer.tokenizeInputFile(); err != nil {
				return err
			}
			continue
		}

		// Supported statement: short echo statement: `<?= "123";`
		if lexer.nextN(3) == "<?=" {
			lexer.eatN(3)
			lexer.pushToken(StartTagToken, "")
			// Spec: https://phplang.org/spec/04-basic-concepts.html#program-structure
			// If `<?=` is used as the start-tag, the Engine proceeds as if the statement-list started with an echo statement.
			lexer.pushKeywordToken("echo")

			if err := lexer.tokenizeInputFile(); err != nil {
				return err
			}
			continue
		}

		// Supported statement: short open tag: `<? 1 + 2;`
		if lexer.ini.GetBool("short_open_tag") && lexer.nextN(2) == "<?" {
			lexer.eatN(2)
			lexer.pushToken(StartTagToken, "")
			if err := lexer.tokenizeInputFile(); err != nil {
				return err
			}
			continue
		}

		text += lexer.eat()
	}

	// Push optional text token if a start-tag is detected
	pushTextToken()

	return nil
}

func (lexer *Lexer) tokenizeInputFile() error {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-input-file

	// input-file::
	//    input-element
	//    input-file   input-element

	// input-element::
	//    comment
	//    white-space
	//    token

	for !lexer.isEof() {
		// End-Tag
		if lexer.nextN(2) == "?>" {
			lexer.eatN(2)
			if lexer.lastToken().TokenType != OpOrPuncToken || lexer.lastToken().Value != ";" {
				lexer.pushToken(OpOrPuncToken, ";")
			}
			lexer.pushToken(EndTagToken, "")

			// Line breaks directly after closing tags are discarded
			if lexer.at() == "\n" {
				lexer.eat()
			}
			return nil
		}

		// comment
		if lexer.at() == "#" || lexer.nextN(2) == "//" || lexer.nextN(2) == "/*" {
			if err := lexer.tokenizeComment(); err != nil {
				return err
			}
			continue
		}

		// white-space
		if lexer.isWhiteSpace(true) {
			continue
		}

		// token
		if err := lexer.tokenizeToken(); err != nil {
			return err
		}
	}

	return nil
}

func (lexer *Lexer) tokenizeComment() error {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#comments

	// comment::
	//    single-line-comment
	//    delimited-comment

	// single-line-comment::
	//    //   input-characters(opt)
	//    #   input-characters(opt)

	// input-characters::
	//    input-character
	//    input-characters   input-character

	//input-character::
	//    Any source character except   new-line

	// new-line::
	//    Carriage-return character (0x0D)
	//    Line-feed character (0x0A)
	//    Carriage-return character (0x0D) followed by line-feed character (0x0A)

	// delimited-comment::
	//    /*   No characters or any source character sequence except */   */

	isSingleLineComment := true
	if lexer.at() == "#" {
		lexer.eat()
	} else if lexer.nextN(2) == "//" {
		lexer.eatN(2)
	} else if lexer.nextN(2) == "/*" {
		lexer.eatN(2)
		isSingleLineComment = false
	} else {
		return fmt.Errorf("Syntax error: Unexpected start of a comment in %s:%d:%d", lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrCol)
	}

	for !lexer.isEof() {
		if isSingleLineComment {
			// single-line-comment

			// Spec: https://phplang.org/spec/09-lexical-structure.html#comments
			// Except within a string literal or a comment, the characters // or # start a single-line comment, which ends with a
			// new line. That new line is not part of the comment. However, if the single-line comment is the last source element
			// in an embedded script, the trailing new line can be omitted.
			if lexer.isNewLine(false) {
				return nil
			}
		} else {
			// delimited-comment

			// Spec: https://phplang.org/spec/09-lexical-structure.html#comments
			// Except within a string literal or a comment, the characters /* start a delimited comment, which ends with the characters */.
			if lexer.nextN(2) == "*/" {
				lexer.eatN(2)
				return nil
			}
		}
		lexer.eat()
	}

	// Delimited-comment must be closed with */
	if !isSingleLineComment {
		return fmt.Errorf("Unterminated delimited comment detected in %s:%d:%d", lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrCol)
	}

	return nil
}

func (lexer *Lexer) tokenizeToken() error {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#tokens

	// token::
	//    variable-name
	//    name
	//    keyword
	//    integer-literal
	//    floating-literal
	//    string-literal
	//    operator-or-punctuator

	for !lexer.isEof() {
		// string-literal
		if strings.ToLower(lexer.nextN(2)) == `b"` || lexer.at() == `"` ||
			strings.ToLower(lexer.nextN(2)) == "b'" || lexer.at() == "'" ||
			strings.ToLower(lexer.nextN(4)) == "b<<<" || lexer.nextN(3) == "<<<" {
			_, err := lexer.getStringLiteral(false)
			if err != nil {
				return err
			}
			str, _ := lexer.getStringLiteral(true)
			lexer.pushToken(StringLiteralToken, str)
			return nil
		}

		// variable-name
		if lexer.at() == "$" {
			// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-variable-name

			// variable-name::
			//    $   name

			lexer.pushSnapShot()
			lexer.eat()
			if name := lexer.getName(false); name != "" {
				lexer.getName(true)
				lexer.pushToken(VariableNameToken, "$"+name)
				lexer.popSnapShot(false)
				return nil
			}
			lexer.popSnapShot(true)
		}

		// keyword or name
		if common.IsNameNondigit(lexer.at()) {
			if name := lexer.getName(false); name != "" {
				lexer.getName(true)

				// Spec: https://phplang.org/spec/09-lexical-structure.html#keywords
				// Note carefully that yield from is a single token that contains whitespace. However, comments are not permitted in
				// that whitespace.
				lexer.pushSnapShot()
				lexer.isWhiteSpace(true)
				if name == "yield" && strings.ToLower(lexer.nextN(4)) == "from" {
					lexer.eatN(4)
					name += " from"
					lexer.popSnapShot(false)
				} else {
					lexer.popSnapShot(true)
				}

				// keyword

				// Spec: https://phplang.org/spec/09-lexical-structure.html#keywords
				// Also, all magic constants are also treated as keywords.
				if common.IsKeyword(name) || common.IsContextDependentConstant(name) || common.IsCorePredefinedConstant(name) ||
					common.IsCastTypeKeyword(name) {
					lexer.pushKeywordToken(name)
					return nil
				}

				// name
				lexer.pushToken(NameToken, name)
				return nil
			}
			return fmt.Errorf("Unsupported name detected in %s:%d:%d", lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrCol)
		}

		// integer-literal or floating-literal
		if common.IsDigit(lexer.at()) || (lexer.at() == "." && (common.IsDigit(lexer.next(0)) || (strings.ToLower(lexer.next(0)) == "e" && common.IsDigit(lexer.next(1))))) {
			// floating-literal
			if float := lexer.getFloatingPointLiteral(false); float != "" {
				lexer.getFloatingPointLiteral(true)
				lexer.pushToken(FloatingLiteralToken, float)
				return nil
			}

			// integer-literal
			if int := lexer.getIntegerLiteral(false); int != "" {
				lexer.getIntegerLiteral(true)
				lexer.pushToken(IntegerLiteralToken, int)
				return nil
			}

			return fmt.Errorf("Unsupported number format detected in %s:%d:%d", lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrCol)
		}

		// operator-or-punctuator
		if op := lexer.getOperatorOrPunctuator(false); op != "" {
			lexer.getOperatorOrPunctuator(true)
			lexer.pushToken(OpOrPuncToken, op)
			return nil
		}

		return fmt.Errorf("Uncaught char '%s' in %s:%d:%d", lexer.at(), lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrCol)
	}

	return nil
}

func (lexer *Lexer) getName(eat bool) string {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-name

	// name::
	//    name-nondigit
	//    name   name-nondigit
	//    name   digit

	// name-nondigit::
	//    nondigit
	//    one of the characters 0x80–0xff

	// nondigit:: one of
	//    _
	//    a   b   c   d   e   f   g   h   i   j   k   l   m
	//    n   o   p   q   r   s   t   u   v   w   x   y   z
	//    A   B   C   D   E   F   G   H   I   J   K   L   M
	//    N   O   P   Q   R   S   T   U   V   W   X   Y   Z

	name := ""

	lexer.pushSnapShot()

	if !common.IsNameNondigit(lexer.at()) {
		lexer.popSnapShot(false)
		return name
	}
	name += lexer.eat()

	for !lexer.isEof() {
		if !common.IsNameNondigit(lexer.at()) && !common.IsDigit(lexer.at()) {
			break
		}
		name += lexer.eat()
	}

	lexer.popSnapShot(!eat)

	return name
}

func (lexer *Lexer) getIntegerLiteral(eat bool) string {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-integer-literal

	// 	integer-literal::
	//    decimal-literal
	//    octal-literal
	//    hexadecimal-literal
	//    binary-literal

	intStr := ""
	lexer.pushSnapShot()

	// ------------------- binary-literal -------------------

	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-binary-literal

	// binary-literal::
	//    binary-prefix   binary-digit
	//    binary-literal   binary-digit

	// binary-prefix:: one of
	//    0b   0B

	// binary-digit:: one of
	//    0   1

	if strings.ToLower(lexer.nextN(2)) == "0b" {
		intStr += lexer.nextN(2)
		lexer.eatN(2)

		for !lexer.isEof() {
			if lexer.at() != "0" && lexer.at() != "1" && lexer.at() != "_" {
				break
			}
			intStr += lexer.eat()
		}

		if common.IsBinaryLiteral(intStr, false) {
			lexer.popSnapShot(!eat)
			return intStr
		}

		lexer.popSnapShot(true)
		return ""
	}

	// ------------------- hexadecimal-literal -------------------

	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-hexadecimal-literal

	// hexadecimal-literal::
	//    hexadecimal-prefix   hexadecimal-digit
	//    hexadecimal-literal   hexadecimal-digit

	// hexadecimal-prefix:: one of
	//    0x   0X

	if strings.ToLower(lexer.nextN(2)) == "0x" {
		intStr += lexer.nextN(2)
		lexer.eatN(2)

		for !lexer.isEof() {
			if !common.IsHexadecimalDigit(lexer.at()) && lexer.at() != "_" {
				break
			}
			intStr += lexer.eat()
		}

		if common.IsHexadecimalLiteral(intStr, false) {
			lexer.popSnapShot(!eat)
			return intStr
		}

		lexer.popSnapShot(true)
		return ""
	}

	// All other integer cases
	for !lexer.isEof() {
		if !common.IsDigit(lexer.at()) && lexer.at() != "_" {
			break
		}
		intStr += lexer.eat()
	}

	// ------------------- decimal-literal -------------------

	if common.IsDecimalLiteral(intStr, false) {
		lexer.popSnapShot(!eat)
		return intStr
	}

	// ------------------- octal-literal -------------------

	if common.IsOctalLiteral(intStr, false) {
		lexer.popSnapShot(!eat)
		return intStr
	}

	lexer.popSnapShot(true)
	return ""
}

func (lexer *Lexer) getFloatingPointLiteral(eat bool) string {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-floating-literal

	// floating-literal::
	//    fractional-literal   exponent-part(opt)
	//    digit-sequence   exponent-part

	// fractional-literal::
	//    digit-sequence(opt)   .   digit-sequence
	//    digit-sequence   .

	// exponent-part::
	//    e   sign(opt)   digit-sequence
	//    E   sign(opt)   digit-sequence

	// sign:: one of
	//    +   -

	// digit-sequence::
	//    digit
	//    digit-sequence   digit

	var floatStr strings.Builder
	lexer.pushSnapShot()
	state := "beforeDot"

	for !lexer.isEof() {
		if state == "beforeDot" {
			if common.IsDigit(lexer.at()) || lexer.at() == "_" {
				floatStr.WriteString(lexer.eat())
				continue
			}
			if lexer.at() == "." {
				floatStr.WriteString(lexer.eat())

				state = "afterDot"
				continue
			}
			if strings.ToLower(lexer.at()) == "e" {
				floatStr.WriteString(lexer.eat())
				state = "exponent"
				continue
			}
			// Invalid
			break
		}
		if state == "afterDot" {
			if common.IsDigit(lexer.at()) || lexer.at() == "_" {
				floatStr.WriteString(lexer.eat())
				continue
			}
			if strings.ToLower(lexer.at()) == "e" {
				floatStr.WriteString(lexer.eat())
				state = "exponent"
				continue
			}
			// Invalid
			break
		}
		if state == "exponent" {
			if lexer.at() == "+" || lexer.at() == "-" || common.IsDigit(lexer.at()) {
				floatStr.WriteString(lexer.eat())
				state = "exponentDigit"
				continue
			}
			// Invalid
			break
		}
		if state == "exponentDigit" {
			if common.IsDigit(lexer.at()) || lexer.at() == "_" {
				floatStr.WriteString(lexer.eat())
				continue
			}
			break
		}
	}

	lexer.popSnapShot(!eat)

	if common.IsFloatingLiteral(floatStr.String(), false) {
		return floatStr.String()
	}

	lexer.popSnapShot(true)
	return ""
}

func (lexer *Lexer) getStringLiteral(eat bool) (string, phpError.Error) {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-string-literal

	// string-literal::
	//    single-quoted-string-literal
	//    double-quoted-string-literal
	//    heredoc-string-literal
	//    nowdoc-string-literal

	strValue := ""
	lexer.pushSnapShot()

	// ------------------- single-quoted-string-literal -------------------

	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-single-quoted-string-literal

	// single-quoted-string-literal::
	//    b-prefix(opt)   '   sq-char-sequence(opt)   '

	// sq-char-sequence::
	//    sq-char
	//    sq-char-sequence   sq-char

	// sq-char::
	//    sq-escape-sequence
	//    \(opt)   any member of the source character set except single-quote (') or backslash (\)

	// sq-escape-sequence:: one of
	//    \'   \\

	//  b-prefix:: one of
	//    b   B

	// Supported expression: single quoted string: `'Hi World!'`
	if strings.ToLower(lexer.nextN(2)) == "b'" || lexer.at() == "'" {
		if strings.ToLower(lexer.nextN(2)) == "b'" {
			strValue += lexer.nextN(2)
			lexer.eatN(2)
		} else if lexer.at() == "'" {
			strValue += lexer.eat()
		}

		for !lexer.isEof() {
			if lexer.at() == "'" {
				strValue += lexer.eat()
				break
			}
			if lexer.nextN(2) == `\'` || lexer.nextN(2) == `\\` ||
				(lexer.at() == `\` && lexer.next(0) != "'" && lexer.next(0) != `\`) {
				strValue += lexer.nextN(2)
				lexer.eatN(2)
				continue
			}
			if lexer.at() != `\` && lexer.at() != "'" {
				strValue += lexer.eat()
				continue
			}
		}

		if common.IsSingleQuotedStringLiteral(strValue) {
			lexer.popSnapShot(!eat)
			return strValue, nil
		}

		lexer.popSnapShot(true)
		return "", phpError.NewError("Invalid single quoted string literal detected in %s:%d:%d", lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrPos)
	}

	// ------------------- double-quoted-string-literal -------------------

	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-double-quoted-string-literal

	// double-quoted-string-literal::
	//    b-prefix(opt)   "   dq-char-sequence(opt)   "

	//  b-prefix:: one of
	//    b   B

	// dq-char-sequence::
	//    dq-char
	//    dq-char-sequence   dq-char

	// dq-char::
	//    dq-escape-sequence
	//    any member of the source character set except double-quote (") or backslash (\)
	//    \   any member of the source character set except "\$efnrtvxX or   octal-digit

	// dq-escape-sequence::
	//    dq-simple-escape-sequence
	//    dq-octal-escape-sequence
	//    dq-hexadecimal-escape-sequence
	//    dq-unicode-escape-sequence

	// dq-simple-escape-sequence:: one of
	//    \"   \\   \$   \e   \f   \n   \r   \t   \v

	// dq-octal-escape-sequence::
	//    \   octal-digit
	//    \   octal-digit   octal-digit
	//    \   octal-digit   octal-digit   octal-digit

	// octal-digit:: one of
	//    0   1   2   3   4   5   6   7

	// dq-hexadecimal-escape-sequence::
	//    \x   hexadecimal-digit   hexadecimal-digit(opt)
	//    \X   hexadecimal-digit   hexadecimal-digit(opt)

	// dq-unicode-escape-sequence::
	//    \u{   codepoint-digits   }

	// codepoint-digits::
	//    hexadecimal-digit
	//    hexadecimal-digit   codepoint-digits

	// hexadecimal-digit:: one of
	//    0   1   2   3   4   5   6   7   8   9
	//    a   b   c   d   e   f
	//    A   B   C   D   E   F

	// Supported expression: double quoted string: `"Hi $world!"`
	if strings.ToLower(lexer.nextN(2)) == `b"` || lexer.at() == `"` {
		if strings.ToLower(lexer.nextN(2)) == `b"` {
			strValue += lexer.nextN(2)
			lexer.eatN(2)
		} else if lexer.at() == `"` {
			strValue += lexer.eat()
		}

		for !lexer.isEof() {
			if lexer.at() == `"` {
				strValue += lexer.eat()
				break
			}
			if lexer.at() == `\` && lexer.next(0) != "" {
				strValue += lexer.nextN(2)
				lexer.eatN(2)
				continue
			}
			if lexer.at() != `\` {
				strValue += lexer.eat()
				continue
			}
		}

		if common.IsDoubleQuotedStringLiteral(strValue) {
			lexer.popSnapShot(!eat)
			return strValue, nil
		}

		lexer.popSnapShot(true)
		return "", phpError.NewError("Invalid double quoted string literal detected in %s:%d:%d", lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrPos)
	}

	// ------------------- heredoc-string-literal -------------------

	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-heredoc-string-literal

	// heredoc-string-literal::
	//    b-prefix(opt)   <<<   hd-start-identifier   new-line   hd-body(opt)   hd-end-identifier   ;(opt)   new-line

	// hd-start-identifier::
	//    name
	//    "   name   "

	// hd-end-identifier::
	//    name

	// hd-body::
	//    hd-char-sequence(opt)   new-line

	// hd-char-sequence::
	//    hd-char
	//    hd-char-sequence   hd-char

	// hd-char::
	//    hd-escape-sequence
	//    any member of the source character set except backslash (\)
	//    \ any member of the source character set except \$efnrtvxX or   octal-digit

	// hd-escape-sequence::
	//    hd-simple-escape-sequence
	//    dq-octal-escape-sequence
	//    dq-hexadecimal-escape-sequence
	//    dq-unicode-escape-sequence

	// hd-simple-escape-sequence:: one of
	//    \\   \$   \e   \f   \n   \r   \t   \v

	// Supported expression: heredoc string: `"<<<EOF\nHi $world!\nEOF;"`
	if strings.ToLower(lexer.nextN(4)) == `b<<<` || lexer.nextN(3) == "<<<" {
		// Opening symbol
		if lexer.nextN(3) == "<<<" {
			strValue += lexer.eatN(3)
		} else {
			strValue += lexer.eatN(4)
		}

		// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-heredoc-string-literal
		// Only horizontal white space is permitted between <<< and the start identifier.
		for lexer.isWhiteSpaceChar(lexer.at()) {
			lexer.eat()
		}

		// Check if hd-start-identifier is in quotes
		hasQuotes := lexer.at() == `"`
		if hasQuotes {
			strValue += lexer.eat()
		}

		// Get hd-start-identifier
		hdStartIdentifier := lexer.getName(true)
		strValue += hdStartIdentifier

		// Process closing quote
		if hasQuotes {
			if lexer.at() != `"` {
				return "", phpError.NewError(`Invalid heredoc string literal: Expected closing quote '"', Got: '%s' in %s:%d:%d`, lexer.at(), lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrPos)
			}
			strValue += lexer.eat()
		}

		// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-heredoc-string-literal
		// No white space is permitted between the start identifier and the new-line that follows.
		if !lexer.isNewLine(false) {
			return "", phpError.NewError("Invalid heredoc string literal: Expected new line, Got: '%s' in %s:%d:%d", lexer.at(), lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrPos)
		}
		strValue += lexer.getAndEatNewLine()

		lastWasNewLine := true
		for !lexer.isEof() {
			// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-heredoc-string-literal
			// No white space is permitted between the new-line and the end identifier that follows.
			if lastWasNewLine && lexer.nextN(len(hdStartIdentifier)) == hdStartIdentifier {
				strValue += lexer.eatN(len(hdStartIdentifier))
				// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-heredoc-string-literal
				// Except for an optional semicolon (;)
				if lexer.at() == ";" {
					break
				}

				// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-heredoc-string-literal
				// no characters—-not even comments or white space-—are permitted between the end identifier and the new-line that terminates that source line.
				if !lexer.isNewLine(false) {
					return "", phpError.NewError("Invalid heredoc string literal: Expected new line, Got: '%s' in %s:%d:%d", lexer.at(), lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrPos)
				}

				strValue += lexer.getAndEatNewLine()
				break
			}

			lastWasNewLine = lexer.isNewLine(false)
			if lexer.at() == `\` && lexer.next(0) != "" {
				strValue += lexer.eatN(2)
				continue
			}
			if lexer.at() != `\` {
				strValue += lexer.eat()
				continue
			}
		}

		if common.IsHeredocStringLiteral(strValue) {
			lexer.popSnapShot(!eat)
			return strValue, nil
		}

		lexer.popSnapShot(true)
		return "", phpError.NewError("Invalid heredoc string literal detected in %s:%d:%d", lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrPos)
	}

	// TODO nowdoc-string-literal

	return "", phpError.NewError("Unsupported string literal detected in %s:%d:%d", lexer.file.Filename, lexer.currPos.CurrLine, lexer.currPos.CurrPos)
}

func (lexer *Lexer) getOperatorOrPunctuator(eat bool) string {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#operators-and-punctuators

	// operator-or-punctuator:: one of
	//    [   ]   (   )   {   }   .   ->   ++   --   **   *   +   -   ~   !
	//    $   /   %   <<   >>   <   >   <=   >=   ==   ===   !=   !==   ^   |
	//    &   &&   ||   ?   :   ;   =   **=   *=   /=   %=   +=   -=   .=   <<=
	//    >>=   &=   ^=   |=   ,   ??   <=>   ...   \
	// Spec-Fix: =>   @   <<<   ::

	if op := lexer.nextN(3); slices.Contains([]string{"===", "!==", "**=", "<<=", ">>=", "<=>", "...", "<<<"}, op) {
		if eat {
			lexer.eatN(3)
		}
		return op
	}
	if op := lexer.nextN(2); slices.Contains([]string{
		"->", "++", "--", "**", "<<", ">>", "<=", ">=", "==", "!=", "&&",
		"||", "*=", "/=", "%=", "+=", "-=", ".=", "&=", "^=", "|=", "??",
		"=>", "::",
	}, op) {
		if eat {
			lexer.eatN(2)
		}
		return op
	}
	if op := lexer.at(); slices.Contains([]string{
		"[", "]", "(", ")", "{", "}", ".", "*", "+", "-", "~", "!", "$",
		"/", "%", "<", ">", "^", "|", "&", "?", ":", ";", "=", ",", "\\",
		"@",
	}, op) {
		if eat {
			lexer.eat()
		}
		return op
	}
	return ""
}

func (lexer *Lexer) isNewLineChar(char string) bool {
	return char == "\n" || char == "\r"
}

func (lexer *Lexer) isNewLine(eat bool) bool {
	// Spec: https://phplang.org/spec/19-grammar.html#grammar-new-line

	// new-line::
	//    Carriage-return character (0x0D)
	//    Line-feed character (0x0A)
	//    Carriage-return character (0x0D) followed by line-feed character (0x0A)

	if lexer.nextN(2) == "\r\n" {
		if eat {
			lexer.eatN(2)
		}
		return true
	}
	if lexer.at() == "\n" || lexer.at() == "\r" {
		if eat {
			lexer.eat()
		}
		return true
	}
	return false
}

func (lexer *Lexer) getAndEatNewLine() string {
	if lexer.nextN(2) == "\r\n" {
		return lexer.eatN(2)
	}
	if lexer.at() == "\n" || lexer.at() == "\r" {
		return lexer.eat()
	}
	return ""
}

func (lexer *Lexer) isWhiteSpaceChar(char string) bool {
	return char == " " || char == "\t"
}

func (lexer *Lexer) isWhiteSpace(eat bool) bool {
	// Spec: https://phplang.org/spec/19-grammar.html#grammar-white-space

	// white-space::
	//    white-space-character
	//    white-space   white-space-character

	// white-space-character::
	//    new-line
	//    Space character (0x20)
	//    Horizontal-tab character (0x09)

	if lexer.isNewLine(eat) {
		lexer.pushSnapShot()
		if !eat {
			lexer.currPos.CurrPos++
		}

		lexer.isWhiteSpace(eat)

		lexer.popSnapShot(!eat)

		return true
	}

	if lexer.at() == " " || lexer.at() == "\t" {
		if eat {
			lexer.eat()
		}

		lexer.pushSnapShot()
		if !eat {
			lexer.currPos.CurrPos++
		}

		lexer.isWhiteSpace(eat)

		lexer.popSnapShot(!eat)

		return true
	}

	return false
}
