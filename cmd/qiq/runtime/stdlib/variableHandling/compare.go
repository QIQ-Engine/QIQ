package variableHandling

import (
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/runtime/values"
	"strings"
)

// -------------------------------------- CompareRelation -------------------------------------- MARK: CompareRelation

func CompareRelation(lhs values.RuntimeValue, operator string, rhs values.RuntimeValue, leadingNumeric bool) (*values.Slot, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression

	// Note that greater-than semantics is implemented as the reverse of less-than, i.e. "$a > $b" is the same as "$b < $a".
	// This may lead to confusing results if the operands are not well-ordered
	// - such as comparing two objects not having comparison semantics, or comparing arrays.

	// Operator "<=>" represents comparison operator between two expressions,
	// with the result being an integer less than "0" if the expression on the left is less than the expression on the right
	// (i.e. if "$a < $b" would return "TRUE"), as defined below by the semantics of the operator "<",
	// integer "0" if those expressions are equal (as defined by the semantics of the == operator) and
	// integer greater than 0 otherwise.

	// Operator "<" represents less-than, operator ">" represents greater-than, operator "<=" represents less-than-or-equal-to,
	// and operator ">=" represents greater-than-or-equal-to. The type of the result is bool.

	// The following table shows the result for comparison of different types, with the left operand displayed vertically
	// and the right displayed horizontally. The conversions are performed according to type conversion rules.

	// See in compareRelation[Type] ...

	// "<" means that the left operand is always less than the right operand.
	// ">" means that the left operand is always greater than the right operand.
	// "->" means that the left operand is converted to the type of the right operand.
	// "<-" means that the right operand is converted to the type of the left operand.

	// A number means one of the cases below:
	//   2. If one of the operands has arithmetic type, is a resource, or a numeric string,
	//      which can be represented as int or float without loss of precision,
	//      the operands are converted to the corresponding arithmetic type, with float taking precedence over int,
	//      and resources converting to int. The result is the numerical comparison of the two operands after conversion.
	//
	//   3. If only one operand has object type, if the object has comparison handler, that handler defines the result.
	//      Otherwise, if the object can be converted to the other operand’s type, it is converted and the result is used for the comparison.
	//      Otherwise, the object compares greater-than any other operand type.
	//
	//   4. If both operands are non-numeric strings, the result is the lexical comparison of the two operands.
	//      Specifically, the strings are compared byte-by-byte starting with their first byte.
	//      If the two bytes compare equal and there are no more bytes in either string, the strings are equal and the comparison ends;
	//      otherwise, if this is the final byte in one string, the shorter string compares less-than the longer string and the comparison ends.
	//      If the two bytes compare unequal, the string having the lower-valued byte compares less-than the other string, and the comparison ends.
	//      If there are more bytes in the strings, the process is repeated for the next pair of bytes.
	//
	//   6. When comparing two objects, if any of the object types has its own compare semantics, that would define the result,
	//      with the left operand taking precedence. Otherwise, if the objects are of different types, the comparison result is FALSE.
	//      If the objects are of the same type, the properties of the objects are compares using the array comparison described above.

	// Reduce code complexity and duplication by only implementing less-than and less-than-or-equal-to
	switch operator {
	case ">":
		return CompareRelation(rhs, "<", lhs, leadingNumeric)
	case ">=":
		return CompareRelation(rhs, "<=", lhs, leadingNumeric)
	}

	switch lhs.GetType() {
	case values.ArrayValue:
		return compareRelationArray(lhs.(*values.Array), operator, rhs, leadingNumeric)
	case values.BoolValue:
		return compareRelationBoolean(lhs.(*values.Bool), operator, rhs)
	case values.FloatValue:
		return compareRelationFloating(lhs.(*values.Float), operator, rhs, leadingNumeric)
	case values.IntValue:
		return compareRelationInteger(lhs.(*values.Int), operator, rhs, leadingNumeric)
	case values.StrValue:
		return compareRelationString(lhs.(*values.Str), operator, rhs, leadingNumeric)
	case values.NullValue:
		return compareRelationNull(operator, rhs, leadingNumeric)
	case values.ObjectValue:
		return compareRelationObject(lhs.(*values.Object), operator, rhs)
	default:
		return values.NewVoidSlot(), phpError.NewError(`compareRelation: Type "%s" not implemented`, lhs.GetType())
	}

}

func compareRelationArray(lhs *values.Array, operator string, rhs values.RuntimeValue, leadingNumeric bool) (*values.Slot, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression

	//        NULL  bool  int  float  string  array  object  resource
	// array   <-    ->    >    >      >       5      3       >

	//   5. If both operands have array type, if the arrays have different numbers of elements,
	//      the one with the fewer is considered less-than the other one, regardless of the keys and values in each, and the comparison ends.
	//      For arrays having the same numbers of elements, the keys from the left operand are considered one by one,
	//      if the next key in the left-hand operand exists in the right-hand operand, the corresponding values are compared.
	//      If they are unequal, the array containing the lesser value is considered less-than the other one, and the comparison ends;
	//      otherwise, the process is repeated with the next element.
	//      If the next key in the left-hand operand does not exist in the right-hand operand, the arrays cannot be compared and FALSE is returned.
	//      If all the values are equal, then the arrays are considered equal.

	// TODO compareRelationArray - object
	// TODO compareRelationArray - resource

	if rhs.GetType() == values.NullValue {
		var err phpError.Error
		rhs, err = ArrayVal(rhs)
		if err != nil {
			return values.NewVoidSlot(), err
		}
	}

	switch rhs.GetType() {
	case values.ArrayValue:
		rhsArray := rhs.(*values.Array)
		var result int64 = 0
		if len(lhs.Keys) != len(rhsArray.Keys) {
			if len(lhs.Keys) < len(rhsArray.Keys) {
				result = -1
			} else {
				result = 1
			}
		} else {
			for _, key := range lhs.Keys {
				lhsValue, _ := lhs.GetElement(key)
				rhsValue, found := rhsArray.GetElement(key)
				if found {
					equal, err := Compare(lhsValue.Value, "===", rhsValue.Value)
					if err != nil {
						return values.NewVoidSlot(), err
					}
					if equal.Value.(*values.Bool).Value {
						continue
					}
					lessThan, err := CompareRelation(lhsValue.Value, operator, rhsValue.Value, leadingNumeric)
					if err != nil {
						return values.NewVoidSlot(), err
					}
					if lessThan.GetType() == values.BoolValue {
						if lessThan.Value.(*values.Bool).Value {
							result = -1
						} else {
							result = 1
						}
					}
					if lessThan.GetType() == values.IntValue {
						result = lessThan.Value.(*values.Int).Value
					}
				}
			}
		}

		switch operator {
		case "<":
			return values.NewBoolSlot(result == -1), nil
		case "<=":
			return values.NewBoolSlot(result < 1), nil
		case "<=>":
			return values.NewIntSlot(result), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationArray: Operator "%s" not implemented`, operator)
		}

	case values.BoolValue:
		lhsBoolean, err := BoolVal(lhs)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationBoolean(values.NewBool(lhsBoolean), operator, rhs)

	case values.FloatValue, values.IntValue, values.StrValue:
		switch operator {
		case "<", "<=":
			return values.NewBoolSlot(false), nil
		case "<=>":
			return values.NewIntSlot(1), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationArray: Operator "%s" not implemented`, operator)
		}

	default:
		return values.NewVoidSlot(), phpError.NewError(`compareRelationArray: Type "%s" not implemented`, rhs.GetType())
	}
}

func compareRelationBoolean(lhs *values.Bool, operator string, rhs values.RuntimeValue) (*values.Slot, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression

	//       NULL  bool  int  float  string  array  object  resource
	// bool   <-    1     <-   <-     <-      <-     <-      <-

	//   1. If either operand has type bool, the other operand is converted to that type.
	//      The result is the logical comparison of the two operands after conversion, where FALSE is defined to be less than TRUE.

	rhsBoolean, err := BoolVal(rhs)
	if err != nil {
		return values.NewVoidSlot(), err
	}
	// TODO compareRelationBoolean - object - implement in variableHandling.BoolVal(
	// TODO compareRelationBoolean - resource - implement in variableHandling.BoolVal(
	lhsInt, err := IntVal(lhs, false)
	if err != nil {
		return values.NewVoidSlot(), err
	}
	rhsInt, err := IntVal(values.NewBool(rhsBoolean), false)
	if err != nil {
		return values.NewVoidSlot(), err
	}

	switch operator {
	case "<":
		return values.NewBoolSlot(lhsInt < rhsInt), nil

	case "<=":
		return values.NewBoolSlot(lhsInt <= rhsInt), nil

	case "<=>":
		if lhsInt > rhsInt {
			return values.NewIntSlot(1), nil
		}
		if lhsInt == rhsInt {
			return values.NewIntSlot(0), nil
		}
		return values.NewIntSlot(-1), nil

	default:
		return values.NewVoidSlot(), phpError.NewError(`compareRelationBoolean: Operator "%s" not implemented`, operator)
	}
}

func compareRelationFloating(lhs *values.Float, operator string, rhs values.RuntimeValue, leadingNumeric bool) (*values.Slot, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression

	//        NULL  bool  int  float  string  array  object  resource
	// float   <-    ->    2    2      <-      <      3       <-

	// TODO compareRelationFloating - object
	// TODO compareRelationFloating - resource

	if rhs.GetType() == values.StrValue {
		rhsStr := rhs.(*values.Str).Value
		if strings.Trim(rhsStr, " \t") == "" {
			switch operator {
			case "<", "<=":
				return values.NewBoolSlot(false), nil
			case "<=>":
				return values.NewIntSlot(1), nil
			default:
				return values.NewVoidSlot(), phpError.NewError(`compareRelationFloating: Operator "%s" not implemented for type string`, operator)
			}
		}
		if !leadingNumeric && !common.IsIntegerLiteralWithSign(rhsStr, !leadingNumeric) && !common.IsFloatingLiteralWithSign(rhsStr, !leadingNumeric) {
			switch operator {
			case "<", "<=":
				return values.NewBoolSlot(true), nil
			case "<=>":
				return values.NewIntSlot(-1), nil
			default:
				return values.NewVoidSlot(), phpError.NewError(`compareRelationFloating: Operator "%s" not implemented for type string`, operator)
			}
		}
	}

	if rhs.GetType() == values.NullValue || rhs.GetType() == values.IntValue || rhs.GetType() == values.StrValue {
		var err phpError.Error
		rhs, err = ToValueType(values.FloatValue, rhs, leadingNumeric)
		if err != nil {
			return values.NewVoidSlot(), err
		}
	}

	switch rhs.GetType() {
	case values.ArrayValue:
		switch operator {
		case "<", "<=":
			return values.NewBoolSlot(true), nil
		case "<=>":
			return values.NewIntSlot(-1), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationFloating: Operator "%s" not implemented for type array`, operator)
		}

	case values.BoolValue:
		lhsBoolean, err := BoolVal(lhs)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationBoolean(values.NewBool(lhsBoolean), operator, rhs)

	case values.FloatValue:
		rhsFloat := rhs.(*values.Float).Value
		switch operator {
		case "<":
			return values.NewBoolSlot(lhs.Value < rhsFloat), nil
		case "<=":
			return values.NewBoolSlot(lhs.Value <= rhsFloat), nil
		case "<=>":
			if lhs.Value > rhsFloat {
				return values.NewIntSlot(1), nil
			}
			if lhs.Value == rhsFloat {
				return values.NewIntSlot(0), nil
			}
			return values.NewIntSlot(-1), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationFloating: Operator "%s" not implemented`, operator)
		}

	default:
		return values.NewVoidSlot(), phpError.NewError(`compareRelationFloating: Type "%s" not implemented`, rhs.GetType())
	}
}

func compareRelationInteger(lhs *values.Int, operator string, rhs values.RuntimeValue, leadingNumeric bool) (*values.Slot, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression

	//      NULL  bool  int  float  string  array  object  resource
	// int   <-    ->    2    2      <-      <      3       <-

	// TODO compareRelationInteger - object
	// TODO compareRelationInteger - resource

	if rhs.GetType() == values.StrValue {
		rhsStr := rhs.(*values.Str).Value
		if strings.Trim(rhsStr, " \t") == "" {
			switch operator {
			case "<", "<=":
				return values.NewBoolSlot(false), nil
			case "<=>":
				return values.NewIntSlot(1), nil
			default:
				return values.NewVoidSlot(), phpError.NewError(`compareRelationInteger: Operator "%s" not implemented for type string`, operator)
			}
		}
		if !leadingNumeric && !common.IsIntegerLiteralWithSign(rhsStr, leadingNumeric) && !common.IsFloatingLiteralWithSign(rhsStr, leadingNumeric) {
			switch operator {
			case "<", "<=":
				return values.NewBoolSlot(true), nil
			case "<=>":
				return values.NewIntSlot(-1), nil
			default:
				return values.NewVoidSlot(), phpError.NewError(`compareRelationInteger: Operator "%s" not implemented for type string`, operator)
			}
		}
	}

	if rhs.GetType() == values.NullValue || rhs.GetType() == values.StrValue {
		var err phpError.Error
		rhs, err = ToValueType(values.IntValue, rhs, leadingNumeric)
		if err != nil {
			return values.NewVoidSlot(), err
		}
	}

	switch rhs.GetType() {
	case values.ArrayValue:
		switch operator {
		case "<", "<=":
			return values.NewBoolSlot(true), nil
		case "<=>":
			return values.NewIntSlot(-1), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationInteger: Operator "%s" not implemented for type array`, operator)
		}

	case values.BoolValue:
		lhsBoolean, err := BoolVal(lhs)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationBoolean(values.NewBool(lhsBoolean), operator, rhs)

	case values.FloatValue:
		lhsFloat, err := FloatVal(lhs, true)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationFloating(values.NewFloat(lhsFloat), operator, rhs, leadingNumeric)

	case values.IntValue:
		rhsInt := rhs.(*values.Int).Value
		switch operator {
		case "<":
			return values.NewBoolSlot(lhs.Value < rhsInt), nil
		case "<=":
			return values.NewBoolSlot(lhs.Value <= rhsInt), nil
		case "<=>":
			if lhs.Value > rhsInt {
				return values.NewIntSlot(1), nil
			}
			if lhs.Value == rhsInt {
				return values.NewIntSlot(0), nil
			}
			return values.NewIntSlot(-1), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationInteger: Operator "%s" not implemented`, operator)
		}

	default:
		return values.NewVoidSlot(), phpError.NewError(`compareRelationInteger: Type "%s" not implemented`, rhs.GetType())
	}
}

func compareRelationNull(operator string, rhs values.RuntimeValue, leadingNumeric bool) (*values.Slot, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression

	//       NULL  bool  int  float  string  array  object  resource
	// NULL   =     ->    ->   ->     ->      ->     <       <

	// "=" means the result is always “equals”, i.e. strict comparisons are always FALSE and equality comparisons are always TRUE.

	switch rhs.GetType() {
	case values.ArrayValue:
		lhs, err := ArrayVal(values.NewNull())
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationArray(lhs, operator, rhs, leadingNumeric)

	case values.BoolValue:
		lhs, err := BoolVal(values.NewNull())
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationBoolean(values.NewBool(lhs), operator, rhs)

	case values.FloatValue:
		lhs, err := FloatVal(values.NewNull(), false)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationFloating(values.NewFloat(lhs), operator, rhs, leadingNumeric)

	case values.IntValue:
		lhs, err := IntVal(values.NewNull(), false)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationInteger(values.NewInt(lhs), operator, rhs, leadingNumeric)

	case values.NullValue:
		switch operator {
		case "<":
			return values.NewBoolSlot(false), nil
		case "<=":
			return values.NewBoolSlot(true), nil
		case "<=>":
			return values.NewIntSlot(0), nil
		}
		return values.NewVoidSlot(), phpError.NewError(`compareRelationNull: Operator "%s" not implemented for type NULL`, operator)

		// TODO compareRelationNull - object
		// TODO compareRelationNull - resource

	case values.StrValue:
		lhs, err := StrVal(values.NewNull())
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationString(values.NewStr(lhs), operator, rhs, leadingNumeric)

	default:
		return values.NewVoidSlot(), phpError.NewError(`compareRelationNull: Type "%s" not implemented`, rhs.GetType())
	}
}

func compareRelationString(lhs *values.Str, operator string, rhs values.RuntimeValue, leadingNumeric bool) (*values.Slot, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression

	//         NULL  bool  int  float  string  array  object  resource
	// string   <-    ->    ->   ->     2, 4    <      3       2

	// TODO compareRelationString - object
	// TODO compareRelationString - resource

	if rhs.GetType() == values.FloatValue || rhs.GetType() == values.IntValue {
		lhsStr := lhs.Value
		if strings.Trim(lhsStr, " \t") == "" {
			switch operator {
			case "<", "<=":
				return values.NewBoolSlot(true), nil
			case "<=>":
				return values.NewIntSlot(-1), nil
			default:
				return values.NewVoidSlot(), phpError.NewError(`compareRelationInteger: Operator "%s" not implemented for type array`, operator)
			}
		}
		if !leadingNumeric && !common.IsIntegerLiteralWithSign(lhsStr, leadingNumeric) && !common.IsFloatingLiteralWithSign(lhsStr, leadingNumeric) {
			switch operator {
			case "<", "<=":
				return values.NewBoolSlot(false), nil
			case "<=>":
				return values.NewIntSlot(1), nil
			default:
				return values.NewVoidSlot(), phpError.NewError(`compareRelationInteger: Operator "%s" not implemented for type array`, operator)
			}
		}
	}

	if rhs.GetType() == values.NullValue {
		var err phpError.Error
		rhs, err = ToValueType(values.StrValue, rhs, true)
		if err != nil {
			return values.NewVoidSlot(), err
		}
	}

	switch rhs.GetType() {
	case values.ArrayValue:
		switch operator {
		case "<", "<=":
			return values.NewBoolSlot(true), nil
		case "<=>":
			return values.NewIntSlot(-1), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationArray: Operator "%s" not implemented`, operator)
		}

	case values.BoolValue:
		lhs, err := BoolVal(lhs)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationBoolean(values.NewBool(lhs), operator, rhs)

	case values.FloatValue:
		lhs, err := FloatVal(lhs, true)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationFloating(values.NewFloat(lhs), operator, rhs, leadingNumeric)

	case values.IntValue:
		lhs, err := IntVal(lhs, true)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationInteger(values.NewInt(lhs), operator, rhs, leadingNumeric)

	case values.StrValue:
		// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression
		//   2. If one of the operands [...] is a [...] numeric string,
		//      which can be represented as int or float without loss of precision,
		//      the operands are converted to the corresponding arithmetic type, with float taking precedence over int,
		//      and resources converting to int. The result is the numerical comparison of the two operands after conversion.
		rhsStr := rhs.(*values.Str).Value
		if common.IsFloatingLiteralWithSign(lhs.Value, false) && (common.IsIntegerLiteralWithSign(rhsStr, false) || common.IsFloatingLiteralWithSign(rhsStr, false)) {
			lhs, err := FloatVal(lhs, false)
			if err != nil {
				return values.NewVoidSlot(), err
			}
			return compareRelationFloating(values.NewFloat(lhs), operator, rhs, leadingNumeric)
		}
		if common.IsIntegerLiteralWithSign(lhs.Value, false) && (common.IsIntegerLiteralWithSign(rhsStr, false) || common.IsFloatingLiteralWithSign(rhsStr, false)) {
			lhs, err := IntVal(lhs, false)
			if err != nil {
				return values.NewVoidSlot(), err
			}
			return compareRelationInteger(values.NewInt(lhs), operator, rhs, leadingNumeric)
		}

		// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression
		//   4. If both operands are non-numeric strings, the result is the lexical comparison of the two operands.
		//      Specifically, the strings are compared byte-by-byte starting with their first byte.
		//      If the two bytes compare equal and there are no more bytes in either string, the strings are equal and the comparison ends;
		//      otherwise, if this is the final byte in one string, the shorter string compares less-than the longer string and the comparison ends.
		//      If the two bytes compare unequal, the string having the lower-valued byte compares less-than the other string, and the comparison ends.
		//      If there are more bytes in the strings, the process is repeated for the next pair of bytes.
		var result int64 = 0
		for index, lhsByte := range []byte(lhs.Value) {
			if index >= len(rhsStr) {
				result = 1
				break
			}
			rhsByte := rhsStr[index]
			if lhsByte > rhsByte {
				result = 1
				break
			}
			if lhsByte < rhsByte {
				result = -1
				break
			}
		}
		if result == 0 && len(lhs.Value) < len(rhsStr) {
			result = -1
		}
		switch operator {
		case "<":
			return values.NewBoolSlot(result == -1), nil
		case "<=":
			return values.NewBoolSlot(result < 1), nil
		case "<=>":
			return values.NewIntSlot(result), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationString: Operator "%s" not implemented`, operator)
		}

	default:
		return values.NewVoidSlot(), phpError.NewError(`compareRelationString: Type "%s" not implemented`, rhs.GetType())
	}
}

func compareRelationObject(lhs *values.Object, operator string, rhs values.RuntimeValue) (*values.Slot, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression
	//         NULL  bool  int  float  string  array  object  resource
	// object   >     ->    3    3      3       3      6       3

	switch rhs.GetType() {
	case values.NullValue:
		switch operator {
		case "<", "<=":
			return values.NewBoolSlot(false), nil
		case "<=>":
			return values.NewIntSlot(1), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationObject: Operator "%s" not implemented for type null`, operator)
		}

	case values.ObjectValue:
		// TODO 6. When comparing two objects, if any of the object types has its own compare semantics, that would define the result, with the left operand taking precedence. Otherwise, if the objects are of different types, the comparison result is FALSE. If the objects are of the same type, the properties of the objects are compares using the array comparison described above.
		switch operator {
		case "<=>":
			if lhs == rhs.(*values.Object) {
				return values.NewIntSlot(0), nil
			}
			return values.NewIntSlot(-1), nil
		default:
			return values.NewVoidSlot(), phpError.NewError(`compareRelationObject: Operator "%s" not implemented`, operator)
		}

	case values.BoolValue:
		lhs, err := BoolVal(lhs)
		if err != nil {
			return values.NewVoidSlot(), err
		}
		return compareRelationBoolean(values.NewBool(lhs), operator, rhs)

	// TODO compareRelationObject - int
	// TODO compareRelationObject - float
	// TODO compareRelationObject - string
	// TODO compareRelationObject - array
	// TODO compareRelationObject - resource

	default:
		return values.NewVoidSlot(), phpError.NewError(`compareRelationObject: Type "%s" not implemented`, rhs.GetType())
	}
}

// TODO compareRelationResource
// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression
//           NULL  bool  int  float  string  array  object  resource
// resource   >     ->    ->   ->     2       <      3       2

// -------------------------------------- comparison -------------------------------------- MARK: comparison

func Compare(lhs values.RuntimeValue, operator string, rhs values.RuntimeValue) (*values.Slot, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-equality-expression

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-equality-expression
	// Operator == represents value equality, operators != and <> are equivalent and represent value inequality.
	// For operators ==, !=, and <>, the operands of different types are converted and compared according to the same rules as in relational operators.
	// Two objects of different types are always not equal.
	if operator == "<>" {
		operator = "!="
	}
	if operator == "==" || operator == "!=" {
		resultRuntimeValue, err := CompareRelation(lhs, "<=>", rhs, false)
		if err != nil {
			return values.NewSlot(values.NewBool(false)), err
		}
		result := resultRuntimeValue.Value.(*values.Int).Value == 0

		if operator == "!=" {
			return values.NewSlot(values.NewBool(!result)), nil
		} else {
			return values.NewSlot(values.NewBool(result)), nil
		}
	}

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-equality-expression
	// Operator === represents same type and value equality, or identity, comparison,
	// and operator !== represents the opposite of ===.
	// The values are considered identical if they have the same type and compare as equal, with the additional conditions below:
	//    When comparing two objects, identity operators check to see if the two operands are the exact same object,
	//    not two different objects of the same type and value.
	//    Arrays must have the same elements in the same order to be considered identical.
	//    Strings are identical if they contain the same characters, unlike value comparison operators no conversions are performed for numeric strings.
	if operator == "===" || operator == "!==" {
		result := lhs.GetType() == rhs.GetType()
		if result {
			switch lhs.GetType() {
			case values.ArrayValue:
				lhsArray := lhs.(*values.Array)
				rhsArray := rhs.(*values.Array)
				if len(lhsArray.Keys) != len(rhsArray.Keys) {
					result = false
				} else {
					for _, key := range lhsArray.Keys {
						lhsValue, found := lhsArray.GetElement(key)
						if !found {
							result = false
							break
						}
						rhsValue, found := rhsArray.GetElement(key)
						if !found {
							result = false
							break
						}
						equal, err := Compare(lhsValue.Value, "===", rhsValue.Value)
						if err != nil {
							return values.NewSlot(values.NewBool(false)), err
						}
						if !equal.Value.(*values.Bool).Value {
							result = false
							break
						}
					}
				}
			case values.BoolValue:
				result = lhs.(*values.Bool).Value == rhs.(*values.Bool).Value
			case values.FloatValue:
				result = lhs.(*values.Float).Value == rhs.(*values.Float).Value
			case values.IntValue:
				result = lhs.(*values.Int).Value == rhs.(*values.Int).Value
			case values.NullValue:
				result = true
			case values.StrValue:
				result = lhs.(*values.Str).Value == rhs.(*values.Str).Value
			case values.ObjectValue:
				result = lhs.(*values.Object) == rhs.(*values.Object)
			default:
				return values.NewSlot(values.NewBool(false)), phpError.NewError(`compare: Runtime type %s for operator "===" not implemented`, lhs.GetType())
			}
		}

		if operator == "!==" {
			return values.NewSlot(values.NewBool(!result)), nil
		} else {
			return values.NewSlot(values.NewBool(result)), nil
		}
	}

	return values.NewSlot(values.NewBool(false)), phpError.NewError(`compare: Operator "%s" not implemented`, operator)
}
