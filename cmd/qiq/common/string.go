package common

import (
	"regexp"
	"strings"
)

const (
	nameRegex string = `[_a-zA-Z\x80-\xff][_a-zA-Z0-9\x80-\xff]*`
)

func IsVariableName(name string) bool {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-variable-name

	// variable-name::
	//    $   name

	match, _ := regexp.MatchString(`^\$`+nameRegex+`$`, name)
	return match
}

func IsQualifiedName(name string) bool {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-qualified-name

	// qualified-name::
	//    namespace-name-as-a-prefix(opt)   name

	// namespace-name-as-a-prefix::
	//    \
	//    \(opt)   namespace-name   \
	//    namespace   \
	//    namespace   \   namespace-name

	match, _ := regexp.MatchString(`^\\?(`+nameRegex+`\\)*`+nameRegex+`$`, name)
	return match
}

func IsName(name string) bool {
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

	// digit:: one of
	//    0   1   2   3   4   5   6   7   8   9

	match, _ := regexp.MatchString(`^`+nameRegex+`$`, name)
	return match
}

func IsNondigit(char string) bool {
	// Spec: https://phplang.org/spec/19-grammar.html#grammar-nondigit

	// nondigit:: one of
	//    _
	//    a   b   c   d   e   f   g   h   i   j   k   l   m
	//    n   o   p   q   r   s   t   u   v   w   x   y   z
	//    A   B   C   D   E   F   G   H   I   J   K   L   M
	//    N   O   P   Q   R   S   T   U   V   W   X   Y   Z

	match, _ := regexp.MatchString("^[a-zA-Z_]$", char)
	return match
}

func IsNameNondigit(char string) bool {
	// Spec: https://phplang.org/spec/19-grammar.html#grammar-name-nondigit

	// name-nondigit::
	//    nondigit
	//    one of the characters 0x80–0xff

	b := []byte(char)[0]
	return IsNondigit(char) || b >= 0x80
}

func IsSingleQuotedStringLiteral(str string) bool {
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

	match, _ := regexp.MatchString(`^[bB]?'(\\?[^'\\]|\\'|\\\\)*'`, str)
	return match
}

func ReplaceSingleQuoteControlChars(str string) string {
	findCtrlChars := regexp.MustCompile(`(\\?[^'\\]|\\'|\\\\)`)
	return findCtrlChars.ReplaceAllStringFunc(str,
		func(sub string) string {
			switch sub {
			case `\'`:
				return `'`
			case `\\`:
				return `\`
			default:
				return sub
			}
		},
	)
}

func extractStringContent(str string) string {
	if strings.ToLower(str[0:1]) == "b" {
		return str[2 : len(str)-1]
	}
	return str[1 : len(str)-1]
}

func SingleQuotedStringLiteralToString(str string) string {
	return ReplaceSingleQuoteControlChars(extractStringContent(str))
}

func IsDoubleQuotedStringLiteral(str string) bool {
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

	match, _ := regexp.MatchString(
		`^[bB]?"(`+
			`(\\"|\\\\|\\\$|\\e|\\f|\\n|\\r|\\t|\\v)|`+
			`(\\[0-7]{1,3})|`+
			`(\\[xX][0-9a-fA-F]{1,2})|`+
			`(\\u\{[0-9a-fA-F]+\})|`+
			`([^"\\])|`+
			`(\\[^"\\$efnrtvxX]|\\[0-7])`+
			`)*"$`,
		str)
	return match
}

func ReplaceDoubleQuoteControlChars(str string) string {
	findCtrlChars := regexp.MustCompile(
		`(\\"|\\\\|\\\$|\\e|\\f|\\n|\\r|\\t|\\v)|` +
			`(\\[0-7]{1,3})|` +
			`(\\[xX][0-9a-fA-F]{1,2})|` +
			`(\\u\{[0-9a-fA-F]+\})|` +
			`([^"\\])|` +
			`(\\[^"\\$efnrtvxX]|\\[0-7])`,
	)

	return findCtrlChars.ReplaceAllStringFunc(str,
		func(sub string) string {
			switch sub {
			case `\"`:
				return `"`
			case `\\`:
				return `\`
			case `\r`:
				return "\r"
			case `\n`:
				return "\n"
			case `\t`:
				return "\t"
			case `\f`:
				return "\f"
			case `\v`:
				return "\v"
			default:
				return sub
			}
		},
	)
}

func DoubleQuotedStringLiteralToString(str string) string {
	return ReplaceDoubleQuoteControlChars(extractStringContent(str))
}

func IsHeredocStringLiteral(str string) bool {
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

	match, _ := regexp.MatchString(
		`^[bB]?<<<[ \t]*"?`+nameRegex+`"?\r?\n`+
			`((`+
			`(\\"|\\\\|\\\$|\\e|\\f|\\n|\\r|\\t|\\v)|`+
			`(\\[0-7]{1,3})|`+
			`(\\[xX][0-9a-fA-F]{1,2})|`+
			`(\\u\{[0-9a-fA-F]+\})|`+
			`([^\\])|`+
			`(\\[^"\\$efnrtvxX]|\\[0-7])]`+
			`)*\r?\n)?`+
			nameRegex+`;?\r?\n?$`,
		str)
	return match
}

func HeredocStringLiteralToString(str string) string {
	lines := strings.SplitN(str, "\n", 2)
	if len(lines) < 2 {
		return ""
	}

	body := lines[1]
	bodyLines := strings.Split(body, "\n")
	if len(bodyLines) < 2 {
		return ""
	}
	body = strings.Join(bodyLines[:len(bodyLines)-1], "\n")

	return ReplaceHeredocControlChars(body)
}

func ReplaceHeredocControlChars(str string) string {
	findCtrlChars := regexp.MustCompile(
		`(\\\\|\\\$|\\e|\\f|\\n|\\r|\\t|\\v)|` +
			`(\\[0-7]{1,3})|` +
			`(\\[xX][0-9a-fA-F]{1,2})|` +
			`(\\u\{[0-9a-fA-F]+\})|` +
			`([^"\\])|` +
			`(\\[^"\\$efnrtvxX]|\\[0-7])`,
	)

	return findCtrlChars.ReplaceAllStringFunc(str,
		func(sub string) string {
			switch sub {
			case `\\`:
				return `\`
			case `\r`:
				return "\r"
			case `\n`:
				return "\n"
			case `\t`:
				return "\t"
			default:
				return sub
			}
		},
	)
}

func TrimTrailingLineBreak(str string) string {
	if len(str) > 0 && str[len(str)-1] == '\n' {
		str = str[:len(str)-1]
	}
	return str
}

func TrimTrailingLineBreaks(str string) string {
	for len(str) > 0 && str[len(str)-1] == '\n' {
		str = str[:len(str)-1]
	}
	return str
}

func TrimLineBreaks(str string) string { return strings.Trim(str, "\n") }

func ReplaceUnderscores(str string) string { return strings.ReplaceAll(str, "_", "") }

func ReplaceAtPos(str string, new string, pos int) string {
	out := []rune(str)
	out[pos] = []rune(new)[0]
	return string(out)
}

func ExtendWithSpaces(str string, length int) string {
	input := []rune(str)
	if len(input) > length {
		return str
	}

	padding := make([]rune, length-len(input))
	for i := range padding {
		padding[i] = ' '
	}
	return string(append(input, padding...))
}
