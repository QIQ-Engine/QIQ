package interpreter

import (
	"QIQ/cmd/qiq/config"
	"QIQ/cmd/qiq/ini"
	"QIQ/cmd/qiq/phpError"
	"fmt"
	"testing"
)

// -------------------------------------- ctype -------------------------------------- MARK: ctype

func TestLibCType(t *testing.T) {
	// ctype_alnum
	testInputOutput(t, `<?php var_dump(ctype_alnum(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_alnum(" "));`, "bool(false)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-alnum.php
	testInputOutput(t, `<?php var_dump(ctype_alnum("AbCd1zyZ9"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_alnum('foo!#$bar'));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_alnum("áeiou"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_alnum("aeiöu"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_alnum("ñ"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_alnum("aeiou"));`, "bool(true)\n")

	// ctype_alpha
	testInputOutput(t, `<?php var_dump(ctype_alpha(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_alpha(" "));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_alpha("aeiou"));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-alpha.php
	testInputOutput(t, `<?php var_dump(ctype_alpha("KjgWZC"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_alpha("arf12"));`, "bool(false)\n")

	// ctype_cntrl
	testInputOutput(t, `<?php var_dump(ctype_cntrl(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_cntrl(" "));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_cntrl("\n"));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-cntrl.php
	testInputOutput(t, `<?php var_dump(ctype_cntrl("arf12"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_cntrl("\n\r\t"));`, "bool(true)\n")

	// ctype_digit
	testInputOutput(t, `<?php var_dump(ctype_digit(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_digit(" "));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_digit("0"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_digit("1"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_digit("99"));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-digit.php
	testInputOutput(t, `<?php var_dump(ctype_digit("10002"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_digit("1820.20"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_digit("wsl!12"));`, "bool(false)\n")

	// ctype_graph
	testInputOutput(t, `<?php var_dump(ctype_graph(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_graph(" "));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_graph("ä"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_graph("a"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_graph("#"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_graph("a_-;."));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-graph.php
	testInputOutput(t, `<?php var_dump(ctype_graph("asdf\n\r\t"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_graph("arf12"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_graph("LKA#@%.54"));`, "bool(true)\n")

	// ctype_lower
	testInputOutput(t, `<?php var_dump(ctype_lower(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_lower(" "));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_lower("A"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_lower("a"));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-lower.php
	testInputOutput(t, `<?php var_dump(ctype_lower("aac123"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_lower("qiutoas"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_lower("QASsdks"));`, "bool(false)\n")

	// ctype_print
	testInputOutput(t, `<?php var_dump(ctype_print(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_print("ä"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_print("€"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_print(" "));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_print("A"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_print("a"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_print("_"));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-print.php
	testInputOutput(t, `<?php var_dump(ctype_print("asdf\n\r\t"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_print("arf12"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_print("LKA#@%.54"));`, "bool(true)\n")

	// ctype_punct
	testInputOutput(t, `<?php var_dump(ctype_punct(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_punct(" "));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_punct("a"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_punct("°"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_punct("²"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_punct("³"));`, "bool(false)\n")
	testInputOutput(t, "<?php var_dump(ctype_punct('`'));", "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_punct('.+*~\'#-_.:,;<>|^!"$%&/{([)]=}?\\'));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-punct.php
	testInputOutput(t, `<?php var_dump(ctype_punct('ABasdk!@!$#'));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_punct('!@ # $'));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_punct('*&$()'));`, "bool(true)\n")

	// ctype_space
	testInputOutput(t, `<?php var_dump(ctype_space(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_space("a"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_space("1"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_space(" "));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_space("\n"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_space("\t"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_space("\r"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_space("\v"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_space("\f"));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-space.php
	testInputOutput(t, `<?php var_dump(ctype_space("\n\r\t"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_space("\narf12"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_space('\n\r\t'));`, "bool(false)\n")

	// ctype_upper
	testInputOutput(t, `<?php var_dump(ctype_upper(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_upper(" "));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_upper("a"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_upper("1"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_upper("-"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_upper("A"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_upper("ABC"));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-upper.php
	testInputOutput(t, `<?php var_dump(ctype_upper("AKLWC139"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_upper("LMNSDO"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_upper("akwSKWsm"));`, "bool(false)\n")

	// ctype_xdigit
	testInputOutput(t, `<?php var_dump(ctype_xdigit(""));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_xdigit(" "));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_xdigit("G"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_xdigit("ABCDEF"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_xdigit("1234567890"));`, "bool(true)\n")
	// Examples from https://www.php.net/manual/en/function.ctype-xdigit.php
	testInputOutput(t, `<?php var_dump(ctype_xdigit("AB10BC99"));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(ctype_xdigit("AR1012"));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ctype_xdigit("ab12bc99"));`, "bool(true)\n")
}

// -------------------------------------- math -------------------------------------- MARK: math

func TestLibMath(t *testing.T) {
	// abs
	testInputOutput(t, `<?php var_dump(abs(-4.2));`, "float(4.2)\n")
	testInputOutput(t, `<?php var_dump(abs(5));`, "int(5)\n")
	testInputOutput(t, `<?php var_dump(abs(-5));`, "int(5)\n")

	// acos
	testInputOutput(t, `<?php var_dump(acos(1.0));`, "float(0)\n")
	testInputOutput(t, `<?php var_dump(acos(0.5)/M_PI*180);`, "float(60)\n")

	// acosh
	testInputOutput(t, `<?php var_dump(acosh(1.0));`, "float(0)\n")

	// asin
	testInputOutput(t, `<?php var_dump(asin(0.0));`, "float(0)\n")

	// asinh
	testInputOutput(t, `<?php var_dump(asinh(0.0));`, "float(0)\n")

	// clamp
	// Examples from https://php.watch/versions/8.6/clamp
	// - Integer clamping
	testInputOutput(t, `<?php var_dump(clamp(5, 0, 100));`, "int(5)\n")
	testInputOutput(t, `<?php var_dump(clamp(0, 0, 100));`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(clamp(-5, 0, 100));`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(clamp(100, 0, 100));`, "int(100)\n")
	testInputOutput(t, `<?php var_dump(clamp(105, 0, 100));`, "int(100)\n")
	testInputOutput(t, `<?php var_dump(clamp(105, 100, 100));`, "int(100)\n")
	// - Float clamping
	testInputOutput(t, `<?php var_dump(clamp(3.01, 1.6, 4.2));`, "float(3.01)\n")
	testInputOutput(t, `<?php var_dump(clamp(10.0, 1.6, 4.2));`, "float(4.2)\n")
	testInputOutput(t, `<?php var_dump(clamp(0, M_1_PI, M_2_PI));`, "float(0.3183098861837907)\n")
	// - Integer and float clamping
	testInputOutput(t, `<?php var_dump(clamp(5, 10, 12.5));`, "int(10)\n")
	testInputOutput(t, `<?php var_dump(clamp(5, 10.0, 12));`, "float(10)\n")
	testInputOutput(t, `<?php var_dump(clamp(3.14, 10, 20));`, "int(10)\n")
	testInputOutput(t, `<?php var_dump(clamp(3.14, 0, 20));`, "float(3.14)\n")
	// - String values
	testInputOutput(t, `<?php var_dump(clamp('P', 'A', 'Z'));`, "string(1) \"P\"\n")
	testInputOutput(t, `<?php var_dump(clamp('P', 'X', 'Z'));`, "string(1) \"X\"\n")
	testInputOutput(t, `<?php var_dump(clamp('P', 'A', 'C'));`, "string(1) \"C\"\n")
	testInputOutput(t, `<?php var_dump(clamp('AAA', 'AA', 'Z'));`, "string(3) \"AAA\"\n")
	// - Boolean values
	testInputOutput(t, `<?php var_dump(clamp(5, false, true));`, "int(5)\n")
	testInputOutput(t, `<?php var_dump(clamp(true, false, true));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(clamp(true, false, false));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(clamp(false, true, true));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(clamp(false, true, 5));`, "bool(true)\n")
	// - Arrays
	testInputOutput(t, `<?php var_dump(clamp(5, [], []));`, "array(0) {\n}\n")
	testInputOutput(t, `<?php var_dump(clamp(5, 0, []));`, "int(5)\n")
	testInputOutput(t, `<?php var_dump(clamp(5, false, []));`, "int(5)\n")
	testInputOutput(t, `<?php var_dump(clamp([3], [1], [5]));`, "array(1) {\n  [0]=>\n  int(3)\n}\n")
	testInputOutput(t, `<?php var_dump(clamp([1], [3], [5]));`, "array(1) {\n  [0]=>\n  int(3)\n}\n")
	testInputOutput(t, `<?php var_dump(clamp([1, 4], [3], [1, 5]));`, "array(2) {\n  [0]=>\n  int(1)\n  [1]=>\n  int(4)\n}\n")
	// - Objects
	// TODO clamp - objects
	// clamp(new DateTimeImmutable('2026-01-08'), new DateTimeImmutable('2026-01-01'), new DateTimeImmutable('2026-12-31')); // DateTimeImmutable('2026-01-08')
	// clamp(new BCMath\Number('36'), new BCMath\Number('16'), new BCMath\Number('42')); // BCMath\Number('36')

	// Examples from https://wiki.php.net/rfc/clamp_v2
	// - Special numbers
	// TODO clamp - special numbers
	// clamp(M_PI, -INF, INF) // 3.141592653589793
	// clamp(NAN, 4, 6) // NAN
	// - Errors
	// TODO clamp - errors
	// TODO clamp - errors - change error type ValueError: clamp(): Argument #2 ($min) must be smaller than or equal to argument #3 ($max)
	testForError(t, `<?php clamp(4, 8, 6);`, phpError.NewError("clamp(): Argument #2 ($min) must be smaller than or equal to argument #3 ($max) in %s:1:7", TEST_FILE_NAME))
	// clamp(4, NAN, 6) // Throws ValueError: clamp(): Argument #2 ($min) cannot be NAN
	// clamp(4, 6, NAN) // Throws ValueError: clamp(): Argument #3 ($max) cannot be NAN

	// pi
	testInputOutput(t, `<?php var_dump(M_PI === pi());`, "bool(true)\n")
}

// -------------------------------------- ob_ functions -------------------------------------- MARK: ob_ functions

func TestObFunctions(t *testing.T) {
	// Implicit flush
	testInputOutput(t, `<?php ob_start(); echo '123';`, "123")

	// ob_clean
	testInputOutput(t, `<?php ob_start(); echo '123'; ob_clean(); echo '456';`, "456")
	// ob_flush
	testInputOutput(t, `<?php ob_start(); echo '123'; ob_flush(); echo '456'; ob_end_clean();`, "123")
	// ob_end_clean
	testInputOutput(t, `<?php ob_start(); echo '123'; ob_end_clean(); echo '456';`, "456")
	// ob_end_flush
	testInputOutput(t, `<?php ob_start(); echo '123'; ob_end_flush(); echo '456';`, "123456")
	// ob_get_clean
	testInputOutput(t, `<?php ob_start(); echo '123'; $ob = ob_get_clean(); echo '456' . $ob;`, "456123")
	// ob_get_flush
	testInputOutput(t, `<?php ob_start(); echo '123'; $ob = ob_get_flush(); echo '456' . $ob;`, "123456123")
	// ob_get_contents
	testInputOutput(t, `<?php ob_start(); echo '123'; $ob = ob_get_contents(); ob_end_clean(); echo '456' . $ob;`, "456123")

	// ob_get_length
	testInputOutput(t, `<?php ob_start();
		echo "Helloä "; $len1 = ob_get_length();
		echo "World"; $len2 = ob_get_length();
		ob_end_clean(); echo $len1 . ", " . $len2;`,
		"8, 13")
	// No buffering actrive
	testInputOutput(t, `<?php
		echo "Hello "; $len1 = ob_get_length();
		echo serialize($len1);`,
		"Hello b:0;")

	// ob_get_level
	testInputOutput(t, `<?php ob_start(); echo 'A' . ob_get_level(); ob_start(); echo 'B' . ob_get_level();`, "A1B2")
	// Stacked output buffers
	testInputOutput(t,
		`<?php
            echo 0;
                ob_start();
                    ob_start();
                        ob_start();
                            ob_start();
                                echo 1;
                            ob_end_flush();
                            echo 2;
                        $ob = ob_get_clean();
                    echo 3;
                    ob_flush();
                    ob_end_clean();
                echo 4;
                ob_end_flush();
            echo '-' . $ob;
        ?>`,
		"034-12")
}

// -------------------------------------- date -------------------------------------- MARK: date

func TestLibDate(t *testing.T) {
	// checkdate
	testInputOutput(t, `<?php var_dump(checkdate(12, 31, 2000));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(checkdate(2, 29, 2001));`, "bool(false)\n")

	// date
	// Day
	testInputOutput(t, `<?php var_dump(date('d', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"04\"\n")
	testInputOutput(t, `<?php var_dump(date('j', mktime(12, 13, 14, 05, 04, 2024)));`, "string(1) \"4\"\n")
	testInputOutput(t, `<?php var_dump(date('z', mktime(12, 13, 14, 05, 04, 2024)));`, "string(3) \"124\"\n")
	testInputOutput(t, `<?php var_dump(date('w', mktime(12, 13, 14, 05, 04, 2024)));`, "string(1) \"6\"\n")
	testInputOutput(t, `<?php var_dump(date('w', mktime(12, 13, 14, 05, 05, 2024)));`, "string(1) \"0\"\n")
	testInputOutput(t, `<?php var_dump(date('N', mktime(12, 13, 14, 05, 04, 2024)));`, "string(1) \"6\"\n")
	testInputOutput(t, `<?php var_dump(date('N', mktime(12, 13, 14, 05, 05, 2024)));`, "string(1) \"7\"\n")
	// Week
	testInputOutput(t, `<?php var_dump(date('W', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"18\"\n")
	// Month
	testInputOutput(t, `<?php var_dump(date('m', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"05\"\n")
	testInputOutput(t, `<?php var_dump(date('n', mktime(12, 13, 14, 05, 04, 2024)));`, "string(1) \"5\"\n")
	testInputOutput(t, `<?php var_dump(date('t', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"31\"\n")
	// Year
	testInputOutput(t, `<?php var_dump(date('Y', mktime(12, 13, 14, 05, 04, 2024)));`, "string(4) \"2024\"\n")
	testInputOutput(t, `<?php var_dump(date('y', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"24\"\n")
	testInputOutput(t, `<?php var_dump(date('L', mktime(12, 13, 14, 05, 04, 2024)));`, "string(1) \"1\"\n")
	// Time
	testInputOutput(t, `<?php var_dump(date('i', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"13\"\n")
	testInputOutput(t, `<?php var_dump(date('i', mktime(12, 00, 14, 05, 04, 2024)));`, "string(2) \"00\"\n")
	testInputOutput(t, `<?php var_dump(date('s', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"14\"\n")
	testInputOutput(t, `<?php var_dump(date('s', mktime(12, 13, 00, 05, 04, 2024)));`, "string(2) \"00\"\n")
	testInputOutput(t, `<?php var_dump(date('G', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"12\"\n")
	testInputOutput(t, `<?php var_dump(date('G', mktime(20, 13, 14, 05, 04, 2024)));`, "string(2) \"20\"\n")
	testInputOutput(t, `<?php var_dump(date('H', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"12\"\n")
	testInputOutput(t, `<?php var_dump(date('H', mktime(03, 13, 14, 05, 04, 2024)));`, "string(2) \"03\"\n")
	testInputOutput(t, `<?php var_dump(date('H', mktime(20, 13, 14, 05, 04, 2024)));`, "string(2) \"20\"\n")
	testInputOutput(t, `<?php var_dump(date('g', mktime(00, 13, 14, 05, 04, 2024)));`, "string(2) \"12\"\n")
	testInputOutput(t, `<?php var_dump(date('g', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"12\"\n")
	testInputOutput(t, `<?php var_dump(date('g', mktime(14, 13, 14, 05, 04, 2024)));`, "string(1) \"2\"\n")
	testInputOutput(t, `<?php var_dump(date('h', mktime(00, 13, 14, 05, 04, 2024)));`, "string(2) \"12\"\n")
	testInputOutput(t, `<?php var_dump(date('h', mktime(12, 13, 14, 05, 04, 2024)));`, "string(2) \"12\"\n")
	testInputOutput(t, `<?php var_dump(date('h', mktime(14, 13, 14, 05, 04, 2024)));`, "string(2) \"02\"\n")

	// getdate
	testInputOutput(t, `<?php print_r(getdate(1722707036));`,
		"Array\n(\n    [seconds] => 56\n    [minutes] => 43\n    [hours] => 17\n    [mday] => 3\n    [wday] => 6\n    [mon] => 8\n"+
			"    [year] => 2024\n    [yday] => 215\n    [weekday] => Saturday\n    [month] => August\n    [0] => 1722707036\n)\n",
	)
}

// -------------------------------------- strings -------------------------------------- MARK: strings

func TestLibStrings(t *testing.T) {
	// bin2hex
	testInputOutput(t, `<?php var_dump(bin2hex('Hello world!'));`, "string(24) \"48656c6c6f20776f726c6421\"\n")
	testInputOutput(t, `<?php var_dump(bin2hex('Äàßê'));`, "string(16) \"c384c3a0c39fc3aa\"\n")
	testInputOutput(t, `<?php var_dump(bin2hex(''));`, "string(0) \"\"\n")

	// chr
	testInputOutput(t, `<?php var_dump(chr(60));`, "string(1) \"<\"\n")
	testInputOutput(t, `<?php var_dump(chr(60-256));`, "string(1) \"<\"\n")
	testInputOutput(t, `<?php var_dump(chr(60+256));`, "string(1) \"<\"\n")

	// hex2bin
	testInputOutput(t, `<?php var_dump(hex2bin(''));`, "string(0) \"\"\n")
	testInputOutput(t, `<?php var_dump(hex2bin('6578616d706c65206865782064617461'));`, "string(16) \"example hex data\"\n")
	testInputOutput(t, `<?php var_dump(hex2bin('6'));`, fmt.Sprintf("\nWarning: Hexadecimal input string must have an even length in %s:1:24\nbool(false)\n", TEST_FILE_NAME))

	// lcfirst
	testInputOutput(t, `<?php var_dump(lcfirst('ABC'));`, "string(3) \"aBC\"\n")
	testInputOutput(t, `<?php var_dump(lcfirst('Abc'));`, "string(3) \"abc\"\n")
	testInputOutput(t, `<?php var_dump(lcfirst('abc'));`, "string(3) \"abc\"\n")
	testInputOutput(t, `<?php var_dump(lcfirst(''));`, "string(0) \"\"\n")

	// md5
	testInputOutput(t, `<?php var_dump(md5('apple'));`, "string(32) \"1f3870be274f6c49b3e31a0c6728957f\"\n")
	testInputOutput(t, `<?php var_dump(bin2hex(md5('apple', true)));`, "string(32) \"1f3870be274f6c49b3e31a0c6728957f\"\n")
	testInputOutput(t, `<?php var_dump(md5('hello world'));`, "string(32) \"5eb63bbbe01eeed093cb22bb8f5acdc3\"\n")
	testInputOutput(t, `<?php var_dump(bin2hex(md5('hello world', true)));`, "string(32) \"5eb63bbbe01eeed093cb22bb8f5acdc3\"\n")

	// nl2br
	testInputOutput(t, "<?php var_dump(nl2br(\"a\nb\"));", "string(9) \"a<br />\nb\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"c\nd\", false));", "string(7) \"c<br>\nd\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"e\rf\"));", "string(9) \"e<br />\rf\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"g\rh\", false));", "string(7) \"g<br>\rh\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"i\r\nj\"));", "string(10) \"i<br />\r\nj\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"k\r\nl\", false));", "string(8) \"k<br>\r\nl\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"m\n\ro\"));", "string(10) \"m<br />\n\ro\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"p\n\rq\", false));", "string(8) \"p<br>\n\rq\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"r\r\rs\"));", "string(16) \"r<br />\r<br />\rs\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"t\r\ru\", false));", "string(12) \"t<br>\r<br>\ru\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"v\n\nw\"));", "string(16) \"v<br />\n<br />\nw\"\n")
	testInputOutput(t, "<?php var_dump(nl2br(\"x\n\ny\", false));", "string(12) \"x<br>\n<br>\ny\"\n")

	// quotemeta
	testInputOutput(t, `<?php var_dump(quotemeta('. \ + * ? [ ^ ] ( $ )'));`, `string(31) "\. \\ \+ \* \? \[ \^ ] \( \$ \)"`+"\n")
	testInputOutput(t, `<?php var_dump(quotemeta('Hello. (can you hear me?)'));`, `string(29) "Hello\. \(can you hear me\?\)"`+"\n")
	testInputOutput(t, `<?php var_dump(quotemeta(''));`, "bool(false)\n")

	// sha1
	testInputOutput(t, `<?php var_dump(sha1('apple'));`, "string(40) \"d0be2dc421be4fcd0172e5afceea3970e2f3d940\"\n")
	testInputOutput(t, `<?php var_dump(bin2hex(sha1('apple', true)));`, "string(40) \"d0be2dc421be4fcd0172e5afceea3970e2f3d940\"\n")
	testInputOutput(t, `<?php var_dump(sha1('hello world'));`, "string(40) \"2aae6c35c94fcfb415dbe95f408b9ce91ee846ed\"\n")
	testInputOutput(t, `<?php var_dump(bin2hex(sha1('hello world', true)));`, "string(40) \"2aae6c35c94fcfb415dbe95f408b9ce91ee846ed\"\n")

	// str_contains
	testInputOutput(t, `<?php var_dump(str_contains('abc', ''));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(str_contains('The lazy fox', 'lazy'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(str_contains('The lazy fox', 'Lazy'));`, "bool(false)\n")

	// str_ends_with
	testInputOutput(t, `<?php var_dump(str_ends_with('abc', ''));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(str_ends_with('The lazy fox', 'fox'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(str_ends_with('The lazy fox', 'Fox'));`, "bool(false)\n")

	// str_repeat
	testInputOutput(t, `<?php var_dump(str_repeat('abc', 0));`, "string(0) \"\"\n")
	testInputOutput(t, `<?php var_dump(str_repeat('abc', 1));`, "string(3) \"abc\"\n")
	testInputOutput(t, `<?php var_dump(str_repeat('abc', 2));`, "string(6) \"abcabc\"\n")
	testForError(t, `<?php var_dump(str_repeat('abc', -1));`, phpError.NewError("Uncaught ValueError: str_repeat(): Argument #2 ($times) must be greater than or equal to 0"))

	// str_starts_with
	testInputOutput(t, `<?php var_dump(str_starts_with('abc', ''));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(str_starts_with('The lazy fox', 'The'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(str_starts_with('The lazy fox', 'the'));`, "bool(false)\n")

	// strlen
	testInputOutput(t, `<?php var_dump(strlen('abcdef'));`, "int(6)\n")
	testInputOutput(t, `<?php var_dump(strlen(' ab cd '));`, "int(7)\n")
	testInputOutput(t, `<?php var_dump(strlen(' äb ćd '));`, "int(9)\n")

	// strtolower
	testInputOutput(t, `<?php var_dump(strtolower('Mary Had A Little Lamb and She LOVED It So'));`, "string(42) \"mary had a little lamb and she loved it so\"\n")
	testInputOutput(t, `<?php var_dump(strtolower(''));`, "string(0) \"\"\n")
	testInputOutput(t, `<?php var_dump(strtolower('AÄOÖUÜSß'));`, "string(12) \"aÄoÖuÜsß\"\n")

	// strtoupper
	testInputOutput(t, `<?php var_dump(strtoupper('Mary Had A Little Lamb and She LOVED It So'));`, "string(42) \"MARY HAD A LITTLE LAMB AND SHE LOVED IT SO\"\n")
	testInputOutput(t, `<?php var_dump(strtoupper(''));`, "string(0) \"\"\n")
	testInputOutput(t, `<?php var_dump(strtoupper('aäoöuüsß'));`, "string(12) \"AäOöUüSß\"\n")

	// substr
	// Spec: https://www.php.net/manual/en/function.substr.php
	testInputOutput(t, `<?php var_dump(substr("abcdef", 1));`, "string(5) \"bcdef\"\n")
	testInputOutput(t, `<?php var_dump(substr("abcdef", 1, null));`, "string(5) \"bcdef\"\n")
	testInputOutput(t, `<?php var_dump(substr("abcdef", -1));`, "string(1) \"f\"\n")
	testInputOutput(t, `<?php var_dump(substr("abcdef", -2));`, "string(2) \"ef\"\n")
	testInputOutput(t, `<?php var_dump(substr("abcdef", -3, 1));`, "string(1) \"d\"\n")
	testInputOutput(t, `<?php var_dump(substr("abcdef", 0, -1));`, "string(5) \"abcde\"\n")
	testInputOutput(t, `<?php var_dump(substr("abcdef", 2, -1));`, "string(3) \"cde\"\n")
	testInputOutput(t, `<?php var_dump(substr("abcdef", 4, -4));`, "string(0) \"\"\n")
	testInputOutput(t, `<?php var_dump(substr("abcdef", -3, -1));`, "string(2) \"de\"\n")

	// ucfirst
	testInputOutput(t, `<?php var_dump(ucfirst('ABC'));`, "string(3) \"ABC\"\n")
	testInputOutput(t, `<?php var_dump(ucfirst('abc'));`, "string(3) \"Abc\"\n")
	testInputOutput(t, `<?php var_dump(ucfirst('Abc'));`, "string(3) \"Abc\"\n")
	testInputOutput(t, `<?php var_dump(ucfirst(''));`, "string(0) \"\"\n")
}

// -------------------------------------- option_info -------------------------------------- MARK: option_info

func TestLibOptionInfo(t *testing.T) {
	// ini_get
	testInputOutput(t, `<?php var_dump(ini_get('none_existing'));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ini_get('variables_order'));`, "string(5) \"EGPCS\"\n")
	testInputOutput(t, `<?php var_dump(ini_get('error_reporting'));`, "string(5) \"32767\"\n")

	// ini_set
	testInputOutput(t, `<?php var_dump(ini_set('none_existing', true));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ini_set('variables_order', true));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(ini_set('error_reporting', E_ERROR)); var_dump(ini_set('error_reporting', E_ERROR));`, "string(5) \"32767\"\nstring(1) \"1\"\n")

	// phpversion
	testInputOutput(t, `<?= phpversion();`, config.Version)

	// zend_version
	testInputOutput(t, `<?= zend_version();`, config.QIQVersion)
}

// -------------------------------------- error_reporting -------------------------------------- MARK: error_reporting

func TestLibErrorReporting(t *testing.T) {
	// Spec: https://www.php.net/manual/en/function.error-reporting.php - Example #1

	// Turn off all error reporting
	testInputOutput(t, `<?php error_reporting(0); echo error_reporting();`, "0")
	// Report simple running errors
	testInputOutput(t, `<?php error_reporting(E_ERROR | E_WARNING | E_PARSE); echo error_reporting();`, "7")
	// Reporting E_NOTICE can be good too (to report uninitialized variables or catch variable name misspellings ...)
	testInputOutput(t, `<?php error_reporting(E_ERROR | E_WARNING | E_PARSE | E_NOTICE); echo error_reporting();`, "15")
	// Report all errors except E_NOTICE
	testInputOutput(t, `<?php error_reporting(E_ALL & ~E_NOTICE); echo error_reporting();`, "32759")
	// Report all PHP errors
	testInputOutput(t, `<?php error_reporting(E_ALL); echo error_reporting();`, "32767")
	// Report all PHP errors
	testInputOutput(t, `<?php error_reporting(-1); echo error_reporting();`, "32767")
}

// -------------------------------------- classes_object -------------------------------------- MARK: classesobject

func TestLibClassesObject(t *testing.T) {
	// class_alias
	testInputOutput(t, `<?php class C {} var_dump(class_alias('C', 'B')); $b = new B;  var_dump(get_class($b));`, "bool(true)\nstring(1) \"C\"\n")
	testInputOutput(t, `<?php class C {} var_dump(class_alias('D', 'B'));`, fmt.Sprintf("\nWarning: Class \"D\" not found in %s:1:27\nbool(false)\n", TEST_FILE_NAME))
	// TODO Add support for namespace in object creation
	// testInputOutput(t, `<?php class C {} var_dump(class_alias('C', 'Space\B')); $b = new \Space\B;  var_dump(get_class($b));`, "bool(true)\nstring(1) \"C\"\n")

	// class_exists
	testInputOutput(t, `<?php class C {} var_dump(class_exists('C'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C {} var_dump(class_exists('c'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(class_exists('c'));`, "bool(false)\n")

	// get_class
	testInputOutput(t, `<?php class Foo {} $bar = new Foo; var_dump(get_class($bar));`, "string(3) \"Foo\"\n")
	testInputOutput(t, `<?php namespace Space; class Foo {} $bar = new Foo; var_dump(get_class($bar));`, "string(9) \"Space\\Foo\"\n")

	// get_class_methods
	testInputOutput(t, `<?php var_dump(get_class_methods('stdclass'));`, "array(0) {\n}\n")
	testInputOutput(t,
		`<?php
			class C { function myfunc1() { } function __construct() { } function myfunc2() { } }
			var_dump(get_class_methods('C'));
			$c = new C; var_dump(get_class_methods($c));
		`,
		"array(3) {\n  [0]=>\n  string(7) \"myfunc1\"\n  [1]=>\n  string(11) \"__construct\"\n  [2]=>\n  string(7) \"myfunc2\"\n}\n"+
			"array(3) {\n  [0]=>\n  string(7) \"myfunc1\"\n  [1]=>\n  string(11) \"__construct\"\n  [2]=>\n  string(7) \"myfunc2\"\n}\n",
	)

	// get_class_vars
	testInputOutput(t, `<?php var_dump(get_class_vars('stdclass'));`, "array(0) {\n}\n")
	testInputOutput(t, `<?php
		class C {
			public $e = 'hi';
			public $d = 42;
			private $c = 4.5;
			protected $b = 'str';
			public $a;
		}
		var_dump(get_class_vars('c'));
		`,
		"array(3) {\n  [\"e\"]=>\n  string(2) \"hi\"\n  [\"d\"]=>\n  int(42)\n  [\"a\"]=>\n  NULL\n}\n",
	)

	// get_parent_class
	testInputOutput(t, `<?php class Dad {} class Child extends Dad {} $c = new Child; var_dump(get_parent_class($c));`, "string(3) \"Dad\"\n")
	testInputOutput(t, `<?php namespace Space; class Dad {} class Child extends Dad {} $c = new Child; var_dump(get_parent_class($c));`, "string(9) \"Space\\Dad\"\n")

	// is_a
	testInputOutput(t, `<?php var_dump(is_a('stdclass', 'StdClass'));`, "bool(false)\n")
	testInputOutput(t, `<?php $c = new StdClass; var_dump(is_a($c, 'StdClass'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(is_a('stdclass', 'StdClass', true));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(is_a('stdclass', 'StdClas', true));`, "bool(false)\n")
	testInputOutput(t, `<?php class C {} $c = new C; var_dump(is_a($c, 'c'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C {} var_dump(is_a('C', 'c', true));`, "bool(true)\n")
	testInputOutput(t, `<?php namespace Space; class C {} $c = new C; var_dump(is_a($c, 'c'));`, "bool(false)\n")
	testInputOutput(t, `<?php namespace Space; class C {} $c = new C; var_dump(is_a($c, 'Space\C'));`, "bool(true)\n")
	testInputOutput(t, `<?php namespace Space; class C {} var_dump(is_a('c', 'c', true));`, "bool(false)\n")
	testInputOutput(t, `<?php namespace Space; class C {} var_dump(is_a('Space\c', 'Space\C', true));`, "bool(true)\n")

	// is_subclass_of
	testInputOutput(t, `<?php class Dad {} class Child extends Dad {} $c = new Child; var_dump(is_subclass_of($c, 'Dad'));`, "bool(true)\n")
	testInputOutput(t, `<?php class Dad {} class Child extends Dad {} $c = new Child; var_dump(is_subclass_of($c, 'dad'));`, "bool(true)\n")
	testInputOutput(t, `<?php class Dad {} class Child extends Dad {} var_dump(is_subclass_of('Child', 'dad'));`, "bool(true)\n")
	testInputOutput(t, `<?php class Dad {} class Child extends Dad {} var_dump(is_subclass_of('someChild', 'Dad'));`, "bool(false)\n")
	testInputOutput(t, `<?php namespace Space; class Dad {} class Child extends Dad {} var_dump(is_subclass_of('someChild', 'Dad'));`, "bool(false)\n")
	testInputOutput(t, `<?php namespace Space; class Dad {} class Child extends Dad {} $c = new Child; var_dump(is_subclass_of('Space\Child', 'space\dad'));`, "bool(true)\n")

	// method_exists
	testInputOutput(t, `<?php var_dump(method_exists('NonexistingClass', 'read'));`, "bool(false)\n")
	testInputOutput(t, `<?php class C { function f() { } } $c = new C; var_dump(method_exists($c, 'f'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C { function f() { } } $c = new C; var_dump(method_exists($c, 'F'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C { function f() { } } $c = new C; var_dump(method_exists($c, 'g'));`, "bool(false)\n")
	testInputOutput(t, `<?php class C { function f() { } } var_dump(method_exists('C', 'f'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C { function f() { } } var_dump(method_exists('C', 'F'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C { function f() { } } var_dump(method_exists('C', 'g'));`, "bool(false)\n")

	// property_exists
	testInputOutput(t, `<?php var_dump(property_exists('C', 'prop'));`, "bool(false)\n")
	testInputOutput(t, `<?php class C { } var_dump(property_exists('C', 'prop'));`, "bool(false)\n")
	testInputOutput(t, `<?php class C { public $prop; } var_dump(property_exists('c', 'prop'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C { public $prop; } var_dump(property_exists('C', 'prop'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C { public $prop; } var_dump(property_exists('C', 'Prop'));`, "bool(false)\n")
	testInputOutput(t, `<?php class C { private $prop; } var_dump(property_exists('C', 'prop'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C { static private $prop; } var_dump(property_exists('C', 'prop'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C { protected $prop; } var_dump(property_exists('C', 'prop'));`, "bool(true)\n")
}

// -------------------------------------- array -------------------------------------- MARK: array

func TestLibArray(t *testing.T) {
	// array_flip
	testInputOutput(t, `<?= serialize(array_flip(["oranges", "apples", "pears"]));`, `a:3:{s:7:"oranges";i:0;s:6:"apples";i:1;s:5:"pears";i:2;}`)
	testInputOutput(t, `<?= serialize(array_flip(["a" => 1, "b" => 1, "c" => 2]));`, `a:2:{i:1;s:1:"b";i:2;s:1:"c";}`)

	// array_keys
	testInputOutput(t, `<?= serialize(array_keys([0 => 100, "colors" => "red"]));`, `a:2:{i:0;i:0;i:1;s:6:"colors";}`)
	testInputOutput(t, `<?= serialize(array_keys(["blue", "red", "green", "blue", "blue"]));`, `a:5:{i:0;i:0;i:1;i:1;i:2;i:2;i:3;i:3;i:4;i:4;}`)
	testInputOutput(t, `<?= serialize(array_keys(["color" => ["blue", "red"], "size" => ["small", "medium"]]));`, `a:2:{i:0;s:5:"color";i:1;s:4:"size";}`)

	// array_rand
	testForError(t, `<?php $a = []; array_rand($a);`, phpError.NewError("Uncaught ValueError: array_rand(): Argument #1 ($array) must not be empty in %s:1:16", TEST_FILE_NAME))
	testForError(t, `<?php $a = [1]; array_rand($a, 0);`, phpError.NewError("Uncaught ValueError: array_rand(): Argument #2 ($num) must be between 1 and the number of elements in argument #1 ($array) in %s:1:17", TEST_FILE_NAME))
	testForError(t, `<?php $a = [1]; array_rand($a, 2);`, phpError.NewError("Uncaught ValueError: array_rand(): Argument #2 ($num) must be between 1 and the number of elements in argument #1 ($array) in %s:1:17", TEST_FILE_NAME))
	testInputOutput(t, `<?php $a = [1,2,3]; var_dump(array_rand($a, 3));`, "array(3) {\n  [0]=>\n  int(0)\n  [1]=>\n  int(1)\n  [2]=>\n  int(2)\n}\n")
	testInputOutput(t, `<?php $a = [1,2,3]; echo gettype(array_rand($a, 1));`, "integer")
	testInputOutput(t, `<?php $a = ["a" => 1, "b" => 2, "c" => 3]; echo gettype(array_rand($a, 1));`, "string")
}

// -------------------------------------- directory -------------------------------------- MARK: directory

func TestLibDirectory(t *testing.T) {
	// getcwd
	testInputOutput(t, `<?php echo getcwd();`, TEST_FILE_PATH)
}

// -------------------------------------- misc -------------------------------------- MARK: misc

func TestLibMisc(t *testing.T) {
	// constant
	testInputOutput(t, `<?php var_dump(constant('E_ALL'));`, "int(32767)\n")
	testForError(t, `<?php constant('NOT_DEFINED_CONSTANT');`, phpError.NewError(`Undefined constant "NOT_DEFINED_CONSTANT"`))
	// TODO Add test cases for user defined constants

	// defined
	testInputOutput(t, `<?php var_dump(defined('PHP_VERSION'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(defined('NOT_DEFINED_CONSTANT'));`, "bool(false)\n")
	// TODO Add test cases for user defined constants

	// highlight_string
	testInputOutput(t, `<?php highlight_string('');`, `<pre><code style="color: #000000"></code></pre>`)
	testInputOutput(t, `<?php highlight_string('abc');`, `<pre><code style="color: #000000">abc</code></pre>`)
	// - custom colors
	devIni := ini.NewDevIni()
	devIni.Set("highlight.html", "#101010", ini.INI_ALL)
	testInputOutputCustomIni(t, devIni, `<?php highlight_string('');`, `<pre><code style="color: #101010"></code></pre>`)
}

// -------------------------------------- variable handling -------------------------------------- MARK: variable handling

func TestLibVariableHandling(t *testing.T) {
	// boolval
	testInputOutput(t, `<?php var_dump(boolval('..9'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(boolval('.9.'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(boolval('9..'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(boolval('9'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(boolval('9.9'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(boolval('9.9.9'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(boolval('9X'));`, "bool(true)\n")
	testInputOutput(t, `<?php class C {} $c = new C; var_dump(boolval($c));`, "bool(true)\n")

	// get_debug_type
	testInputOutput(t, `<?php echo get_debug_type(false);`, "bool")
	testInputOutput(t, `<?php echo get_debug_type(true);`, "bool")
	testInputOutput(t, `<?php echo get_debug_type(0);`, "int")
	testInputOutput(t, `<?php echo get_debug_type(-1);`, "int")
	testInputOutput(t, `<?php echo get_debug_type(42);`, "int")
	testInputOutput(t, `<?php echo get_debug_type(0.0);`, "float")
	testInputOutput(t, `<?php echo get_debug_type(-1.5);`, "float")
	testInputOutput(t, `<?php echo get_debug_type(42.5);`, "float")
	testInputOutput(t, `<?php echo get_debug_type("");`, "string")
	testInputOutput(t, `<?php echo get_debug_type("abc");`, "string")
	testInputOutput(t, `<?php echo get_debug_type([]);`, "array")
	testInputOutput(t, `<?php echo get_debug_type([42]);`, "array")
	testInputOutput(t, `<?php echo get_debug_type(null);`, "null")

	// gettype
	testInputOutput(t, `<?php echo gettype(false);`, "boolean")
	testInputOutput(t, `<?php echo gettype(true);`, "boolean")
	testInputOutput(t, `<?php echo gettype(0);`, "integer")
	testInputOutput(t, `<?php echo gettype(-1);`, "integer")
	testInputOutput(t, `<?php echo gettype(42);`, "integer")
	testInputOutput(t, `<?php echo gettype(0.0);`, "double")
	testInputOutput(t, `<?php echo gettype(-1.5);`, "double")
	testInputOutput(t, `<?php echo gettype(42.5);`, "double")
	testInputOutput(t, `<?php echo gettype("");`, "string")
	testInputOutput(t, `<?php echo gettype("abc");`, "string")
	testInputOutput(t, `<?php echo gettype([]);`, "array")
	testInputOutput(t, `<?php echo gettype([42]);`, "array")
	testInputOutput(t, `<?php echo gettype(null);`, "NULL")

	// is_array
	testInputOutput(t, `<?php $a = [true]; var_dump(is_array($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = 0; var_dump(is_array($a));`, "bool(false)\n")

	// is_bool
	testInputOutput(t, `<?php $a = true; var_dump(is_bool($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = 0; var_dump(is_bool($a));`, "bool(false)\n")

	// is_float
	testInputOutput(t, `<?php $a = 42.0; var_dump(is_float($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = 0; var_dump(is_float($a));`, "bool(false)\n")

	// is_int
	testInputOutput(t, `<?php $a = 42; var_dump(is_int($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = "42"; var_dump(is_int($a));`, "bool(false)\n")

	// is_null
	testInputOutput(t, `<?php $a = null; var_dump(is_null($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = 42; var_dump(is_null($a));`, "bool(false)\n")

	// is_scalar
	testInputOutput(t, `<?php $a = true; var_dump(is_scalar($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = false; var_dump(is_scalar($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = 42; var_dump(is_scalar($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = 3.5; var_dump(is_scalar($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = "abc"; var_dump(is_scalar($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = null; var_dump(is_scalar($a));`, "bool(false)\n")
	testInputOutput(t, `<?php $a = []; var_dump(is_scalar($a));`, "bool(false)\n")

	// is_string
	testInputOutput(t, `<?php $a = " "; var_dump(is_string($a));`, "bool(true)\n")
	testInputOutput(t, `<?php $a = 42; var_dump(is_string($a));`, "bool(false)\n")

	// is_object
	testInputOutput(t, `<?php $a = null; var_dump(is_object($a));`, "bool(false)\n")
	testInputOutput(t, `<?php $a = []; var_dump(is_object($a));`, "bool(false)\n")
	testInputOutput(t, `<?php class C {} $a = new C; var_dump(is_object($a));`, "bool(true)\n")

	// print_r
	testInputOutput(t, `<?php print_r(3.5);`, "3.5")
	testInputOutput(t, `<?php print_r(42);`, "42")
	testInputOutput(t, `<?php print_r("abc");`, "abc")
	testInputOutput(t, `<?php print_r(true);`, "1")
	testInputOutput(t, `<?php print_r(false);`, "")
	testInputOutput(t, `<?php print_r(null);`, "")
	testInputOutput(t, `<?php print_r([]);`, "Array\n(\n)\n")
	testInputOutput(t, `<?php print_r([1,2]);`, "Array\n(\n    [0] => 1\n    [1] => 2\n)\n")
	testInputOutput(t, `<?php print_r([1, [1]]);`, "Array\n(\n    [0] => 1\n    [1] => Array\n        (\n            [0] => 1\n        )\n\n)\n")

	// serialize
	testInputOutput(t, `<?= serialize(null);`, "N;")
	testInputOutput(t, `<?= serialize(false);`, "b:0;")
	testInputOutput(t, `<?= serialize(true);`, "b:1;")
	testInputOutput(t, `<?= serialize(3);`, "i:3;")
	testInputOutput(t, `<?= serialize(-12);`, "i:-12;")
	testInputOutput(t, `<?= serialize(3.5);`, "d:3.5;")
	testInputOutput(t, `<?= serialize("abc");`, `s:3:"abc";`)
	testInputOutput(t, `<?= serialize("abc");`, `s:3:"abc";`)
	testInputOutput(t, `<?= serialize('abc";');`, `s:5:"abc";";`)
	// - Array
	testInputOutput(t, `<?= serialize([]);`, "a:0:{}")
	testInputOutput(t, `<?= serialize([42]);`, "a:1:{i:0;i:42;}")
	testInputOutput(t, `<?= serialize(["a" => 1, 2 => "b"]);`, `a:2:{s:1:"a";i:1;i:2;s:1:"b";}`)
	testInputOutput(t, `<?= serialize([[], 2]);`, `a:2:{i:0;a:0:{}i:1;i:2;}`)
	// - Object
	testInputOutput(t, `<?php class C {}; $c = new C(); echo serialize($c);`, `O:1:"C":0:{}`)
	testInputOutput(t, `<?php namespace N; class C {}; $c = new C(); echo serialize($c);`, `O:3:"N\C":0:{}`)
	testInputOutput(t, `<?php class C { private $c; protected $b; public $a; }; $c = new C(); echo serialize($c);`, "O:1:\"C\":3:{s:4:\"\x00C\x00c\";N;s:4:\"\x00*\x00b\";N;s:1:\"a\";N;}")

	// unserialize
	testInputOutput(t, `<?php var_dump(unserialize('N;'));`, "NULL\n")
	testInputOutput(t, `<?php var_dump(unserialize('b:1;'));`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(unserialize('b:0;'));`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(unserialize('i:3;'));`, "int(3)\n")
	testInputOutput(t, `<?php var_dump(unserialize('i:42;'));`, "int(42)\n")
	testInputOutput(t, `<?php var_dump(unserialize('i:-42;'));`, "int(-42)\n")

	// var_dump
	testInputOutput(t, `<?php var_dump("str");`, "string(3) \"str\"\n")
	testInputOutput(t, `<?php var_dump(3);`, "int(3)\n")
	testInputOutput(t, `<?php var_dump(3.5);`, "float(3.5)\n")
	testInputOutput(t, `<?php var_dump(3.5, 42, true, false, null);`, "float(3.5)\nint(42)\nbool(true)\nbool(false)\nNULL\n")
	testInputOutput(t, `<?php var_dump([]);`, "array(0) {\n}\n")
	testInputOutput(t, `<?php var_dump([1,2]);`, "array(2) {\n  [0]=>\n  int(1)\n  [1]=>\n  int(2)\n}\n")
	testInputOutput(t, `<?php var_dump([1, [1]]);`, "array(2) {\n  [0]=>\n  int(1)\n  [1]=>\n  array(1) {\n    [0]=>\n    int(1)\n  }\n}\n")
	testInputOutput(t, `<?php class C {}; $c = new C; var_dump($c);`, "object(C)#1 (0) {\n}\n")
	testInputOutput(t, `<?php class C { private $p;}; $c = new C; var_dump($c);`, "object(C)#1 (1) {\n  [\"p\":\"C\":private]=>\n  NULL\n}\n")
	testInputOutput(t, `<?php namespace Space; class C {}; $c = new C; var_dump($c);`, "object(Space\\C)#1 (0) {\n}\n")

	// var_export
	testInputOutput(t, `<?php var_export(3.5);`, "3.5")
	testInputOutput(t, `<?php var_export(42);`, "42")
	testInputOutput(t, `<?php var_export("abc");`, "'abc'")
	testInputOutput(t, `<?php var_export(true);`, "true")
	testInputOutput(t, `<?php var_export(false);`, "false")
	testInputOutput(t, `<?php var_export(null);`, "NULL")
	testInputOutput(t, `<?php var_export([]);`, "array (\n)")
	testInputOutput(t, `<?php var_export([1,2]);`, "array (\n  0 => 1,\n  1 => 2,\n)")
	testInputOutput(t, `<?php var_export([1, [1]]);`, "array (\n  0 => 1,\n  1 => \n  array (\n    0 => 1,\n  ),\n)")
}
