package strings

import (
	"QIQ/cmd/qiq/ast"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/funcParamValidator"
	"QIQ/cmd/qiq/runtime/stdlib/variableHandling"
	"QIQ/cmd/qiq/runtime/values"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"regexp"
	goStrings "strings"
)

func Register(environment runtime.Environment) {
	// Category: String Functions
	environment.AddNativeFunction("bin2hex", nativeFn_bin2hex)
	environment.AddNativeFunction("chr", nativeFn_chr)
	environment.AddNativeFunction("hex2bin", nativeFn_hex2bin)
	environment.AddNativeFunction("implode", nativeFn_implode)
	environment.AddNativeFunction("join", nativeFn_implode)
	environment.AddNativeFunction("lcfirst", nativeFn_lcfirst)
	environment.AddNativeFunction("md5", nativeFn_md5)
	environment.AddNativeFunction("nl2br", nativeFn_nl2br)
	environment.AddNativeFunction("quotemeta", nativeFn_quotemeta)
	environment.AddNativeFunction("sha1", nativeFn_sha1)
	environment.AddNativeFunction("str_contains", nativeFn_str_contains)
	environment.AddNativeFunction("str_ends_with", nativeFn_str_ends_with)
	environment.AddNativeFunction("str_repeat", nativeFn_str_repeat)
	environment.AddNativeFunction("str_starts_with", nativeFn_str_starts_with)
	environment.AddNativeFunction("strlen", nativeFn_strlen)
	environment.AddNativeFunction("strtolower", nativeFn_strtolower)
	environment.AddNativeFunction("strtoupper", nativeFn_strtoupper)
	environment.AddNativeFunction("substr", nativeFn_substr)
	environment.AddNativeFunction("ucfirst", nativeFn_ucfirst)

	// Const Category: String Constants
	// Spec: https://www.php.net/manual/en/string.constants.php
	environment.AddPredefinedConstant("CRYPT_SALT_LENGTH", values.NewInt(1))
	environment.AddPredefinedConstant("CRYPT_STD_DES", values.NewInt(1))
	environment.AddPredefinedConstant("CRYPT_EXT_DES", values.NewInt(1))
	environment.AddPredefinedConstant("CRYPT_MD5", values.NewInt(1))
	environment.AddPredefinedConstant("CRYPT_BLOWFISH", values.NewInt(1))
	environment.AddPredefinedConstant("CRYPT_SHA256", values.NewInt(1))
	environment.AddPredefinedConstant("CRYPT_SHA512", values.NewInt(0))
	environment.AddPredefinedConstant("HTML_SPECIALCHARS", values.NewInt(1))
	environment.AddPredefinedConstant("HTML_ENTITIES", values.NewInt(2))
	environment.AddPredefinedConstant("ENT_COMPAT", values.NewInt(3))
	environment.AddPredefinedConstant("ENT_QUOTES", values.NewInt(0))
	environment.AddPredefinedConstant("ENT_NOQUOTES", values.NewInt(4))
	environment.AddPredefinedConstant("ENT_IGNORE", values.NewInt(8))
	environment.AddPredefinedConstant("ENT_SUBSTITUTE", values.NewInt(128))
	environment.AddPredefinedConstant("ENT_DISALLOWED", values.NewInt(0))
	environment.AddPredefinedConstant("ENT_HTML401", values.NewInt(16))
	environment.AddPredefinedConstant("ENT_XML1", values.NewInt(32))
	environment.AddPredefinedConstant("ENT_XHTML", values.NewInt(48))
	environment.AddPredefinedConstant("ENT_HTML5", values.NewInt(127))
	environment.AddPredefinedConstant("CHAR_MAX", values.NewInt(0))
	environment.AddPredefinedConstant("LC_CTYPE", values.NewInt(1))
	environment.AddPredefinedConstant("LC_NUMERIC", values.NewInt(2))
	environment.AddPredefinedConstant("LC_TIME", values.NewInt(3))
	environment.AddPredefinedConstant("LC_COLLATE", values.NewInt(4))
	environment.AddPredefinedConstant("LC_MONETARY", values.NewInt(6))
	environment.AddPredefinedConstant("LC_ALL", values.NewInt(5))
	environment.AddPredefinedConstant("LC_MESSAGES", values.NewInt(0))
	environment.AddPredefinedConstant("STR_PAD_LEFT", values.NewInt(1))
	environment.AddPredefinedConstant("STR_PAD_RIGHT", values.NewInt(2))
}

// -------------------------------------- bin2hex -------------------------------------- MARK: bin2hex

func nativeFn_bin2hex(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.bin2hex.php

	args, err := funcParamValidator.NewValidator("bin2hex").AddParam("$string", []string{"string"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	input := args[0].(*values.Str).Value

	var output goStrings.Builder
	for i := 0; i < len(input); i++ {
		output.WriteString(goStrings.ToLower(hex.EncodeToString([]byte{input[i]})))
	}

	return values.NewStr(output.String()), nil
}

// -------------------------------------- chr -------------------------------------- MARK: chr

func nativeFn_chr(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.chr.php

	args, err := funcParamValidator.NewValidator("chr").AddParam("$codepoint", []string{"int"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	codepoint := args[0].(*values.Int).Value

	// Spec: https://www.php.net/manual/en/function.chr.php
	// Values outside the valid range (0..255) will be bitwise and'ed with 255
	for codepoint < 0 {
		codepoint += 256
	}
	codepoint %= 256

	return values.NewStr(string(rune(codepoint))), nil
}

// -------------------------------------- hex2bin -------------------------------------- MARK: hex2bin

func nativeFn_hex2bin(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.hex2bin.php

	args, phpErr := funcParamValidator.NewValidator("hex2bin").AddParam("$string", []string{"string"}, nil).Validate(args)
	if phpErr != nil {
		return values.NewVoid(), phpErr
	}

	input := args[0].(*values.Str).Value

	// hex string must have even length
	if len(input)%2 != 0 {
		context.Interpreter.PrintError(phpError.NewWarning("Hexadecimal input string must have an even length in %s", context.Stmt.(*ast.FunctionCallExpression).Arguments[0].GetPosString()))
		return values.NewBool(false), nil
	}

	decoded, err := hex.DecodeString(input)
	if err != nil {
		return values.NewBool(false), nil
	}

	return values.NewStr(string(decoded)), nil
}

// -------------------------------------- implode -------------------------------------- MARK: implode

func nativeFn_implode(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.implode.php

	isAlternative := false
	var valArgs []values.RuntimeValue
	var err phpError.Error

	// Spec: https://www.php.net/manual/en/function.implode.php
	// implode(array $array): string
	if len(args) == 1 {
		valArgs, err = funcParamValidator.NewValidator("implode").
			AddParam("$array", []string{"array"}, nil).
			Validate(args)

		isAlternative = err == nil
	}

	// Spec: https://www.php.net/manual/en/function.implode.php
	//  implode(string $separator, array $array): string
	if !isAlternative {
		valArgs, err = funcParamValidator.NewValidator("implode").
			AddParam("$separator", []string{"string"}, nil).
			AddParam("$array", []string{"array"}, nil).
			Validate(args)
	}

	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.implode.php
	// separator [...] Defaults to an empty string.
	var separator = " "
	if !isAlternative {
		separator = valArgs[0].(*values.Str).Value
	}

	var array *values.Array
	if isAlternative {
		array = args[0].(*values.Array)
	} else {
		array = args[1].(*values.Array)
	}

	var result goStrings.Builder
	for i, key := range array.Keys {
		value, _ := array.GetElement(key)
		strValue, err := variableHandling.StrVal(value.Value)
		if err != nil {
			return values.NewStr(result.String()), err
		}
		result.WriteString(strValue)
		if i < len(array.Keys)-1 {
			result.WriteString(separator)
		}
	}

	return values.NewStr(result.String()), nil
}

// -------------------------------------- lcfirst -------------------------------------- MARK: lcfirst

func nativeFn_lcfirst(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.lcfirst.php

	args, err := funcParamValidator.NewValidator("lcfirst").AddParam("$string", []string{"string"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	input := args[0].(*values.Str).Value

	if len(input) == 0 {
		return values.NewStr(""), nil
	}

	// Spec: https://www.php.net/manual/en/function.lcfirst.php
	// Returns a string with the first character of string lowercased
	// if that character is an ASCII character in the range "A" (0x41) to "Z" (0x5a).
	if input[0] >= 'A' && input[0] <= 'Z' {
		input = string(input[0]+32) + input[1:]
	}

	return values.NewStr(input), nil
}

// -------------------------------------- md5 -------------------------------------- MARK: md5

func nativeFn_md5(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.md5.php

	args, err := funcParamValidator.NewValidator("md5").
		AddParam("$string", []string{"string"}, nil).
		AddParam("$binary", []string{"bool"}, values.NewBool(false)).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	binary := args[1].(*values.Bool).Value

	hasher := md5.New()
	hasher.Write([]byte(args[0].(*values.Str).Value))

	if binary {
		return values.NewStr(string(hasher.Sum(nil))), nil
	}

	return values.NewStr(hex.EncodeToString(hasher.Sum(nil))), nil
}

// -------------------------------------- nl2br -------------------------------------- MARK: nl2br

func nativeFn_nl2br(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.nl2br.php

	args, err := funcParamValidator.NewValidator("nl2br").
		AddParam("$string", []string{"string"}, nil).
		AddParam("$use_xhtml", []string{"bool"}, values.NewBool(true)).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	var br string
	if args[1].(*values.Bool).Value {
		br = "<br />"
	} else {
		br = "<br>"
	}

	input := args[0].(*values.Str).Value
	pattern := regexp.MustCompile(`\r\n|\n\r|\n|\r`)
	result := pattern.ReplaceAllStringFunc(input, func(s string) string {
		return br + s
	})

	return values.NewStr(result), nil
}

// -------------------------------------- quotemeta -------------------------------------- MARK: quotemeta

func nativeFn_quotemeta(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.quotemeta.php

	args, err := funcParamValidator.NewValidator("quotemeta").AddParam("$string", []string{"string"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	input := args[0].(*values.Str).Value

	if input == "" {
		return values.NewBool(false), nil
	}

	var output goStrings.Builder
	for i := 0; i < len(input); i++ {
		if goStrings.ContainsAny(string(input[i]), `.\+*?[^($)`) {
			output.WriteByte('\\')
		}
		output.WriteByte(input[i])
	}

	return values.NewStr(output.String()), nil
}

// -------------------------------------- sha1 -------------------------------------- MARK: sha1

func nativeFn_sha1(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.sha1.php

	args, err := funcParamValidator.NewValidator("sha1").
		AddParam("$string", []string{"string"}, nil).
		AddParam("$binary", []string{"bool"}, values.NewBool(false)).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	binary := args[1].(*values.Bool).Value

	hasher := sha1.New()
	hasher.Write([]byte(args[0].(*values.Str).Value))

	if binary {
		return values.NewStr(string(hasher.Sum(nil))), nil
	}

	return values.NewStr(hex.EncodeToString(hasher.Sum(nil))), nil
}

// -------------------------------------- str_contains -------------------------------------- MARK: str_contains

func nativeFn_str_contains(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.str-contains.php

	args, err := funcParamValidator.NewValidator("str_contains").
		AddParam("$haystack", []string{"string"}, nil).
		AddParam("$needle", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	haystack := args[0].(*values.Str).Value
	needle := args[1].(*values.Str).Value

	return values.NewBool(goStrings.Contains(haystack, needle)), nil
}

// -------------------------------------- str_ends_with -------------------------------------- MARK: str_ends_with

func nativeFn_str_ends_with(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.str-ends-with.php

	args, err := funcParamValidator.NewValidator("str_ends_with").
		AddParam("$haystack", []string{"string"}, nil).
		AddParam("$needle", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	haystack := args[0].(*values.Str).Value
	needle := args[1].(*values.Str).Value

	return values.NewBool(goStrings.HasSuffix(haystack, needle)), nil
}

// -------------------------------------- str_repeat -------------------------------------- MARK: str_repeat

func nativeFn_str_repeat(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.str-repeat.php

	args, err := funcParamValidator.NewValidator("str_repeat").
		AddParam("$string", []string{"string"}, nil).
		AddParam("$times", []string{"int"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	input := args[0].(*values.Str).Value
	times := args[1].(*values.Int).Value

	// Spec: https://www.php.net/manual/en/function.str-repeat.php
	// times has to be greater than or equal to 0.
	// If the times is set to 0, the function will return an empty string.
	if times < 0 {
		return values.NewVoid(), phpError.NewError("Uncaught ValueError: str_repeat(): Argument #2 ($times) must be greater than or equal to 0")
	}

	return values.NewStr(goStrings.Repeat(input, int(times))), nil
}

// -------------------------------------- str_starts_with -------------------------------------- MARK: str_starts_with

func nativeFn_str_starts_with(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.str-starts-with.php

	args, err := funcParamValidator.NewValidator("str_starts_with").
		AddParam("$haystack", []string{"string"}, nil).
		AddParam("$needle", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	haystack := args[0].(*values.Str).Value
	needle := args[1].(*values.Str).Value

	return values.NewBool(goStrings.HasPrefix(haystack, needle)), nil
}

// -------------------------------------- strlen -------------------------------------- MARK: strlen

func nativeFn_strlen(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.strlen.php

	args, err := funcParamValidator.NewValidator("strlen").AddParam("$string", []string{"string"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewInt(int64(len(args[0].(*values.Str).Value))), nil
}

// -------------------------------------- strtolower -------------------------------------- MARK: strtolower

func nativeFn_strtolower(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.strtolower.php

	args, err := funcParamValidator.NewValidator("strtolower").AddParam("$string", []string{"string"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.strtolower.php
	// Bytes in the range "A" (0x41) to "Z" (0x5a) will be converted to the corresponding lowercase letter by adding 32 to each byte value.

	input := args[0].(*values.Str).Value

	for i := 0; i < len(input); i++ {
		if input[i] >= 'A' && input[i] <= 'Z' {
			input = input[:i] + string(input[i]+32) + input[i+1:]
		}
	}

	return values.NewStr(input), nil
}

// -------------------------------------- strtoupper -------------------------------------- MARK: strtoupper

func nativeFn_strtoupper(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.strtoupper.php

	args, err := funcParamValidator.NewValidator("strtoupper").AddParam("$string", []string{"string"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// Spec: https://www.php.net/manual/en/function.strtoupper.php
	// Bytes in the range "a" (0x61) to "z" (0x7a) will be converted to the corresponding uppercase letter by subtracting 32 from each byte value.

	input := args[0].(*values.Str).Value

	for i := 0; i < len(input); i++ {
		if input[i] >= 'a' && input[i] <= 'z' {
			input = input[:i] + string(input[i]-32) + input[i+1:]
		}
	}

	return values.NewStr(input), nil
}

// -------------------------------------- substr -------------------------------------- MARK: substr

func nativeFn_substr(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.substr.php

	args, err := funcParamValidator.NewValidator("substr").
		AddParam("$string", []string{"string"}, nil).
		AddParam("$offset", []string{"int"}, nil).
		AddParam("$length", []string{"null", "int"}, values.NewNull()).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	input := args[0].(*values.Str).Value
	offset := int(args[1].(*values.Int).Value)
	lengthVal := args[2]

	strLen := len(input)

	// Handle negative offset
	if offset < 0 {
		offset = max(strLen+offset, 0)
	}
	if offset > strLen {
		return values.NewStr(""), nil
	}

	// Determine length
	var length int
	if lengthVal.GetType() == values.NullValue {
		length = strLen - offset
	} else {
		length = int(lengthVal.(*values.Int).Value)
		if length < 0 {
			length = (strLen - offset) + length
			if length < 0 {
				return values.NewStr(""), nil
			}
		}
	}

	// Clamp length to available string
	end := min(offset+length, strLen)
	if end < offset {
		return values.NewStr(""), nil
	}

	return values.NewStr(input[offset:end]), nil
}

// -------------------------------------- ucfirst -------------------------------------- MARK: ucfirst

func nativeFn_ucfirst(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ucfirst.php

	args, err := funcParamValidator.NewValidator("ucfirst").AddParam("$string", []string{"string"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	input := args[0].(*values.Str).Value

	if len(input) == 0 {
		return values.NewStr(""), nil
	}

	// Spec: https://www.php.net/manual/en/function.ucfirst.php
	// Returns a string with the first character of string capitalized,
	// if that character is an ASCII character in the range from "a" (0x61) to "z" (0x7a).
	if input[0] >= 'a' && input[0] <= 'z' {
		input = string(input[0]-32) + input[1:]
	}

	return values.NewStr(input), nil
}

// TODO addcslashes
// TODO addslashes
// TODO chop
// TODO chunk_split
// TODO convert_uudecode
// TODO convert_uuencode
// TODO count_hars
// TODO crc32
// TODO crypt
// TODO explode
// TODO fprintf
// TODO get_html_translation_table
// TODO hebrev
// TODO html_entity_decode
// TODO htmlentities
// TODO htmlspecialchars
// TODO htmlspecialchars_decode
// TODO levenshtein
// TODO localeconv
// TODO ltrim
// TODO md5_file
// TODO metaphone
// TODO money_format
// TODO nl_langinfo
// TODO number_format
// TODO ord
// TODO parse_str
// TODO printf
// TODO quoted_printable_decode
// TODO quoted_printable_encode
// TODO rtrim
// TODO setlocale
// TODO sha1_file
// TODO similar_text
// TODO soundex
// TODO sprintf
// TODO sscanf
// TODO str_decrement
// TODO str_getcsv
// TODO str_increment
// TODO str_ireplace
// TODO str_pad
// TODO str_replace
// TODO str_rot13
// TODO str_shuffle
// TODO str_split
// TODO str_word_count
// TODO strcasecmp
// TODO strchr
// TODO strcmp
// TODO strcoll
// TODO strcspn
// TODO strip_tags
// TODO stripcslashes
// TODO stripos
// TODO stripslashes
// TODO stristr
// TODO strnatcasecmp
// TODO strnatcmp
// TODO strncasecmp
// TODO strncmp
// TODO strpbrk
// TODO strpos
// TODO strrchr
// TODO strrev
// TODO strripos
// TODO strrpos
// TODO strspn
// TODO strstr
// TODO strtok
// TODO strtr
// TODO substr_compare
// TODO substr_count
// TODO substr_replace
// TODO trim
// TODO ucwords
// TODO vfprintf
// TODO vprintf
// TODO vsprintf
// TODO wordwrap
// Deprecated:
// TODO convert_cyr_string
// TODO hebrevc
// TODO utf8_decode
// TODO utf8_encode
