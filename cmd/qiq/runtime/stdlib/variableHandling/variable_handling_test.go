package variableHandling

import (
	"QIQ/cmd/qiq/ast"
	"QIQ/cmd/qiq/runtime/values"
	"testing"
)

// -------------------------------------- boolval -------------------------------------- MARK: boolval

func TestLibBoolval(t *testing.T) {
	doTest := func(runtimeValue values.RuntimeValue, expected bool) {
		actual, err := BoolVal(runtimeValue)
		if err != nil {
			t.Errorf(`Unexpected error: "%s"`, err)
			return
		}
		if actual != expected {
			t.Errorf(`Expected: "%t", Got "%t"`, expected, actual)
		}
	}

	// array to boolean
	doTest(values.NewArray(), false)
	array := values.NewArray()
	array.SetElement(nil, values.NewInt(42))
	doTest(array, true)

	// boolean to boolean
	doTest(values.NewBool(true), true)
	doTest(values.NewBool(false), false)

	// integer to boolean
	doTest(values.NewInt(0), false)
	doTest(values.NewInt(-0), false)
	doTest(values.NewInt(1), true)
	doTest(values.NewInt(42), true)
	doTest(values.NewInt(-2), true)

	// floating to boolean
	doTest(values.NewFloat(0.0), false)
	doTest(values.NewFloat(1.5), true)
	doTest(values.NewFloat(42.0), true)
	doTest(values.NewFloat(-2.0), true)

	// string to boolean
	doTest(values.NewStr(""), false)
	doTest(values.NewStr("0"), false)
	doTest(values.NewStr("Hi"), true)

	// null to boolean
	doTest(values.NewNull(), false)

	// object to bool
	doTest(values.NewObject(ast.NewClassDeclarationStmt(0, nil, "C", false, false)), true)
}

// -------------------------------------- floatval -------------------------------------- MARK: floatval

func TestLibFloatval(t *testing.T) {
	doTest := func(runtimeValue values.RuntimeValue, expected float64) {
		actual, err := FloatVal(runtimeValue, true)
		if err != nil {
			t.Errorf(`Unexpected error: "%s"`, err)
			return
		}
		if actual != expected {
			t.Errorf(`Expected: "%g", Got "%g"`, expected, actual)
		}
	}

	// boolean to floating
	doTest(values.NewBool(true), 1)
	doTest(values.NewBool(false), 0)

	// integer to floating
	doTest(values.NewInt(0), 0)
	doTest(values.NewInt(-0), 0)
	doTest(values.NewInt(1), 1)
	doTest(values.NewInt(42), 42)
	doTest(values.NewInt(-2), -2)

	// floating to floating
	doTest(values.NewFloat(0.0), 0)
	doTest(values.NewFloat(1.5), 1.5)
	doTest(values.NewFloat(42.0), 42)
	doTest(values.NewFloat(-2.0), -2)

	// string to floating
	doTest(values.NewStr(""), 0)
	doTest(values.NewStr("0"), 0)
	doTest(values.NewStr("42"), 42)
	doTest(values.NewStr("+42"), 42)
	doTest(values.NewStr("-42.4"), -42.4)
	doTest(values.NewStr("Hi"), 0)

	// null to floating
	doTest(values.NewNull(), 0)
}

// -------------------------------------- intval -------------------------------------- MARK: intval

func TestLibIntval(t *testing.T) {
	doTest := func(runtimeValue values.RuntimeValue, expected int64) {
		actual, err := IntVal(runtimeValue, true)
		if err != nil {
			t.Errorf(`Unexpected error: "%s"`, err)
			return
		}
		if actual != expected {
			t.Errorf(`Expected: "%d", Got "%d"`, expected, actual)
		}
	}

	// array to integer
	doTest(values.NewArray(), 0)
	array := values.NewArray()
	array.SetElement(nil, values.NewInt(42))
	doTest(array, 1)

	// boolean to integer
	doTest(values.NewBool(true), 1)
	doTest(values.NewBool(false), 0)

	// integer to integer
	doTest(values.NewInt(0), 0)
	doTest(values.NewInt(-0), 0)
	doTest(values.NewInt(1), 1)
	doTest(values.NewInt(42), 42)
	doTest(values.NewInt(-2), -2)

	// floating to integer
	doTest(values.NewFloat(0.0), 0)
	doTest(values.NewFloat(1.5), 1)
	doTest(values.NewFloat(42.0), 42)
	doTest(values.NewFloat(-2.0), -2)

	// string to integer
	doTest(values.NewStr(""), 0)
	doTest(values.NewStr("0"), 0)
	doTest(values.NewStr("1"), 1)
	doTest(values.NewStr("42"), 42)
	doTest(values.NewStr("+42"), 42)
	doTest(values.NewStr("-42"), -42)
	doTest(values.NewStr("Hi"), 0)

	// null to integer
	doTest(values.NewNull(), 0)
}

// -------------------------------------- strval -------------------------------------- MARK: strval

func TestLibStrval(t *testing.T) {
	doTest := func(runtimeValue values.RuntimeValue, expected string) {
		actual, err := StrVal(runtimeValue)
		if err != nil {
			t.Errorf(`Unexpected error: "%s"`, err)
			return
		}
		if actual != expected {
			t.Errorf(`Expected: "%s", Got "%s"`, expected, actual)
		}
	}

	// array to string
	doTest(values.NewArray(), "Array")

	// boolean to string
	doTest(values.NewBool(true), "1")
	doTest(values.NewBool(false), "")

	// integer to string
	doTest(values.NewInt(0), "0")
	doTest(values.NewInt(-0), "0")
	doTest(values.NewInt(1), "1")
	doTest(values.NewInt(42), "42")
	doTest(values.NewInt(-2), "-2")

	// floating to string
	doTest(values.NewFloat(0.0), "0")
	doTest(values.NewFloat(1.5), "1.5")
	doTest(values.NewFloat(42.0), "42")
	doTest(values.NewFloat(-2.0), "-2")

	// string to string
	doTest(values.NewStr(""), "")
	doTest(values.NewStr("0"), "0")
	doTest(values.NewStr("Hi"), "Hi")

	// null to string
	doTest(values.NewNull(), "")
}
