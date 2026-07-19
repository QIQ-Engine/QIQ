package interpreter

import (
	"QIQ/cmd/qiq/ast"
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/common/os"
	"QIQ/cmd/qiq/ini"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/request"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/values"
	"fmt"
	"testing"
)

var TEST_FILE_NAME string = getTestFileName()
var TEST_FILE_PATH string = common.GetCurrentFilePath(0)

func getTestFileName() string {
	if os.IS_WIN {
		return common.GetCurrentFilePath(0) + `\test.php`
	} else {
		return common.GetCurrentFilePath(0) + "/test.php"
	}
}

// -------------------------------------- function tests -------------------------------------- MARK: function tests

func TestVariableExprToVariableName(t *testing.T) {
	// simple-variable-expression

	// $var
	interpreter, err := NewInterpreter(runtime.NewExecutionContext(), ini.NewDevIni(), &request.Request{}, "test.php")
	if err != nil {
		t.Errorf(`Unexpected error: "%s"`, err)
		return
	}
	actual, err := interpreter.varExprToVarName(ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$var")), interpreter.env)
	if err != nil {
		t.Errorf(`Unexpected error: "%s"`, err)
		return
	}
	expected := "$var"
	if actual != expected {
		t.Errorf(`Expected: "%s", Got "%s"`, expected, actual)
	}

	// $$var
	interpreter, err = NewInterpreter(runtime.NewExecutionContext(), ini.NewDevIni(), &request.Request{}, "test.php")
	if err != nil {
		t.Errorf(`Unexpected error: "%s"`, err)
		return
	}
	interpreter.env.declareVariable("$var", values.NewStr("hi"))
	actual, err = interpreter.varExprToVarName(
		ast.NewSimpleVariableExpr(0,
			ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$var"))), interpreter.env)
	if err != nil {
		t.Errorf(`Unexpected error: "%s"`, err)
		return
	}
	expected = "$hi"
	if actual != expected {
		t.Errorf(`Expected: "%s", Got "%s"`, expected, actual)
	}

	// $$$var
	interpreter, err = NewInterpreter(runtime.NewExecutionContext(), ini.NewDevIni(), &request.Request{}, "test.php")
	if err != nil {
		t.Errorf(`Unexpected error: "%s"`, err)
		return
	}
	interpreter.env.declareVariable("$var1", values.NewStr("hi"))
	interpreter.env.declareVariable("$var", values.NewStr("var1"))
	actual, err = interpreter.varExprToVarName(
		ast.NewSimpleVariableExpr(0,
			ast.NewSimpleVariableExpr(0,
				ast.NewSimpleVariableExpr(0, ast.NewVariableNameExpr(0, nil, "$var")))), interpreter.env)
	if err != nil {
		t.Errorf(`Unexpected error: "%s"`, err)
		return
	}
	expected = "$hi"
	if actual != expected {
		t.Errorf(`Expected: "%s", Got "%s"`, expected, actual)
	}

	// ${$a}
	testInputOutput(t, `<?php $b = "bb";$bb = "baa";$c = ${$b};var_dump($c);`, "string(3) \"baa\"\n")
	testInputOutput(t, `<?php $b = "bb";$a = "aa";$bb = "baa";$c = ${$a=$b};var_dump($c);`, "string(3) \"baa\"\n")
}

// -------------------------------------- input output tests -------------------------------------- MARK: input output tests

func testForErrorCustomIni(t *testing.T, ini *ini.Ini, php string, expected phpError.Error) {
	interpreter, err := NewInterpreter(runtime.NewExecutionContext(), ini, &request.Request{}, TEST_FILE_NAME)
	if err != nil {
		t.Errorf("\nCode: \"%s\"\nUnexpected error: \"%s\"", php, err)
		return
	}
	_, err = interpreter.Process(php)
	if err == nil {
		t.Errorf("\nCode: \"%s\"\nExpected: %s\nGot:      nil", php, expected)
		return
	}
	if err.GetErrorType() != expected.GetErrorType() || err.GetMessage() != expected.GetMessage() {
		t.Errorf("\nCode: \"%s\"\nExpected: %s\nGot:      %s", php, expected, err)
	}
}

func testForError(t *testing.T, php string, expected phpError.Error) {
	testForErrorCustomIni(t, ini.NewDevIni(), php, expected)
}

func testInputOutputCustomIni(t *testing.T, ini *ini.Ini, php string, output string) *Interpreter {
	// Always use "\n" for tests so that they also pass on Windows
	os.EOL = "\n"
	interpreter, err := NewInterpreter(runtime.NewExecutionContext(), ini, &request.Request{}, TEST_FILE_NAME)
	if err != nil {
		t.Errorf("\nCode: \"%s\"\nUnexpected error: \"%s\"", php, err)
		return interpreter
	}
	actual, err := interpreter.Process(php)
	if err != nil {
		t.Errorf("\nCode: \"%s\"\nUnexpected error: \"%s\"", php, err)
		return interpreter
	}
	if actual != output {
		t.Errorf("\nCode: \"%s\"\nExpected: \"%s\",\nGot:      \"%s\"", php, output, actual)
	}
	return interpreter
}

func testInputOutput(t *testing.T, php string, output string) *Interpreter {
	return testInputOutputCustomIni(t, ini.NewDevIni(), php, output)
}

func TestOutput(t *testing.T) {
	// Empty string
	testInputOutput(t, "", "")

	// Without PHP
	testInputOutput(t, "<html>...</html>", "<html>...</html>")

	// Echo short tag
	testInputOutput(t, `<html><?= "abc\n" ?><?= 42; ?></html>`, "<html>abc\n42</html>")
	// Echo
	testInputOutput(t,
		`<html><?php echo "abc\t", 42 ?><?php echo "def", 24; ?></html>`, "<html>abc\t42def24</html>",
	)

	// Simple variable substitution
	testInputOutput(t, `<?php $a = 42; echo "a{$a}b";`, "a42b")
	testInputOutput(t, `<?php $a = 42; $b = 'abc'; echo "a{$a}b{$b}";`, "a42babc")
	testInputOutput(t, `<?php $a = 42; echo "$a";`, "42")
	testInputOutput(t, `<?php $a = 42; $b = 'abc'; echo "$a $b";`, "42 abc")
	testInputOutput(t, `<?php echo "$a";`, fmt.Sprintf("\nWarning: Undefined variable $a in %s:1:12\n", TEST_FILE_NAME))

	// Unicode escape sequence
	testInputOutput(t, `<?php var_dump("\u{61}");`, "string(1) \"a\"\n")
	testInputOutput(t, `<?php var_dump("\u{FF}");`, "string(2) \"ÿ\"\n")
	testInputOutput(t, `<?php var_dump("\u{ff}");`, "string(2) \"ÿ\"\n")
	testInputOutput(t, `<?php var_dump("\u{2603}");`, "string(3) \"☃\"\n")
	testInputOutput(t, `<?php var_dump("\u{1F602}");`, "string(4) \"😂\"\n")
	testInputOutput(t, `<?php var_dump("\u{0000001F602}");`, "string(4) \"😂\"\n")

	// Print
	// From https://www.php.net/manual/en/function.print.php
	testInputOutput(t, `<?php print "hello";print "world";print "\n";`, "helloworld\n")
	testInputOutput(t, `<?php print 6*7;`, "42")
	testInputOutput(t, `<?php $foo = "example"; print "foo is {$foo}";`, "foo is example")
	testInputOutput(t, `<?php if ((print "hello") === 1) { echo " y"; } else { echo " n"; }`, "hello y")
	testInputOutput(t, `<?php ( 1 === 1 ) ? print 'true' : print 'false';`, "true")
	testInputOutput(t, `<?php print("hello");`, "hello")
	testInputOutput(t, `<?php print(1 + 2) * 3;`, "9")
}

func TestTryStmt(t *testing.T) {
	testInputOutput(t, `<?php try { echo "try."; } finally { echo "finally."; }`, "try.finally.")

	testForError(t, `<?php try {} finally { f1(); }`, phpError.NewError("Call to undefined function f1() in %s:1:24", TEST_FILE_NAME))
}

func TestConstants(t *testing.T) {
	// Predefined constants
	testInputOutput(t, `<?php echo E_USER_NOTICE;`, fmt.Sprintf("%d", phpError.E_USER_NOTICE))
	testInputOutput(t, `<?php echo E_ALL;`, fmt.Sprintf("%d", phpError.E_ALL))

	// Magic constants
	testInputOutput(t, "<?php var_dump(__DIR__);", fmt.Sprintf("string(%d) \"%s\"\n", len(TEST_FILE_PATH), TEST_FILE_PATH))
	testInputOutput(t, "<?php var_dump(__FUNCTION__); function func() { var_dump(__FUNCTION__); } func();", "string(0) \"\"\nstring(4) \"func\"\n")
	testInputOutput(t, "<?php var_dump(__METHOD__); function func() { var_dump(__METHOD__); } func();", "string(0) \"\"\nstring(4) \"func\"\n")
	testInputOutput(t, "<?php var_dump(__FILE__);", fmt.Sprintf("string(%d) \"%s\"\n", len(TEST_FILE_NAME), TEST_FILE_NAME))
	testInputOutput(t, "<?php var_dump(__LINE__);\nvar_dump(__LINE__);", "int(1)\nint(2)\n")
	testInputOutput(t, "<?php var_dump(__CLASS__); class c { function __construct() { var_dump(__CLASS__); } }; new c;", "string(0) \"\"\nstring(1) \"c\"\n")
	testInputOutput(t, "<?php class c { function __construct() { var_dump(__FUNCTION__); var_dump(__METHOD__); } }; new c;", "string(11) \"__construct\"\nstring(14) \"c::__construct\"\n")
	testInputOutput(t, "<?php namespace Space; class c { function __construct() { var_dump(__FUNCTION__); var_dump(__METHOD__); } }; new c;", "string(11) \"__construct\"\nstring(20) \"Space\\c::__construct\"\n")

	// Userdefined constants
	testInputOutput(t, `<?php const TRUTH = 42; const PI = "3.141";echo TRUTH, PI;`, "423.141")

	// Already defined constants
	testForError(t, `<?php const M_PI = 3;`, phpError.NewWarning("Constant M_PI already defined in %s:1:7", TEST_FILE_NAME))
}

func TestFileIncludes(t *testing.T) {
	testInputOutput(t, `<?php include "../../../test/include.php"; includedFunc();`, "includedFunc")
	if os.IS_WIN {
		testInputOutput(t, `<?php include "..\\..\\..\\test\\include.php"; includedFunc();`, "includedFunc")
	}

	testForError(t, `<?php require "include.php";`,
		phpError.NewError("Uncaught Error: Failed opening required 'include.php' (include_path='%s') in %s:1:15", TEST_FILE_PATH, TEST_FILE_NAME),
	)
	testForError(t, `<?php require_once "include.php";`,
		phpError.NewError("Uncaught Error: Failed opening required 'include.php' (include_path='%s') in %s:1:20", TEST_FILE_PATH, TEST_FILE_NAME),
	)
	testForError(t, `<?php include "include.php";`,
		phpError.NewWarning("include(): Failed opening 'include.php' for inclusion (include_path='%s') in %s:1:15", TEST_FILE_PATH, TEST_FILE_NAME),
	)
	testForError(t, `<?php include_once "include.php";`,
		phpError.NewWarning("include(): Failed opening 'include.php' for inclusion (include_path='%s') in %s:1:20", TEST_FILE_PATH, TEST_FILE_NAME),
	)

	// QIQ feature: case-sensitive include (on Windows)
	devIni := ini.NewDevIni()
	devIni.Set("qiq.case_sensitive_include", "1", ini.INI_SYSTEM)
	testForErrorCustomIni(t, devIni,
		`<?php include "../../../test/Include.php";`,
		phpError.NewWarning("include(): Failed opening '../../../test/Include.php' for inclusion (include_path='%s') in %s:1:15", TEST_FILE_PATH, TEST_FILE_NAME),
	)
	testForErrorCustomIni(t, devIni,
		`<?php include "../../../Test/Include.php";`,
		phpError.NewWarning("include(): Failed opening '../../../Test/Include.php' for inclusion (include_path='%s') in %s:1:15", TEST_FILE_PATH, TEST_FILE_NAME),
	)
}

func TestVariable(t *testing.T) {
	// Undefined variable
	testInputOutput(t, `<?php echo is_null($a) ? "a" : "b";`, fmt.Sprintf("\nWarning: Undefined variable $a in %s:1:20\na", TEST_FILE_NAME))
	testInputOutput(t, `<?php echo intval($a);`, fmt.Sprintf("\nWarning: Undefined variable $a in %s:1:19\n0", TEST_FILE_NAME))
	testInputOutput(t, `<?php echo intval($$a);`, fmt.Sprintf("\nWarning: Undefined variable $a\n\nWarning: Undefined variable $ in %s:1:20\n0", TEST_FILE_NAME))

	// Simple variable
	testInputOutput(t, `<?php $var = "hi"; $var = "hello"; echo $var, " world";`, "hello world")

	// Variable variable name
	testInputOutput(t, `<?php $var = "hi"; $$var = "hello"; echo $hi, " world";`, "hello world")

	// Chained variable declarations
	testInputOutput(t, `<?php $a = $b = $c = 42; echo $a, $b, $c;`, "424242")

	// Compound assignment
	testInputOutput(t, `<?php $a = 42; echo $a; $a += 2; echo $a; $a += $a; echo $a;`, "424488")

	// Parenthesized LHS
	testForError(t, `<?php ($a) = 42;`,
		phpError.NewParseError(`Statement must end with a semicolon. Got: "=" in %s:1:12`, TEST_FILE_NAME),
	)

	// Global declaration
	testInputOutput(t,
		`<?php
			$foo = 42;
			function f() {
				$foo = "abc";
				var_dump($foo);
				global $foo;
				var_dump($foo);
				$foo = 12;
			}
			f();
			var_dump($foo);`,
		"string(3) \"abc\"\nint(42)\nint(12)\n")
}

func TestConditionals(t *testing.T) {
	// Conditional
	testInputOutput(t, `<?php echo 1 ? "y" : "n"; echo 0 ? "n" : "y"; echo false ?: "y";`, "yyy")

	// Coalesce
	testInputOutput(t,
		`<?php $a = null; echo $a ?? "a"; $a = "b"; echo $a ?? "a"; echo "c" ?? "d";`, "abc",
	)
	testInputOutput(t, `<?php echo $a ?? "a";`, "a")

	// If statement
	testInputOutput(t, `<?php $a = 42; if ($a === 42) { echo "42"; } elseif ($a === 41) { echo "41"; } else { echo "??"; }`, "42")
	testInputOutput(t, `<?php $a = 41; if ($a === 42) { echo "42"; } elseif ($a === 41) { echo "41"; } else { echo "??"; }`, "41")
	testInputOutput(t, `<?php $a = 40; if ($a === 42) { echo "42"; } elseif ($a === 41) { echo "41"; } else { echo "??"; }`, "??")
	// Alternative syntax
	testInputOutput(t, `<?php $a = 42; if ($a === 42): echo "42"; elseif ($a === 41): echo "41"; else: echo "??"; endif;`, "42")
	testInputOutput(t, `<?php $a = 41; if ($a === 42): echo "42"; elseif ($a === 41): echo "41"; else: echo "??"; endif;`, "41")
	testInputOutput(t, `<?php $a = 40; if ($a === 42): echo "42"; elseif ($a === 41): echo "41"; else: echo "??"; endif;`, "??")
	// If statement mixed with text expressions
	testInputOutput(t, `<?php if (true): ?>a<?= 'b' ?>c<?php endif ?>`, "abc")

	// While statement
	testInputOutput(t, `<?php $a = 40; while ($a < 42) { echo "1"; $a++; }`, "11")
	testInputOutput(t, `<?php $a = 42; while ($a < 42) { echo "1"; $a++; }`, "")
	testInputOutput(t, `<?php $a = 0; while (true) { echo "1"; $a++; if ($a == 5) { break; }}`, "11111")
	testInputOutput(t, `<?php $a = 0; while (true) { echo "1"; while (true) { echo "2"; break 2; }}`, "12")
	// Alternative syntax
	testInputOutput(t, `<?php $a = 40; while ($a < 42): echo "1";  $a++; endwhile;`, "11")
	testInputOutput(t, `<?php $a = 42; while ($a < 42): echo "1";  $a++; endwhile;`, "")
	// While statement mixed with text expressions
	testInputOutput(t, `<?php $a = 0; while ($a < 10) { ?>.<?= $a ?>.<?php $a++; } ?>`, ".0..1..2..3..4..5..6..7..8..9.")

	// Do statement
	testInputOutput(t, `<?php $a = 40; do { echo "1"; $a++; } while ($a < 42);`, "11")
	testInputOutput(t, `<?php $a = 42; do { echo "1"; $a++; } while ($a < 42);`, "1")
	testInputOutput(t, `<?php $a = 0; do { echo "1"; $a++; if ($a == 5) { break; }} while (true);`, "11111")
	testInputOutput(t, `<?php $a = 0; do { echo "1"; while (true) { echo "2"; break 2; }} while (true);`, "12")

	// For statement
	testInputOutput(t,
		`<?php for ($i = 1; $i <= 10; ++$i) {
			echo $i." ".($i * $i)."\n"; // output a table of squares
		}`,
		"1 1\n2 4\n3 9\n4 16\n5 25\n6 36\n7 49\n8 64\n9 81\n10 100\n",
	)
	// Omit 1st and 3rd expressions
	testInputOutput(t,
		`<?php $i = 1;
		for (; $i <= 10;):
			echo $i." ".($i * $i)."\n"; // output a table of squares
			++$i;
		endfor;`,
		"1 1\n2 4\n3 9\n4 16\n5 25\n6 36\n7 49\n8 64\n9 81\n10 100\n",
	)
	// Omit all 3 expressions
	testInputOutput(t,
		`<?php $i = 1;
		for (;;) {
			if ($i > 10) break;
			echo $i." ".($i * $i)."\n"; // output a table of squares
			++$i;
		}`,
		"1 1\n2 4\n3 9\n4 16\n5 25\n6 36\n7 49\n8 64\n9 81\n10 100\n",
	)
	// Use groups of expressions
	testInputOutput(t,
		`<?php for ($a = 100, $i = 1; ++$i, $i <= 10; ++$i, $a -= 10) {
			echo $i." ".$a."\n";
		}`,
		"2 100\n4 90\n6 80\n8 70\n10 60\n",
	)

	// Foreach statement
	testInputOutput(t,
		`<?php foreach([1,2,3] as $i) { echo $i; } echo $i;`,
		"1233",
	)
	testInputOutput(t,
		`<?php foreach([2, 4, 8] as $k => $v) { echo $k . ":" . $v . ","; } echo $k . ":" . $v;`,
		"0:2,1:4,2:8,2:8",
	)
	// byRef
	testInputOutput(t,
		`<?php $a = [1,2,3]; foreach($a as &$i) { $i++; } echo implode(',',$a);`,
		"2,3,4",
	)
	// TODO foreach byRef
	// testInputOutput(t,
	// 	`<?php $a = [1,2,3]; foreach($a as &$i) { $i++; } var_dump($a);`,
	// 	"array(3) {\n  [0]=>\n  int(2)\n  [1]=>\n  int(3)\n  [2]=>\n  &int(4)\n}\n",
	// )
	// Foreach loop with object
	testInputOutput(t,
		`<?php class MyData { public $a = 1; public $b = 2; public $c = 3; }
		$data = new MyData();
		foreach ($data as $value) { echo $value . ','; }`,
		"1,2,3,",
	)
	testInputOutput(t,
		`<?php class MyData { public $a = 1; public $b = 2; public $c = 3; }
		$data = new MyData();
		foreach ($data as $key => $value) { echo $key . ' => ' . $value . ','; }`,
		"a => 1,b => 2,c => 3,",
	)
	// Foreach loop with object by ref
	testInputOutput(t,
		`<?php class MyData { public $a = 1; public $b = 2; public $c = 3; }
		$data = new MyData();
		foreach ($data as &$value) { $value *= 10; }
		print_r($data);`,
		"MyData Object\n(\n    [a] => 10\n    [b] => 20\n    [c] => 30\n)\n",
	)
}

func TestIntrinsic(t *testing.T) {
	// Exit
	interpreter := testInputOutput(t, `Hello <?php exit("world");`, "Hello world")
	if interpreter.GetResponse().ExitCode != 0 {
		t.Errorf("Expected: \"%d\", Got \"%d\"", 0, interpreter.GetResponse().ExitCode)
	}
	interpreter = testInputOutput(t, `Hello<?php exit;`, "Hello")
	if interpreter.GetResponse().ExitCode != 0 {
		t.Errorf("Expected: \"%d\", Got \"%d\"", 0, interpreter.GetResponse().ExitCode)
	}
	interpreter = testInputOutput(t, `Hello<?php exit(42);`, "Hello")
	if interpreter.GetResponse().ExitCode != 42 {
		t.Errorf("Expected: \"%d\", Got \"%d\"", 42, interpreter.GetResponse().ExitCode)
	}

	// Empty
	testInputOutput(t, `<?php echo empty(false) ? "y" : "n";`, "y")
	testInputOutput(t, `<?php echo empty(true) ? "y" : "n";`, "n")
	testInputOutput(t, `<?php echo empty(0) ? "y" : "n";`, "y")
	testInputOutput(t, `<?php echo empty(1) ? "y" : "n";`, "n")
	testInputOutput(t, `<?php echo empty(0.0) ? "y" : "n";`, "y")
	testInputOutput(t, `<?php echo empty(2.0) ? "y" : "n";`, "n")
	testInputOutput(t, `<?php echo empty("") ? "y" : "n";`, "y")
	testInputOutput(t, `<?php echo empty("0") ? "y" : "n";`, "y")
	testInputOutput(t, `<?php echo empty("1") ? "y" : "n";`, "n")
	testInputOutput(t, `<?php echo empty("00") ? "y" : "n";`, "n")
	testInputOutput(t, `<?php echo empty(null) ? "y" : "n";`, "y")
	testInputOutput(t, `<?php echo empty($a) ? "y" : "n";`, "y")
	testInputOutput(t, `<?php $a = 1; echo empty($a) ? "y" : "n";`, "n")

	// Eval
	testInputOutput(t, `<?php var_dump(eval("return 12;"));`, "int(12)\n")
	testInputOutput(t, `<?php var_dump(eval("12;"));`, "NULL\n")

	// Isset
	testInputOutput(t, `<?php $a = 1; echo isset($a) ? "y" : "n";`, "y")
	testInputOutput(t, `<?php echo isset($a) ? "y" : "n";`, "n")
	testInputOutput(t, `<?php $a = 1; echo isset($a, $b) ? "y" : "n";`, "n")
	testInputOutput(t, `<?php echo isset($a, $b) ? "y" : "n";`, "n")
	testInputOutput(t,
		`<?php $a = ['foo' => 'bar'];
			var_dump(isset($a));
			var_dump(isset($a['foo']));
			var_dump(isset($a['xyz']));
			var_dump(isset($a['foo']['xyz']));`,
		"bool(true)\nbool(true)\nbool(false)\nbool(false)\n",
	)

	// Unset
	testInputOutput(t, `<?php $a = 1; echo isset($a) ? "y" : "n"; unset($a); echo isset($a) ? "y" : "n";`, "yn")
	testInputOutput(t, `<?php echo isset($a) ? "y" : "n"; unset($a); echo isset($a) ? "y" : "n";`, "nn")
}

func TestString(t *testing.T) {
	// Heredoc string
	testInputOutput(t, "<?php $v = 123; $s = <<< ID\n"+`S'o'me "\"t e\txt; v = $v"`+"\nSome more text\nID; echo \">$s<\";", `>S'o'me "\"t e`+"\t"+`xt; v = 123"`+"\nSome more text<")

	// Read string index
	testInputOutput(t, `<?php $s = 'abc'; var_dump($s[0]);`, "string(1) \"a\"\n")
	testForError(t, `<?php $s = 'abc'; var_dump($s[""]);`, phpError.NewError("Cannot access offset of type string on string in %s:1:31", TEST_FILE_NAME))
	testForError(t, `<?php $s = 'abc'; var_dump($s[]);`, phpError.NewError("Cannot use [] for reading in %s:1:28", TEST_FILE_NAME))
	testForError(t, `<?php $s = 'abc'; var_dump($s[3]);`, phpError.NewError("Uninitialized string offset 3 in %s:1:31", TEST_FILE_NAME))

	// Write string index
	testInputOutput(t, `<?php $s = '123'; $s[0] = '*'; var_dump($s);`, "string(3) \"*23\"\n")
	testInputOutput(t, `<?php $s = '123'; $s[0] = '**'; var_dump($s);`, "string(3) \"*23\"\n")
	testInputOutput(t, `<?php $s = '123'; $s[3] = '**'; var_dump($s);`, "string(4) \"123*\"\n")
	testInputOutput(t, `<?php $s = '123'; $s[5] = '*'; var_dump($s);`, "string(6) \"123  *\"\n")
	testInputOutput(t, `<?php $s = '12345'; $s[1] = $s[1] = $s[3]; var_dump($s);`, "string(5) \"14345\"\n")
	testInputOutput(t, `<?php $s = '12345'; $s[1] = $s[3] = '*'; var_dump($s);`, "string(5) \"1*3*5\"\n")
	testForError(t, `<?php $s = '123'; $s[] = '*';`, phpError.NewError("[] operator not supported for strings in %s:1:19", TEST_FILE_NAME))
	testForError(t, `<?php $s = '123'; $s["abc"] = '*';`, phpError.NewError("Cannot access offset of type string on string in %s:1:22", TEST_FILE_NAME))
}

func TestArray(t *testing.T) {
	testInputOutput(t, `<?php $a = [0, 1, 2]; echo $a[0] === null ? "y" : "n";`, "n")
	testInputOutput(t, `<?php $a = [0, 1, 2]; echo $a[3] === null ? "y" : "n";`, "y")
	testInputOutput(t, `<?php $a = [0, 1]; echo $a[2] = 2; echo $a[2];`, "22")
	testInputOutput(t, `<?php $a = []; $a[] = 1; echo $a[0];`, "1")
	testInputOutput(t, `<?php $a = []; $a[][] = "42"; echo $a[0][0];`, "42")
	testInputOutput(t, `<?php $a = []; $a[][123] = "42"; echo $a[0][123];`, "42")
	testInputOutput(t, `<?php $a = []; $a["a"]["b"]["c"]=1; echo $a["a"]["b"]["c"];`, "1")

	// Implicit declaration
	testInputOutput(t, `<?php $a["b"] = "c"; var_dump($a);`, "array(1) {\n  [\"b\"]=>\n  string(1) \"c\"\n}\n")

	// Determination of the next key
	testInputOutput(t,
		`<?php $a = []; $a['abc'] = 'def'; $a[] = 'ghi'; var_dump($a);`,
		"array(2) {\n  [\"abc\"]=>\n  string(3) \"def\"\n  [0]=>\n  string(3) \"ghi\"\n}\n",
	)
	testInputOutput(t,
		`<?php $a = []; $a[100] = 'def'; $a[] = 'ghi'; var_dump($a);`,
		"array(2) {\n  [100]=>\n  string(3) \"def\"\n  [101]=>\n  string(3) \"ghi\"\n}\n",
	)
	testInputOutput(t,
		`<?php $a = []; $a[100] = 'def'; $a['abc'] = 'def'; $a[] = 'ghi'; var_dump($a);`,
		"array(3) {\n  [100]=>\n  string(3) \"def\"\n  [\"abc\"]=>\n  string(3) \"def\"\n  [101]=>\n  string(3) \"ghi\"\n}\n",
	)

	// Pass by value not reference
	testInputOutput(t,
		`<?php $a = $b = [42]; var_dump($a[0], $b[0]); $b[0] = 43; var_dump($a[0], $b[0]);`,
		"int(42)\nint(42)\nint(42)\nint(43)\n",
	)
	testInputOutput(t,
		`<?php $b = [42]; $a = $b; var_dump($a[0], $b[0]); $b[0] = 43; var_dump($a[0], $b[0]);`,
		"int(42)\nint(42)\nint(42)\nint(43)\n",
	)

	// Key value
	testInputOutput(t,
		`<?php $array = [1.5 => "a"]; var_dump($array[1]);`,
		"string(1) \"a\"\n",
	)
	// Tests from https://www.php.net/manual/en/language.types.array.php
	testInputOutput(t,
		`<?php $array = [1 => "a", "1" => "b", 1.5 => "c", true => "d",]; var_dump($array);`,
		"array(1) {\n  [1]=>\n  string(1) \"d\"\n}\n",
	)
	testInputOutput(t,
		`<?php $array = ["foo" => "bar", "bar" => "foo", 100 => -100, -100 => 100,]; var_dump($array);`,
		"array(4) {\n  [\"foo\"]=>\n  string(3) \"bar\"\n  [\"bar\"]=>\n  string(3) \"foo\"\n  [100]=>\n  int(-100)\n  [-100]=>\n  int(100)\n}\n",
	)
	testInputOutput(t,
		`<?php $array = ["foo", "bar", "hello", "world"]; var_dump($array);`,
		"array(4) {\n  [0]=>\n  string(3) \"foo\"\n  [1]=>\n  string(3) \"bar\"\n  [2]=>\n  string(5) \"hello\"\n  [3]=>\n  string(5) \"world\"\n}\n",
	)
	testInputOutput(t,
		`<?php $array = ["a", "b", 6 => "c", "d"]; var_dump($array);`,
		"array(4) {\n  [0]=>\n  string(1) \"a\"\n  [1]=>\n  string(1) \"b\"\n  [6]=>\n  string(1) \"c\"\n  [7]=>\n  string(1) \"d\"\n}\n",
	)
	testInputOutput(t,
		`<?php $array = [
			1    => 'a',
			'1'  => 'b', // the value "a" will be overwritten by "b"
			1.5  => 'c', // the value "b" will be overwritten by "c"
			-1 => 'd',
			'01'  => 'e', // as this is not an integer string it will NOT override the key for 1
			'1.5' => 'f', // as this is not an integer string it will NOT override the key for 1
			true => 'g', // the value "c" will be overwritten by "g"
			false => 'h',
			'' => 'i',
			null => 'j', // the value "i" will be overwritten by "j"
			'k', // value "k" is assigned the key 2. This is because the largest integer key before that was 1
			2 => 'l', // the value "k" will be overwritten by "l"
		]; var_dump($array);`,
		"array(7) {\n  [1]=>\n  string(1) \"g\"\n  [-1]=>\n  string(1) \"d\"\n  [\"01\"]=>\n  string(1) \"e\"\n  [\"1.5\"]=>\n  string(1) \"f\"\n  [0]=>\n  string(1) \"h\"\n  [\"\"]=>\n  string(1) \"j\"\n  [2]=>\n  string(1) \"l\"\n}\n",
	)
	testInputOutput(t,
		`<?php $array = [-5 => 1, 2]; var_dump($array);`,
		"array(2) {\n  [-5]=>\n  int(1)\n  [-4]=>\n  int(2)\n}\n",
	)

	// Implode
	testInputOutput(t, `<?php $a = [1, 2, 3]; var_dump(implode($a));`, "string(5) \"1 2 3\"\n")
	testInputOutput(t, `<?php $a = [1, 2, 3]; var_dump(implode(' ', $a));`, "string(5) \"1 2 3\"\n")
	testInputOutput(t, `<?php $a = ["1", 2, 3.4]; var_dump(implode('-', $a));`, "string(7) \"1-2-3.4\"\n")

	// Reassignment
	testInputOutput(t, `<?php $a = [1];echo serialize($a);$a = [2];echo serialize($a);`, "a:1:{i:0;i:1;}a:1:{i:0;i:2;}")
}

func TestCastExpression(t *testing.T) {
	testInputOutput(t, `<?php var_dump((array)42);`, "array(1) {\n  [0]=>\n  int(42)\n}\n")
	testInputOutput(t, `<?php var_dump((binary)42);`, `string(2) "42"`+"\n")
	testInputOutput(t, `<?php var_dump((bool)42);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump((boolean)42);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump((double)42);`, "float(42)\n")
	testInputOutput(t, `<?php var_dump((int)42);`, "int(42)\n")
	testInputOutput(t, `<?php var_dump((integer)42);`, "int(42)\n")
	testInputOutput(t, `<?php var_dump((float)42);`, "float(42)\n")
	// TODO testInputOutput(t, `<?php var_dump((object)42);`, "a")
	testInputOutput(t, `<?php var_dump((real)42);`, "float(42)\n")
	testInputOutput(t, `<?php var_dump((string)42);`, `string(2) "42"`+"\n")
}

func TestNumbers(t *testing.T) {
	// Numeric Literal Separator

	// Decimal
	testInputOutput(t, `<?php var_dump(1_000_000_000);`, "int(1000000000)\n")
	testInputOutput(t, `<?php var_dump(135_00);`, "int(13500)\n")
	testInputOutput(t, `<?php var_dump(299_792_458);`, "int(299792458)\n")
	// Invalid
	testForError(t, `<?php var_dump(_100);`, phpError.NewError("Undefined constant \"_100\""))
	testForError(t, `<?php var_dump(100_);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))
	testForError(t, `<?php var_dump(1__1);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))

	// Hexadecimal
	testInputOutput(t, `<?php var_dump(0xCAFE_F00D);`, "int(3405705229)\n")
	testInputOutput(t, `<?php var_dump(0x42_72_6F_77_6E);`, "int(285387749230)\n")
	// Invalid
	testForError(t, `<?php var_dump(0x_123);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))
	testForError(t, `<?php var_dump(0x1_23_);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))
	testForError(t, `<?php var_dump(0x1__23);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))

	// Binary
	testInputOutput(t, `<?php var_dump(0b0101_1111);`, "int(95)\n")
	testInputOutput(t, `<?php var_dump(0b01010100_01101000_01100101_01101111);`, "int(1416127855)\n")
	// Invalid
	testForError(t, `<?php var_dump(0b_101);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))
	testForError(t, `<?php var_dump(0b1_1_);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))
	testForError(t, `<?php var_dump(0b1__1);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))

	// Octal
	testInputOutput(t, `<?php var_dump(0137_041);`, "int(48673)\n")
	testInputOutput(t, `<?php var_dump(0_101);`, "int(65)\n")
	// Invalid
	testForError(t, `<?php var_dump(_010);`, phpError.NewError("Undefined constant \"_010\""))
	testForError(t, `<?php var_dump(010_);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))
	testForError(t, `<?php var_dump(0__10);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))

	// Float
	testInputOutput(t, `<?php var_dump(107_925_284.88);`, "float(107925284.88)\n")
	testInputOutput(t, `<?php var_dump(6.674_083e-11);`, "float(0.00000000006674083)\n")
	// Invalid
	testForError(t, `<?php var_dump(1_.0);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))
	testForError(t, `<?php var_dump(1._0);`, phpError.NewError("Undefined constant \"_0\""))
	testForError(t, `<?php var_dump(1_e2);`, phpError.NewParseError("Unsupported number format detected in %s:1:16", TEST_FILE_NAME))
	testForError(t, `<?php var_dump(1e_2);`, phpError.NewParseError("Expected \",\" or \")\". Got: &{Token - type: NameToken, value: \"e_2\", position: {Position - file: \"%s\", ln: 1, col: 17}}", TEST_FILE_NAME))

	// Convertion
	// intval
	testInputOutput(t, `<?php var_dump(intval('..9'));`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(intval('.9.'));`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(intval('9..'));`, "int(9)\n")
	testInputOutput(t, `<?php var_dump(intval('9'));`, "int(9)\n")
	testInputOutput(t, `<?php var_dump(intval('9.9'));`, "int(9)\n")
	testInputOutput(t, `<?php var_dump(intval('9.9.9'));`, "int(9)\n")
	testInputOutput(t, `<?php var_dump(intval('9X'));`, "int(9)\n")
	// floatval
	testInputOutput(t, `<?php var_dump(floatval('..9'));`, "float(0)\n")
	testInputOutput(t, `<?php var_dump(floatval('.9.'));`, "float(0.9)\n")
	testInputOutput(t, `<?php var_dump(floatval('9..'));`, "float(9)\n")
	testInputOutput(t, `<?php var_dump(floatval('9'));`, "float(9)\n")
	testInputOutput(t, `<?php var_dump(floatval('9.9'));`, "float(9.9)\n")
	testInputOutput(t, `<?php var_dump(floatval('9.9.9'));`, "float(9.9)\n")
	testInputOutput(t, `<?php var_dump(floatval('9X'));`, "float(9)\n")
}

func TestOperators(t *testing.T) {
	// Logical "not"
	testInputOutput(t, `<?php echo !true ? "y" : "y";`, "y")
	testInputOutput(t, `<?php echo !false ? "y" : "y";`, "y")
	testInputOutput(t, `<?php echo !42 ? "y" : "y";`, "y")

	// Logical "and" and "or"
	testInputOutput(t, `<?php echo 4 && 0 ? "t" : "f";`, "f")
	testInputOutput(t, `<?php echo 4 && 1 ? "t" : "f";`, "t")
	testInputOutput(t, `<?php echo 4 && false ? "t" : "f";`, "f")
	testInputOutput(t, `<?php echo 4 && true ? "t" : "f";`, "t")
	testInputOutput(t, `<?php echo 0 || 0 ? "t" : "f";`, "f")
	testInputOutput(t, `<?php echo 4 || 1 ? "t" : "f";`, "t")
	testInputOutput(t, `<?php echo false || false ? "t" : "f";`, "f")
	testInputOutput(t, `<?php echo 4 || true ? "t" : "f";`, "t")
	// Test "short-circuiting"
	testInputOutput(t,
		`<?php function fun() { echo "fn_call "; return false; }
		echo ("true" || fun()) === true ? "true": "false"; echo "\n";
		echo (false || fun()) === true ? "true": "false"; echo "\n";
		echo (false && fun()) === true ? "true": "false"; echo "\n";
		echo ("true" && fun()) === true ? "true": "false"; echo "\n";`,
		"true\nfn_call false\nfalse\nfn_call false\n",
	)

	// Logical "xor, "and" and "or" 2
	testInputOutput(t, `<?php echo (true xor true) ? "t": "f";`, "f")
	testInputOutput(t, `<?php echo (true xor false) ? "t": "f";`, "t")
	testInputOutput(t, `<?php echo (false xor false) ? "t": "f";`, "f")
	testInputOutput(t, `<?php if (4 or 1) { echo "t"; } else { echo "f"; }`, "t")
	testInputOutput(t, `<?php echo "234" or 12 ? "3": "2";`, "1")
	testInputOutput(t, `<?php echo ("234" or 12) ? "3": "2";`, "3")
	testInputOutput(t, `<?php echo "234" and 12 ? "3": "2";`, "1")
	testInputOutput(t, `<?php echo ("234" and 12) ? "3": "2";`, "3")
	testInputOutput(t, `<?php echo (4 and 0) ? "t" : "f";`, "f")
	testInputOutput(t, `<?php echo (4 and 1) ? "t" : "f";`, "t")
	testInputOutput(t, `<?php echo (4 and false) ? "t" : "f";`, "f")
	testInputOutput(t, `<?php echo (4 and true) ? "t" : "f";`, "t")

	// Unary expression
	// Boolean
	testInputOutput(t, `<?php var_dump(+true);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(-true);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(+false);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(-false);`, "int(0)\n")
	// Floating
	testInputOutput(t, `<?php var_dump(+2.5);`, "float(2.5)\n")
	testInputOutput(t, `<?php var_dump(+(-2.5));`, "float(-2.5)\n")
	testInputOutput(t, `<?php var_dump(-3.0);`, "float(-3)\n")
	testInputOutput(t, `<?php var_dump(-(-3.0));`, "float(3)\n")
	testInputOutput(t, `<?php var_dump(~(-3.0));`, "int(2)\n")
	testInputOutput(t, `<?php var_dump(~3.0);`, "int(-4)\n")
	// Integer
	testInputOutput(t, `<?php var_dump(+2);`, "int(2)\n")
	testInputOutput(t, `<?php var_dump(+(-2));`, "int(-2)\n")
	testInputOutput(t, `<?php var_dump(-3);`, "int(-3)\n")
	testInputOutput(t, `<?php var_dump(-(-3));`, "int(3)\n")
	testInputOutput(t, `<?php var_dump(~(-3));`, "int(2)\n")
	testInputOutput(t, `<?php var_dump(~3);`, "int(-4)\n")

	// Post-/Prefix Inc-/Decrement
	// Integer
	testInputOutput(t, `<?php $a = 42; var_dump($a++); var_dump($a);`, "int(42)\nint(43)\n")
	testInputOutput(t, `<?php $a = 42; var_dump($a--); var_dump($a);`, "int(42)\nint(41)\n")
	testInputOutput(t, `<?php $a = 42; var_dump(++$a); var_dump($a);`, "int(43)\nint(43)\n")
	testInputOutput(t, `<?php $a = 42; var_dump(--$a); var_dump($a);`, "int(41)\nint(41)\n")
	// Boolean
	testInputOutput(t, `<?php $a = true; var_dump($a++); var_dump($a);`, "bool(true)\nbool(true)\n")
	testInputOutput(t, `<?php $a = true; var_dump($a--); var_dump($a);`, "bool(true)\nbool(true)\n")
	testInputOutput(t, `<?php $a = false; var_dump($a++); var_dump($a);`, "bool(false)\nbool(false)\n")
	testInputOutput(t, `<?php $a = false; var_dump($a--); var_dump($a);`, "bool(false)\nbool(false)\n")
	// Floating
	testInputOutput(t, `<?php $a = 42.0; var_dump($a++); var_dump($a);`, "float(42)\nfloat(43)\n")
	testInputOutput(t, `<?php $a = 42.0; var_dump($a--); var_dump($a);`, "float(42)\nfloat(41)\n")
	// Null
	testInputOutput(t, `<?php $a = null; var_dump($a++); var_dump($a);`, "NULL\nint(1)\n")
	testInputOutput(t, `<?php $a = null; var_dump($a--); var_dump($a);`, "NULL\nNULL\n")
	// String
	testInputOutput(t, `<?php $a = ""; var_dump($a++); var_dump($a);`, `string(0) ""`+"\n"+`string(1) "1"`+"\n")
	testInputOutput(t, `<?php $a = ""; var_dump($a--); var_dump($a);`, `string(0) ""`+"\nint(-1)\n")

	// Base operators
	// Integer
	testInputOutput(t, `<?php echo 4 >> 2;`, "1")
	testInputOutput(t, `<?php echo 8 << 2;`, "32")
	testInputOutput(t, `<?php $a = 13; echo $a <<= 1;`, "26")
	testInputOutput(t, `<?php echo 4 ^ 4;`, "0")
	testInputOutput(t, `<?php echo 8 ^ 4;`, "12")
	testInputOutput(t, `<?php $a = 13; echo $a ^= 4;`, "9")
	testInputOutput(t, `<?php echo 4 | 4;`, "4")
	testInputOutput(t, `<?php echo 8 | 4;`, "12")
	testInputOutput(t, `<?php $a = 8; echo $a |= 4;`, "12")
	testInputOutput(t, `<?php echo 8 & 4;`, "0")
	testInputOutput(t, `<?php echo 12 & 8;`, "8")
	testInputOutput(t, `<?php $a = 12; echo $a &= 4;`, "4")
	testInputOutput(t, `<?php echo 42 + 1;`, "43")
	testInputOutput(t, `<?php $a = 42; echo $a += 1;`, "43")
	testInputOutput(t, `<?php echo 42 - 1;`, "41")
	testInputOutput(t, `<?php $a = 42; echo $a -= 1;`, "41")
	testInputOutput(t, `<?php echo 42 * 2;`, "84")
	testInputOutput(t, `<?php $a = 42; echo $a *= 2;`, "84")
	testInputOutput(t, `<?php echo 42 / 2;`, "21")
	testInputOutput(t, `<?php $a = 42; echo $a /= 2;`, "21")
	testInputOutput(t, `<?php echo 42 % 5;`, "2")
	testInputOutput(t, `<?php $a = 42; echo $a %= 5;`, "2")
	testInputOutput(t, `<?php echo 2 ** 4;`, "16")
	testInputOutput(t, `<?php echo 2 ** 2 ** 2;`, "16")
	testInputOutput(t, `<?php $a = 2; echo $a **= 4;`, "16")
	// Floating
	testInputOutput(t, `<?php echo 42.0 + 1.5;`, "43.5")
	testInputOutput(t, `<?php $a = 42.0; echo $a += 1.5;`, "43.5")
	testInputOutput(t, `<?php echo 42 - 1.5;`, "40.5")
	testInputOutput(t, `<?php $a = 42.0; echo $a -= 1.5;`, "40.5")
	testInputOutput(t, `<?php echo 42.1 * 2;`, "84.2")
	testInputOutput(t, `<?php $a = 42.1; echo $a *= 2;`, "84.2")
	testInputOutput(t, `<?php echo 43.0 / 2;`, "21.5")
	testInputOutput(t, `<?php $a = 43.0; echo $a /= 2;`, "21.5")
	testInputOutput(t, `<?php echo 2.0 ** 4;`, "16")
	testInputOutput(t, `<?php $a = 2.0; echo $a **= 4;`, "16")
	// String
	testInputOutput(t, `<?php echo "a" . "bc";`, "abc")
	testInputOutput(t, `<?php $a = "a"; echo $a .= "bc";`, "abc")
	// Combined additions and multiplications
	testInputOutput(t, `<?php echo 31 + 21 + 11;`, "63")
	testInputOutput(t, `<?php echo 4 * 3 * 2;`, "24")
	testInputOutput(t, `<?php echo 2 + 3 * 4;`, "14")
	testInputOutput(t, `<?php echo (2 + 3) * 4;`, "20")
	testInputOutput(t, `<?php echo 2 * 3 + 4 * 5 + 6;`, "32")
	testInputOutput(t, `<?php echo 2 * (3 + 4) * 5 + 6;`, "76")
	testInputOutput(t, `<?php echo 2 + 3 * 4 + 5 * 6;`, "44")
}

// -------------------------------------- comparison -------------------------------------- MARK: comparison

func TestStrictComparison(t *testing.T) {
	// Table from https://www.php.net/manual/en/types.comparisons.php
	// true
	testInputOutput(t, `<?php var_dump(true === true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true === false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === 1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === -1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === "1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true === "");`, "bool(false)\n")
	// false
	testInputOutput(t, `<?php var_dump(false === false);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false === 1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false === 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false === -1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false === "1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false === "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false === "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false === null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false === []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false === "");`, "bool(false)\n")
	// 1
	testInputOutput(t, `<?php var_dump(1 === 1);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(1 === 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 === -1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 === "1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 === "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 === "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 === null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 === []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 === "");`, "bool(false)\n")
	// 0
	testInputOutput(t, `<?php var_dump(0 === 0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(0 === -1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 === "1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 === "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 === "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 === null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 === []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 === "");`, "bool(false)\n")
	// -1
	testInputOutput(t, `<?php var_dump(-1 === -1);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-1 === "1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 === "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 === "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 === null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 === []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 === "");`, "bool(false)\n")
	// "1"
	testInputOutput(t, `<?php var_dump("1" === "1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("1" === "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" === "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" === null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" === []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" === "");`, "bool(false)\n")
	// "0"
	testInputOutput(t, `<?php var_dump("0" === "0");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("0" === "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("0" === null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("0" === []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("0" === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("0" === "");`, "bool(false)\n")
	// "-1"
	testInputOutput(t, `<?php var_dump("-1" === "-1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("-1" === null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("-1" === []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("-1" === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("-1" === "");`, "bool(false)\n")
	// null
	testInputOutput(t, `<?php var_dump(null === null);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(null === []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(null === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(null === "");`, "bool(false)\n")
	// []
	testInputOutput(t, `<?php var_dump([] === []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([] === "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] === "");`, "bool(false)\n")
	// "php"
	testInputOutput(t, `<?php var_dump("php" === "php");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("php" === "");`, "bool(false)\n")
	// ""
	testInputOutput(t, `<?php var_dump("" === "");`, "bool(true)\n")

	// Object
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 === $c1);`, "bool(true)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 !== $c1);`, "bool(false)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 === null);`, "bool(false)\n")
}

func TestLooseComparison(t *testing.T) {
	// Table from https://www.php.net/manual/en/types.comparisons.php
	// true
	testInputOutput(t, `<?php var_dump(true == true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true == false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true == 1);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true == 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true == -1);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true == "1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true == "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true == "-1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true == null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true == []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true == "php");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true == "");`, "bool(false)\n")
	// false
	testInputOutput(t, `<?php var_dump(false == false);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false == 1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false == 0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false == -1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false == "1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false == "0");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false == "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false == null);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false == []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false == "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false == "");`, "bool(true)\n")
	// 1
	testInputOutput(t, `<?php var_dump(1 == 1);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(1 == 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 == -1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 == "1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(1 == "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 == "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 == null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 == []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 == "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 == "");`, "bool(false)\n")
	// 0
	testInputOutput(t, `<?php var_dump(0 == 0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(0 == -1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 == "1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 == "0");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(0 == "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 == null);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(0 == []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 == "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 == "");`, "bool(false)\n")
	// -1
	testInputOutput(t, `<?php var_dump(-1 == -1);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-1 == "1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 == "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 == "-1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-1 == null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 == []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 == "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-1 == "");`, "bool(false)\n")
	// "1"
	testInputOutput(t, `<?php var_dump("1" == "1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("1" == "0");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" == "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" == null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" == []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" == "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" == "");`, "bool(false)\n")
	// "0"
	testInputOutput(t, `<?php var_dump("0" == "0");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("0" == "-1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("0" == null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("0" == []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("0" == "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("0" == "");`, "bool(false)\n")
	// "-1"
	testInputOutput(t, `<?php var_dump("-1" == "-1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("-1" == null);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("-1" == []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("-1" == "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("-1" == "");`, "bool(false)\n")
	// null
	testInputOutput(t, `<?php var_dump(null == null);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(null == []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(null == "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(null == "");`, "bool(true)\n")
	// []
	testInputOutput(t, `<?php var_dump([] == []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([] == "php");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] == "");`, "bool(false)\n")
	// "php"
	testInputOutput(t, `<?php var_dump("php" == "php");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("php" == "");`, "bool(false)\n")
	// ""
	testInputOutput(t, `<?php var_dump("" == "");`, "bool(true)\n")

	// Warning-Box from https://www.php.net/manual/en/language.operators.comparison.php
	testInputOutput(t, `<?php var_dump(0 == "a");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("a" == 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" == "01a");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("01a" == "1");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(1 == "1a");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1a" == 1);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("1" == "01");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("01" == "1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("10" == "1e1");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("1e1" == "10");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(100 == "1e2");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(1e2 == "100");`, "bool(true)\n")

	// Object
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 == $c1);`, "bool(true)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 != $c1);`, "bool(false)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 == null);`, "bool(false)\n")
	testInputOutput(t, `<?php class c {} $c1 = new C; var_dump($c1 == true);`, "bool(true)\n")
	testInputOutput(t, `<?php class c {} $c1 = new C; var_dump($c1 == false);`, "bool(false)\n")

	// QIQ feature: always strict comparison
	devIni := ini.NewDevIni()
	devIni.Set("qiq.strict_comparison", "1", ini.INI_SYSTEM)
	testInputOutputCustomIni(t, devIni, `<?php var_dump(true == 1);`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(true == -1);`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(true == "1");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(true == "-1");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(true == "php");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(false == 0);`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(false == "0");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(false == null);`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(false == []);`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(false == "");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(1 == "1");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(0 == "0");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(0 == null);`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(-1 == "-1");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(null == []);`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(null == "");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump("1" == "01");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump("01" == "1");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump("10" == "1e1");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump("1e1" == "10");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(100 == "1e2");`, "bool(false)\n")
	testInputOutputCustomIni(t, devIni, `<?php var_dump(1e2 == "100");`, "bool(false)\n")
}

func TestCompareRelation(t *testing.T) {
	// Array
	// Array - Array
	testInputOutput(t, `<?php var_dump(["a"] < []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(["a"] <= []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(["a"] <=> []);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump([] < []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] <= []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([] <=> []);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(["a"] <=> ["a"]);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump([] < ["a"]);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([] <= ["a"]);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([] <=> ["a"]);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump([21] <=> [42]);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump([21] <=> [21, 42]);`, "int(-1)\n")
	// Array - Boolean
	testInputOutput(t, `<?php var_dump([] < true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([] <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([] <=> true);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump([] < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] <= false);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([] <=> false);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump([42] < true);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([42] <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([42] <=> true);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump([42] < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([42] <= false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([42] <=> false);`, "int(1)\n")
	// Array - Floating
	testInputOutput(t, `<?php var_dump([] < 2.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] <= 2.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] <=> 2.5);`, "int(1)\n")
	// Array - Integer
	testInputOutput(t, `<?php var_dump([] < 2);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] <= 2);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] <=> 2);`, "int(1)\n")
	// Array - Null
	testInputOutput(t, `<?php var_dump([] < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] <= NULL);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump([] <=> NULL);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(["abc"] < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(["abc"] <= NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(["abc"] <=> NULL);`, "int(1)\n")
	// Array - String
	testInputOutput(t, `<?php var_dump([] < "");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] <= "");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump([] <=> "");`, "int(1)\n")

	// Boolean
	// Boolean - Array
	testInputOutput(t, `<?php var_dump(true < []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <= []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <=> []);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(false < []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false <= []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <=> []);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(true < [42]);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <= [42]);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true <=> [42]);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(false < [42]);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <= [42]);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <=> [42]);`, "int(-1)\n")
	// Boolean - Boolean
	testInputOutput(t, `<?php var_dump(true < true);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true <=> true);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(true < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <= false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <=> false);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(false < true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <=> true);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(false < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false <= false);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <=> false);`, "int(0)\n")
	// Boolean - Floating
	testInputOutput(t, `<?php var_dump(true < 2.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <= 2.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true <=> 2.5);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(true < 0.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <= 0.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true <=> 0.5);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(false < 2.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <= 2.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <=> 2.5);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(false < 0.0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false <= 0.0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <=> 0.0);`, "int(0)\n")
	// Boolean - Integer
	testInputOutput(t, `<?php var_dump(true < 2);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <= 2);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(true <=> 2);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(true < 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <= 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <=> 0);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(false < 2);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <= 2);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <=> 2);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(false < 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false <= 0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <=> 0);`, "int(0)\n")
	// Boolean - Null
	testInputOutput(t, `<?php var_dump(true < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <= NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(true <=> NULL);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(false < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(false <= NULL);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(false <=> NULL);`, "int(0)\n")

	// Floating
	// Floating - Array
	testInputOutput(t, `<?php var_dump(2.5 < []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2.5 <= []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2.5 <=> []);`, "int(-1)\n")
	// Floating - Boolean
	testInputOutput(t, `<?php var_dump(2.5 < true);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2.5 <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2.5 <=> true);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(0.5 < true);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0.5 <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(0.5 <=> true);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(2.5 < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2.5 <= false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2.5 <=> false);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(0.0 < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0.0 <= false);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(0.0 <=> false);`, "int(0)\n")
	// Floating - Floating
	testInputOutput(t, `<?php var_dump(2.5 < 3.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2.5 <= 3.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2.5 <=> 3.5);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(3.5 < 3.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(3.5 <= 3.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(3.5 <=> 3.5);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(12.5 < 3.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(12.5 <= 3.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(12.5 <=> 3.5);`, "int(1)\n")
	// Floating - Integer
	testInputOutput(t, `<?php var_dump(2.5 < 3);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2.5 <= 3);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2.5 <=> 3);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(3.0 < 3);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(3.0 <= 3);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(3.0 <=> 3);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(12.5 < 3);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(12.5 <= 3);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(12.5 <=> 3);`, "int(1)\n")
	// Floating - Null
	testInputOutput(t, `<?php var_dump(2.5 < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2.5 <= NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2.5 <=> NULL);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(-2.5 < NULL);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-2.5 <= NULL);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-2.5 <=> NULL);`, "int(-1)\n")

	// Integer
	// Integer - Array
	testInputOutput(t, `<?php var_dump(2 < []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2 <= []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2 <=> []);`, "int(-1)\n")
	// Integer - Boolean
	testInputOutput(t, `<?php var_dump(2 < true);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2 <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2 <=> true);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(2 < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2 <= false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2 <=> false);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(-2 < true);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-2 <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-2 <=> true);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(-2 < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-2 <= false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(-2 <=> false);`, "int(1)\n")
	// Integer - Floating
	testInputOutput(t, `<?php var_dump(2 < 3.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2 <= 3.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2 <=> 3.5);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(3 < 3.0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(3 <= 3.0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(3 <=> 3.0);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(12 < 3.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(12 <= 3.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(12 <=> 3.5);`, "int(1)\n")
	// Integer - Integer
	testInputOutput(t, `<?php var_dump(2 < 2);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2 <= 2);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(2 <=> 2);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(2 < 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2 <= 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2 <=> 0);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(-2 < 2);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-2 <= 2);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-2 <=> 2);`, "int(-1)\n")
	// Integer - Null
	testInputOutput(t, `<?php var_dump(2 < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2 <= NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(2 <=> NULL);`, "int(1)\n")
	testInputOutput(t, `<?php var_dump(0 < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(0 <= NULL);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(0 <=> NULL);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(-2 < NULL);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-2 <= NULL);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(-2 <=> NULL);`, "int(-1)\n")
	// Integer - String
	testInputOutput(t, `<?php var_dump('..9' > 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump('.9.' > 0);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump('9..' > 0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump('9' > 0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump('9.9' > 0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump('9.9.9' > 0);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump('9X' > 0);`, "bool(true)\n")

	// Null
	// Null - Array
	testInputOutput(t, `<?php var_dump(NULL < []);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(NULL <= []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> []);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(NULL < ["abc"]);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <= ["abc"]);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> ["abc"]);`, "int(-1)\n")
	// Null - Boolean
	testInputOutput(t, `<?php var_dump(NULL < true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> true);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(NULL < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(NULL <= false);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> false);`, "int(0)\n")
	// Null - Floating
	testInputOutput(t, `<?php var_dump(NULL < 2.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <= 2.5);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> 2.5);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(NULL < -2.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(NULL <= -2.5);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> -2.5);`, "int(1)\n")
	// Null - Integer
	testInputOutput(t, `<?php var_dump(NULL < 2);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <= 2);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> 2);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump(NULL < -2);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(NULL <= -2);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> -2);`, "int(1)\n")
	// Null - Null
	testInputOutput(t, `<?php var_dump(NULL < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(NULL <= NULL);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> NULL);`, "int(0)\n")
	// Null - String
	testInputOutput(t, `<?php var_dump(NULL < "");`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump(NULL <= "");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> "");`, "int(0)\n")
	testInputOutput(t, `<?php var_dump(NULL < "abc");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <= "abc");`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump(NULL <=> "abc");`, "int(-1)\n")

	// String
	// String - Array
	testInputOutput(t, `<?php var_dump("" < []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("" <= []);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("" <=> []);`, "int(-1)\n")
	// String - Boolean
	testInputOutput(t, `<?php var_dump("" < true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("" <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("" <=> true);`, "int(-1)\n")
	testInputOutput(t, `<?php var_dump("a" < true);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("a" <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("a" <=> true);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump("" < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("" <= false);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("" <=> false);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump("a" < false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("a" <= false);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("a" <=> false);`, "int(1)\n")
	// String - Null
	testInputOutput(t, `<?php var_dump("" < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("" <= NULL);`, "bool(true)\n")
	testInputOutput(t, `<?php var_dump("" <=> NULL);`, "int(0)\n")
	testInputOutput(t, `<?php var_dump("22" < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("22" <= NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php var_dump("22" <=> NULL);`, "int(1)\n")

	// Object
	// Object - Null
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 < NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 <= NULL);`, "bool(false)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 <=> null);`, "int(1)\n")
	// Object - Boolean
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 < true);`, "bool(false)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 <= true);`, "bool(true)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 <=> true);`, "int(0)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 < false);`, "bool(false)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 <= false);`, "bool(false)\n")
	testInputOutput(t, `<?php class c {} $c1 = new c; var_dump($c1 <=> false);`, "int(1)\n")
}

// -------------------------------------- functions -------------------------------------- MARK: functions

func TestUserFunctions(t *testing.T) {
	// Error if function is never declared
	testForError(t, "<?php noneExistingFunction(); ", phpError.NewError("Call to undefined function noneexistingfunction() in %s:1:7", TEST_FILE_NAME))

	// Simple user defined function without types, params, ...
	// Check if function definition can be after the function call
	testInputOutput(t,
		`<?php
			echo "a";
			helloWorld();
			echo "b";
			function helloWorld() { echo "Hello World"; }
			echo "c";
		`,
		"aHello Worldbc",
	)
	testInputOutput(t,
		`<?php
			echo "a";
			helloWorld();
			echo "b";
			{{{ function helloWorld() { echo "Hello World"; } }}}
			echo "c";
		`,
		"aHello Worldbc",
	)

	// Test correct scoping: $a is available in the function body
	testInputOutput(t,
		`<?php
			$a = 1;
			func();
			function func() {
				echo isset($a) ? "y" : "n";
				$b = 2;
			}
			echo isset($b) ? "y" : "n";
		`,
		"nn",
	)

	// Parameters
	testInputOutput(t,
		`<?php
			$a = 1;
			func($a);
			func(2);
			function func($param) {
				echo $param;
			}
		`,
		"12",
	)
	testInputOutput(t,
		`<?php
			$a = 1;
			func($a, 2);
			func(1, 1+1);
			function func($param1, $param2) {
				echo $param1 + $param2;
			}
		`,
		"33",
	)
	// Typed parameters
	testInputOutput(t,
		`<?php
			$a = 1;
			func($a, 2);
			func(1, 1+1);
			function func(int|float $param1, int|float $param2) {
				echo $param1 + $param2;
			}
		`,
		"33",
	)

	// Return type
	testInputOutput(t,
		`<?php
			$a = 1;
			echo func($a, 2);
			echo func(1, 1+1);
			function func(int|float $param1, int|float $param2,): int|float {
				return $param1 + $param2;
			}
		`,
		"33",
	)

	// Dynamic function names
	testInputOutput(t,
		`<?php
			$a = "func";
			$a("abc");
			function func(string $str): void {
				echo "func called: $str";
			}
		`,
		"func called: abc",
	)

	// By reference parameter
	testInputOutput(t,
		`<?php $a = 42; function inc(&$num) { $num++; }  inc($a); var_dump($a);`,
		"int(43)\n",
	)

	// Default value
	testInputOutput(t, `<?php function f($a = 2) { echo $a; } f();`, "2")
	testInputOutput(t, `<?php function f($a = 2) { echo $a; } f(42);`, "42")

	// Redeclared function
	testForError(t, "<?php function f1() {} function F1() {}", phpError.NewError("Cannot redeclare function F1() (previously declared in %s:1:7) in %s:1:24", TEST_FILE_NAME, TEST_FILE_NAME))
	testForError(t, "<?php function f1() {} function f1() {}", phpError.NewError("Cannot redeclare function f1() (previously declared in %s:1:7) in %s:1:24", TEST_FILE_NAME, TEST_FILE_NAME))
	testForError(t, "<?php function var_dump() {}", phpError.NewError("Cannot redeclare function var_dump() in %s:1:7", TEST_FILE_NAME))

	// function_exists
	testInputOutput(t, "<?php var_dump(function_exists('intval'));", "bool(true)\n")
	testInputOutput(t, "<?php var_dump(function_exists('someUndefinedFunc'));", "bool(false)\n")
	testInputOutput(t, "<?php function myUserFunc() {} var_dump(function_exists('myUserFunc'));", "bool(true)\n")
}

// -------------------------------------- classes and objects -------------------------------------- MARK: classes and objects

func TestClasses(t *testing.T) {
	// Constructor
	testInputOutput(t, `<?php class c {} $c = new c; echo "Class created";`, "Class created")
	testInputOutput(t, `<?php class c {} $c = new c(); echo "Class created";`, "Class created")
	testInputOutput(t, `<?php class c { public function __construct(string $name) { echo "Construct(Name: " . $name . ")\n"; } } $c = new c("Max"); echo "Done";`, "Construct(Name: Max)\nDone")

	// Member access
	testInputOutput(t, `<?php class c { public int $i = 42; } $c = new c; var_dump($c->i);`, "int(42)\n")
	// TODO testInputOutput(t, `<?php class c { private int $i = 42; } $c = new c; var_dump($c->i);`, "int(42)\n")
	testForError(t, `<?php class c { } $c = new c; $c->prop;`, phpError.NewError("Undefined property: c::$prop in %s:1:35", TEST_FILE_NAME))
	// Overwrite property value from outside
	testInputOutput(t, `<?php class C { public $p = "ab"; } $c = new C; echo $c->p; $c->p = "cd"; echo $c->p;`, "abcd")
	testInputOutput(t, `<?php class C { public $string = "ab"; } $c = new C; echo $c->string; $c->string = "cd"; echo $c->string;`, "abcd")
	// TODO add tests for private properties to check if they cannot be set
	// Overwrite property value from inside
	testInputOutput(t,
		`<?php class C { public string $p = "foo"; public function __construct() { $this->p = "bar"; } }
		$c = new C; echo $c->p;`,
		"bar",
	)
	testInputOutput(t, `<?php class C { const field = 42; } $c = new C; echo $c::field;`, "42")
	testInputOutput(t, `<?php class C { const field = 42; } echo c::field;`, "42")
	testInputOutput(t, `<?php class C { const field = 42; } echo C::field;`, "42")
	testInputOutput(t, `<?php class C { CONST field = 42; } echo C::field;`, "42")
	testInputOutput(t, `<?php class C { const constant = 42; function f() { echo self::constant; } } $c = new C; echo $c->f();`, "42")
	testInputOutput(t, `<?php class C { const constant = 42; function f() { echo SELF::constant; } } $c = new C; echo $c->f();`, "42")
	testInputOutput(t, `<?php class B { const constant = 41; } class C extends B { const constant = 42; function f() { echo self::constant; } } $c = new C; echo $c->f();`, "42")
	testInputOutput(t, `<?php class B { const constant = 41; } class C extends B { const constant = 42; function f() { echo parent::constant; } } $c = new C; echo $c->f();`, "41")
	testInputOutput(t, `<?php class B { const constant = 41; } class C extends B { const constant = 42; function f() { echo PARENT::constant; } } $c = new C; echo $c->f();`, "41")
	testForError(t, `<?php class C { function f() { echo PARENT::constant; } } $c = new C; echo $c->f();`, phpError.NewError(`Cannot use "parent" when current class scope has no parent in %s:1:45`, TEST_FILE_NAME))

	// Member call
	testInputOutput(t, `<?php class C { function f(): void { echo "f()"; } } $c = new C; $c->f();`, "f()")
	testForError(t, `<?php class C { } $c = new C; $c->g();`, phpError.NewError("Uncaught Error: Call to undefined method C::g() in %s:1:35", TEST_FILE_NAME))
	testForError(t, `<?php $a = []; $a->g();`, phpError.NewError("Uncaught Error: Call to a member function g() on array in %s:1:18", TEST_FILE_NAME))
	testInputOutput(t, `<?php class C { static function f(): void { echo "f()"; } } $c = new C; $c::f();`, "f()")
	testInputOutput(t, `<?php class C { static function f(): void { echo "f()"; } } c::f();`, "f()")
	testInputOutput(t, `<?php class C { static function f(): void { echo "f()"; } } C::f();`, "f()")
	testForError(t, `<?php class C { function f(): void { echo "f()"; } } C::f();`, phpError.NewError("Uncaught Error: Non-static method C::f() cannot be called statically in %s:1:57", TEST_FILE_NAME))

	// Destructor
	testInputOutput(t, `<?php class c { function __destruct() { echo __METHOD__; } } $c = new c; echo "Done\n";`, "Done\nc::__destruct")
	testInputOutput(t, `<?php class c { function __destruct() { echo __METHOD__ . "\n"; } } new c; echo "Done";`, "c::__destruct\nDone")
	testInputOutput(t, `<?php
		class C {
			public function __construct() { echo __METHOD__ . "\n"; }
			public function __destruct() { echo __METHOD__ . "\n"; }
		}
		function func() {
			echo "1\n";
			new C;
			echo "2\n";
		}
		func();

		function func1() {
			echo "3\n";
			$c = new C;
			echo "4\n";
		}
		func1();`,
		"1\nC::__construct\nC::__destruct\n2\n3\nC::__construct\n4\nC::__destruct\n",
	)

	// Class with namespace
	testInputOutput(t, `<?php
		namespace My\Namespace;
		class C { function __construct() { var_dump(__CLASS__); } }
		new C;`,
		`string(14) "My\Namespace\C"`+"\n",
	)

	// Class with interface
	testInputOutput(t, `<?php
		interface I { public function F($p); }
		class C implements I { public function F($p) { } }
		new C; echo 'OK';`,
		"OK",
	)

	// Class that extends unkown class
	testForError(t,
		`<?php class C extends B { }`,
		phpError.NewError(`Class "B" not found in %s:1:7`, TEST_FILE_NAME),
	)

	// Class that implements unkown interface
	testForError(t,
		`<?php class C implements I { }`,
		phpError.NewError(`Interface "I" not found in %s:1:7`, TEST_FILE_NAME),
	)

	// Class that does not implement interface
	testForError(t,
		`<?php interface I { public function F($p); }
		class C implements I { }`,
		phpError.NewError("Class C contains 1 abstract method and must therefore be declared abstract or implement the remaining method (I::F) in %s:2:3", TEST_FILE_NAME),
	)
	testForError(t,
		`<?php interface I { public function F($p); function G(); }
		class C implements I { }`,
		phpError.NewError("Class C contains 2 abstract methods and must therefore be declared abstract or implement the remaining methods (I::F, I::G) in %s:2:3", TEST_FILE_NAME),
	)

	// Class with wrong method signature
	testForError(t,
		`<?php interface I { public function F(null|string $p): ?string; }
		class C implements I { function F($p) {} }`,
		phpError.NewError("Declaration of C::F($p) must be compatible with I::F(?string $p): ?string in %s:2:3", TEST_FILE_NAME),
	)
	testForError(t,
		`<?php namespace Space;
		interface I { public function F(null|string $p): ?string; }
		class C implements I { function F($p) {} }`,
		phpError.NewError(`Declaration of Space\C::F($p) must be compatible with Space\I::F(?string $p): ?string in %s:3:3`, TEST_FILE_NAME),
	)

	// Class with unimplemented function and __call
	testInputOutput(t, `<?php
		class C { function __call($method, $args) { var_dump($method, $args); } }
		$c = new C; $c->f(1);`,
		"string(1) \"f\"\narray(1) {\n  [0]=>\n  int(1)\n}\n",
	)

	// Class with unimplemented static function and __callStatic
	testInputOutput(t, `<?php
		class C { static function __callStatic($method, $args) { var_dump($method, $args); } }
		C::f(1);`,
		"string(1) \"f\"\narray(1) {\n  [0]=>\n  int(1)\n}\n",
	)

	// Class with functions implemented in parent
	testInputOutput(t, `<?php
		class A { public function __construct() { echo __METHOD__; } }
		class B extends A { }
		new B;`,
		"A::__construct",
	)

	// Class redeclaration
	testForError(t, `<?php class C { } class c { }`,
		phpError.NewError(`Cannot redeclare class C (previously declared in %s:1:7) in %s:1:19`, TEST_FILE_NAME, TEST_FILE_NAME),
	)
	testForError(t, `<?php class C { } class C { }`,
		phpError.NewError(`Cannot redeclare class C (previously declared in %s:1:7) in %s:1:19`, TEST_FILE_NAME, TEST_FILE_NAME),
	)
	testForError(t, `<?php class stdClass { }`, phpError.NewError(`Cannot redeclare class stdClass in %s:1:7`, TEST_FILE_NAME))
	testForError(t, "<?php class Traversable { }", phpError.NewError(`Cannot redeclare interface Traversable in %s:1:7`, TEST_FILE_NAME))

	// Interface redeclaration
	testForError(t, `<?php interface I { } interface i { }`,
		phpError.NewError(`Cannot redeclare interface I (previously declared in %s:1:7) in %s:1:23`, TEST_FILE_NAME, TEST_FILE_NAME),
	)
	testForError(t, `<?php interface I { } interface I { }`,
		phpError.NewError(`Cannot redeclare interface I (previously declared in %s:1:7) in %s:1:23`, TEST_FILE_NAME, TEST_FILE_NAME),
	)
	testForError(t, "<?php interface Traversable { }", phpError.NewError(`Cannot redeclare interface Traversable in %s:1:7`, TEST_FILE_NAME))
	testForError(t, "<?php interface stdClass { }", phpError.NewError(`Cannot redeclare class stdClass in %s:1:7`, TEST_FILE_NAME))

	// Reassignment
	testInputOutput(t, `<?php class C {}$o = new C;echo "1";$o = new C;echo "2";`, "12")
}

// TODO Add interface test cases
/*

OK: <?php interface I { public function F(null|string $p): void; }
		class C implements I { function F($p): void {} }

OK: <?php interface I { public function F(null|string $p); }
		class C implements I { function F($p): void {} }

	testForError(t,
		`<?php interface I { public function F(null|string $p): null|string|int; }
		class C implements I { function F($p) {} }`,
		phpError.NewError("Declaration of C::F($p) must be compatible with I::F(?string $p): string|int|null in %s:1:2", TEST_FILE_NAME),
	)

<?php

interface I { public function F(null|string $p): null|string|int;}
class C implements I { public function F() { } }


Declaration of C::F() must be compatible with I::F(?string $p): string|int|null in
*/
