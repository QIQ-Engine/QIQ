package variableHandling

import (
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/funcParamValidator"
	"QIQ/cmd/qiq/runtime/values"
	"fmt"
	"math"
	"slices"
	"strings"
)

func Register(environment runtime.Environment) {
	// Category: Variable Handling Functions
	environment.AddNativeFunction("boolval", nativeFn_boolval)
	environment.AddNativeFunction("doubleval", nativeFn_floatval)
	environment.AddNativeFunction("floatval", nativeFn_floatval)
	environment.AddNativeFunction("get_debug_type", nativeFn_get_debug_type)
	environment.AddNativeFunction("gettype", nativeFn_gettype)
	environment.AddNativeFunction("intval", nativeFn_intval)
	environment.AddNativeFunction("is_array", nativeFn_is_array)
	environment.AddNativeFunction("is_bool", nativeFn_is_bool)
	environment.AddNativeFunction("is_double", nativeFn_is_float)
	environment.AddNativeFunction("is_float", nativeFn_is_float)
	environment.AddNativeFunction("is_int", nativeFn_is_int)
	environment.AddNativeFunction("is_integer", nativeFn_is_int)
	environment.AddNativeFunction("is_long", nativeFn_is_int)
	environment.AddNativeFunction("is_null", nativeFn_is_null)
	environment.AddNativeFunction("is_object", nativeFn_is_object)
	environment.AddNativeFunction("is_scalar", nativeFn_is_scalar)
	environment.AddNativeFunction("is_string", nativeFn_is_string)
	environment.AddNativeFunction("print_r", nativeFn_print_r)
	environment.AddNativeFunction("serialize", nativeFn_serialize)
	environment.AddNativeFunction("strval", nativeFn_strval)
	environment.AddNativeFunction("unserialize", nativeFn_unserialize)
	environment.AddNativeFunction("var_dump", nativeFn_var_dump)
	environment.AddNativeFunction("var_export", nativeFn_var_export)
}

// -------------------------------------- arrayval -------------------------------------- MARK: arrayval

// This is not an official function. But converting different types to array is needed in several places
func ArrayVal(runtimeValue values.RuntimeValue) (*values.Array, phpError.Error) {
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-array-type

	// The result type is array.

	if runtimeValue.GetType() == values.NullValue {
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-array-type
		// If the source value is NULL, the result value is an array of zero elements.
		return values.NewArray(), nil
	}

	// TODO ArrayVal - resource
	if IsScalar(runtimeValue) {
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-array-type
		// If the source type is scalar or resource and it is non-NULL,
		// the result value is an array of one element under the key 0 whose value is that of the source.
		array := values.NewArray()
		array.SetElement(nil, runtimeValue)
		return array, nil
	}

	// TODO ArrayVal - object
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-array-type
	// If the source is an object, the result is an array of zero or more elements, where the elements are key/value pairs corresponding to the object’s instance properties. The order of insertion of the elements into the array is the lexical order of the instance properties in the class-member-declarations list.

	// TODO ArrayVal - instance properties
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-array-type
	// For public instance properties, the keys of the array elements would be the same as the property name.
	// The key for a private instance property has the form “\0class\0name”, where the class is the class name, and the name is the property name.
	// The key for a protected instance property has the form “\0*\0name”, where name is that of the property.
	// The value for each key is that from the corresponding property, or NULL if the property was not initialized.

	return values.NewArray(), phpError.NewError("ArrayVal: Unsupported type %s", runtimeValue.GetType())
}

// -------------------------------------- boolval -------------------------------------- MARK: boolval

func nativeFn_boolval(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("boolval").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	boolean, err := BoolVal(args[0])
	return values.NewBool(boolean), err
}

func BoolVal(runtimeValue values.RuntimeValue) (bool, phpError.Error) {
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-boolean-type
	// Spec: https://www.php.net/manual/en/function.boolval.php

	switch runtimeValue.GetType() {
	case values.ArrayValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-boolean-type
		// If the source is an array with zero elements, the result value is FALSE; otherwise, the result value is TRUE.
		return len(runtimeValue.(*values.Array).Elements) != 0, nil
	case values.BoolValue:
		return runtimeValue.(*values.Bool).Value, nil
	case values.IntValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-boolean-type
		// If the source type is int or float,
		// then if the source value tests equal to 0, the result value is FALSE; otherwise, the result value is TRUE.
		return runtimeValue.(*values.Int).Value != 0, nil
	case values.FloatValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-boolean-type
		// If the source type is int or float,
		// then if the source value tests equal to 0, the result value is FALSE; otherwise, the result value is TRUE.
		return math.Abs(runtimeValue.(*values.Float).Value) != 0, nil
	case values.NullValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-boolean-type
		// If the source value is NULL, the result value is FALSE.
		return false, nil
	case values.StrValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-boolean-type
		// If the source is an empty string or the string “0”, the result value is FALSE; otherwise, the result value is TRUE.
		str := runtimeValue.(*values.Str).Value
		return str != "" && str != "0", nil
	case values.ObjectValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-boolean-type
		// If the source is an object, the result value is TRUE.
		return true, nil
	default:
		return false, phpError.NewError("boolval: Unsupported runtime value %s", runtimeValue.GetType())
	}

	// TODO boolval - resource
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-boolean-type
	// If the source is a resource, the result value is TRUE.
}

// -------------------------------------- floatval -------------------------------------- MARK: floatval

func nativeFn_floatval(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("floatval").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	floating, err := FloatVal(args[0], true)
	return values.NewFloat(floating), err
}

func FloatVal(runtimeValue values.RuntimeValue, leadingNumeric bool) (float64, phpError.Error) {
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-floating-point-type
	// Spec: https://www.php.net/manual/en/function.floatval.php

	switch runtimeValue.GetType() {
	case values.FloatValue:
		return runtimeValue.(*values.Float).Value, nil
	case values.IntValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-floating-point-type
		// If the source type is int,
		// if the precision can be preserved the result value is the closest approximation to the source value;
		// otherwise, the result is undefined.
		return float64(runtimeValue.(*values.Int).Value), nil
	case values.StrValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-floating-point-type
		// If the source is a numeric string or leading-numeric string having integer format,
		// the string’s integer value is treated as described above for a conversion from int.
		// If the source is a numeric string or leading-numeric string having floating-point format,
		// the result value is the closest approximation to the string’s floating-point value.
		// The trailing non-numeric characters in leading-numeric strings are ignored.
		// For any other string, the result value is 0.
		intStr := runtimeValue.(*values.Str).Value
		if common.IsFloatingLiteralWithSign(intStr, leadingNumeric) {
			return common.FloatingLiteralToFloat64WithSign(intStr, leadingNumeric), nil
		}
		if common.IsIntegerLiteralWithSign(intStr, leadingNumeric) {
			intValue, err := common.IntegerLiteralToInt64WithSign(intStr, leadingNumeric)
			if err != nil {
				return 0, phpError.NewError("%s", err.Error())
			}
			return FloatVal(values.NewInt(intValue), leadingNumeric)
		}
		return 0, nil
	default:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-floating-point-type
		// For sources of all other types, the conversion result is obtained by first converting
		// the source value to int and then to float.
		intValue, err := IntVal(runtimeValue, leadingNumeric)
		if err != nil {
			return 0, err
		}
		return FloatVal(values.NewInt(intValue), leadingNumeric)
	}

	// TODO FloatVal - object
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-floating-point-type
	// If the source is an object, if the class defines a conversion function, the result is determined by that function (this is currently available only to internal classes). If not, the conversion is invalid, the result is assumed to be 1.0 and a non-fatal error is produced.
}

// -------------------------------------- get_debug_type -------------------------------------- MARK: get_debug_type

func nativeFn_get_debug_type(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("get_debug_type").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	typeStr, err := lib_get_debug_type(args[0])
	return values.NewStr(typeStr), err
}

func lib_get_debug_type(runtimeValue values.RuntimeValue) (string, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.get-debug-type

	// TODO lib_get_debug_type - object
	// TODO lib_get_debug_type - resource
	// TODO lib_get_debug_type - resource (closed)
	switch runtimeValue.GetType() {
	case values.ArrayValue:
		return "array", nil
	case values.BoolValue:
		return "bool", nil
	case values.FloatValue:
		return "float", nil
	case values.IntValue:
		return "int", nil
	case values.NullValue:
		return "null", nil
	case values.StrValue:
		return "string", nil
	default:
		return "unknown type", nil
	}
}

// -------------------------------------- gettype -------------------------------------- MARK: gettype

func nativeFn_gettype(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("gettype").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	typeStr, err := GetType(args[0])
	return values.NewStr(typeStr), err
}

func GetType(runtimeValue values.RuntimeValue) (string, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.gettype.php

	// TODO GetType - resource
	// TODO GetType - resource (closed)
	switch runtimeValue.GetType() {
	case values.ArrayValue:
		return "array", nil
	case values.BoolValue:
		return "boolean", nil
	case values.FloatValue:
		return "double", nil
	case values.IntValue:
		return "integer", nil
	case values.NullValue:
		return "NULL", nil
	case values.StrValue:
		return "string", nil
	case values.ObjectValue:
		return "object", nil
	default:
		return "unknown type", nil
	}
}

// -------------------------------------- intval -------------------------------------- MARK: intval

func nativeFn_intval(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("intval").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	integer, err := IntVal(args[0], true)
	return values.NewInt(integer), err
}

func IntVal(runtimeValue values.RuntimeValue, leadingNumeric bool) (int64, phpError.Error) {
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-integer-type

	switch runtimeValue.GetType() {
	case values.ArrayValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-integer-type
		// If the source is an array with zero elements, the result value is 0; otherwise, the result value is 1.
		if len(runtimeValue.(*values.Array).Elements) == 0 {
			return 0, nil
		}
		return 1, nil
	case values.BoolValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-integer-type
		// If the source type is bool, then if the source value is FALSE, the result value is 0; otherwise, the result value is 1.
		if runtimeValue.(*values.Bool).Value {
			return 1, nil
		}
		return 0, nil
	case values.FloatValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-integer-type
		// If the source type is float, for the values INF, -INF, and NAN, the result value is zero.
		// For all other values, if the precision can be preserved (that is, the float is within the range of an integer),
		// the fractional part is rounded towards zero.
		return int64(runtimeValue.(*values.Float).Value), nil
	case values.IntValue:
		return runtimeValue.(*values.Int).Value, nil
	case values.NullValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-integer-type
		// If the source value is NULL, the result value is 0.
		return 0, nil
	case values.StrValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-integer-type
		// If the source is a numeric string or leading-numeric string having integer format,
		// if the precision can be preserved the result value is that string’s integer value;
		// otherwise, the result is undefined.
		// If the source is a numeric string or leading-numeric string having floating-point format,
		// the string’s floating-point value is treated as described above for a conversion from float.
		// The trailing non-numeric characters in leading-numeric strings are ignored.
		// For any other string, the result value is 0.
		intStr := runtimeValue.(*values.Str).Value
		if common.IsFloatingLiteralWithSign(intStr, leadingNumeric) {
			return IntVal(values.NewFloat(common.FloatingLiteralToFloat64WithSign(intStr, leadingNumeric)), leadingNumeric)
		}
		if common.IsIntegerLiteralWithSign(intStr, leadingNumeric) {
			intValue, err := common.IntegerLiteralToInt64WithSign(intStr, leadingNumeric)
			if err != nil {
				return 0, phpError.NewError("%s", err.Error())
			}
			return intValue, nil
		}
		return 0, nil
	default:
		return 0, phpError.NewError("IntVal: Unsupported runtime value %s", runtimeValue.GetType())
	}

	// TODO IntVal - object
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-integer-type
	// If the source is an object, if the class defines a conversion function, the result is determined by that function (this is currently available only to internal classes). If not, the conversion is invalid, the result is assumed to be 1 and a non-fatal error is produced.

	// TODO IntVal - resource
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-integer-type
	// If the source is a resource, the result is the resource’s unique ID.
}

// -------------------------------------- is_array -------------------------------------- MARK: is_array

func nativeFn_is_array(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("is_array").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewBool(lib_is_array(args[0])), nil
}

func lib_is_array(runtimeValue values.RuntimeValue) bool {
	// Spec: https://www.php.net/manual/en/function.is-array
	return runtimeValue.GetType() == values.ArrayValue
}

// -------------------------------------- is_bool -------------------------------------- MARK: is_bool

func nativeFn_is_bool(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("is_bool").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewBool(lib_is_bool(args[0])), nil
}

func lib_is_bool(runtimeValue values.RuntimeValue) bool {
	// Spec: https://www.php.net/manual/en/function.is-bool
	return runtimeValue.GetType() == values.BoolValue
}

// -------------------------------------- is_float -------------------------------------- MARK: is_float

func nativeFn_is_float(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("is_float").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewBool(lib_is_float(args[0])), nil
}

func lib_is_float(runtimeValue values.RuntimeValue) bool {
	// Spec: https://www.php.net/manual/en/function.is-float
	return runtimeValue.GetType() == values.FloatValue
}

// -------------------------------------- is_int -------------------------------------- MARK: is_int

func nativeFn_is_int(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("is_int").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewBool(lib_is_int(args[0])), nil
}

func lib_is_int(runtimeValue values.RuntimeValue) bool {
	// Spec: https://www.php.net/manual/en/function.is-int
	return runtimeValue.GetType() == values.IntValue
}

// -------------------------------------- is_null -------------------------------------- MARK: is_null

func nativeFn_is_null(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.is-null.php

	args, err := funcParamValidator.NewValidator("is_null").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewBool(args[0].GetType() == values.NullValue), nil
}

// -------------------------------------- is_object -------------------------------------- MARK: is_object

func nativeFn_is_object(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.is-object.php

	args, err := funcParamValidator.NewValidator("is_object").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewBool(args[0].GetType() == values.ObjectValue), nil
}

// -------------------------------------- is_scalar -------------------------------------- MARK: is_scalar

func nativeFn_is_scalar(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("is_scalar").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewBool(IsScalar(args[0])), nil
}

func IsScalar(runtimeValue values.RuntimeValue) bool {
	// Spec: https://www.php.net/manual/en/function.is-scalar.php
	return slices.Contains([]values.ValueType{values.BoolValue, values.IntValue, values.FloatValue, values.StrValue}, runtimeValue.GetType())
}

// -------------------------------------- is_string -------------------------------------- MARK: is_string

func nativeFn_is_string(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.is-string

	args, err := funcParamValidator.NewValidator("is_string").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewBool(args[0].GetType() == values.StrValue), nil
}

// -------------------------------------- print_r -------------------------------------- MARK: print_r

func nativeFn_print_r(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("print_r").
		AddParam("$value", []string{"mixed"}, nil).
		AddParam("$return", []string{"bool"}, values.NewBool(false)).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return lib_print_r(args[0], args[1].(*values.Bool).Value, context)
}

func lib_print_r(value values.RuntimeValue, returnValue bool, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	result, err := lib_print_r_var(value, 4)
	if err != nil {
		return values.NewVoid(), err
	}

	if returnValue {
		return values.NewStr(result), nil
	} else {
		context.Interpreter.Print(result)
		return values.NewBool(true), nil
	}
}

func lib_print_r_var(value values.RuntimeValue, depth int) (string, phpError.Error) {
	result := ""
	var err phpError.Error
	switch value.GetType() {
	case values.ArrayValue:
		array := value.(*values.Array)
		result = fmt.Sprintf("Array\n%s(\n", strings.Repeat(" ", depth-4))
		for _, key := range array.Keys {
			keyStr, err := lib_print_r_var(key, depth+4)
			if err != nil {
				return "", err
			}
			elementValue, _ := array.GetElement(key)
			valueStr, err := lib_print_r_var(elementValue.Value, depth+8)
			if err != nil {
				return "", err
			}

			result += fmt.Sprintf("%s[%s] => %s\n", strings.Repeat(" ", depth), keyStr, valueStr)
		}
		result += fmt.Sprintf("%s)\n", strings.Repeat(" ", depth-4))
	case values.BoolValue:
		if value.(*values.Bool).Value {
			result = "1"
		} else {
			result = ""
		}
	case values.FloatValue, values.IntValue:
		result, err = StrVal(value)
		if err != nil {
			return "", err
		}
	case values.NullValue:
		result = ""
	case values.StrValue:
		result = value.(*values.Str).Value
	case values.ObjectValue:
		object := value.(*values.Object)
		result = fmt.Sprintf("%s Object\n%s(\n", object.Class.Name, strings.Repeat(" ", depth-4))
		for _, name := range object.PropertyNames {
			value := object.Properties[name]
			valueStr, err := lib_print_r_var(value.Value, depth+8)
			if err != nil {
				return "", err
			}

			result += fmt.Sprintf("%s[%s] => %s\n", strings.Repeat(" ", depth), name[1:], valueStr)
		}
		result += fmt.Sprintf("%s)\n", strings.Repeat(" ", depth-4))
	default:
		return "", phpError.NewError("lib_print_r_var: Unsupported runtime value %s", value.GetType())
	}
	return result, nil
}

// -------------------------------------- serialize -------------------------------------- MARK: serialize

func nativeFn_serialize(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.serialize.php
	args, err := funcParamValidator.NewValidator("serialize").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	str, err := Serialize(args[0])
	return values.NewStr(str), err
}

func Serialize(runtimeValue values.RuntimeValue) (string, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.serialize.php#66147

	switch runtimeValue.GetType() {
	case values.BoolValue:
		if runtimeValue.(*values.Bool).Value {
			return "b:1;", nil
		}
		return "b:0;", nil

	case values.NullValue:
		return "N;", nil
	case values.IntValue:
		// i:<value>;
		return fmt.Sprintf("i:%d;", runtimeValue.(*values.Int).Value), nil
	case values.FloatValue:
		// d:<value>;
		return fmt.Sprintf("d:%s;", runtimeValue.(*values.Float).ToPhpString()), nil
	case values.StrValue:
		// s:<size>:"<value>";
		// String values are always in double quotes
		str := runtimeValue.(*values.Str).Value
		return fmt.Sprintf(`s:%d:"%s";`, len(str), str), nil

	case values.ArrayValue:
		// a:<size>:{<key definition><value definition>(repeated per element)}
		array := runtimeValue.(*values.Array)
		var result strings.Builder
		fmt.Fprintf(&result, "a:%d:{", len(array.Keys))
		for _, key := range array.Keys {
			// Array key
			valueStr, err := Serialize(key)
			if err != nil {
				return "", err
			}
			result.WriteString(valueStr)
			// Array value
			slot, _ := array.GetElement(key)
			valueStr, err = Serialize(slot.Value)
			if err != nil {
				return "", err
			}
			result.WriteString(valueStr)
		}
		result.WriteString("}")
		return result.String(), nil

	case values.ObjectValue:
		// O:<strlen(object name)>:"<object name>":<object size>:{<property name definition)><property value definition>(repeated per property)}
		object := runtimeValue.(*values.Object)
		var result strings.Builder
		fmt.Fprintf(&result, `O:%d:"%s":%d:{`,
			len(object.Class.GetQualifiedName()),
			object.Class.GetQualifiedName(),
			len(object.PropertyNames))
		for _, property := range object.PropertyNames {
			// Property name
			// Remove the $ prefix
			propertyName := property[1:]
			switch object.Class.Properties[property].Visibility {
			case "private":
				propertyName = "\x00" + object.Class.Name + "\x00" + propertyName
			case "protected":
				propertyName = "\x00*\x00" + propertyName
			}
			fmt.Fprintf(&result, `s:%d:"%s";`, len(propertyName), propertyName)

			// Property value
			value, _ := object.GetProperty(property)
			valueStr, err := Serialize(value)
			if err != nil {
				return "", err
			}
			result.WriteString(valueStr)
		}
		result.WriteString("}")
		return result.String(), nil

	default:
		return "", phpError.NewError("Serialize: Unsupported runtime value %s", runtimeValue.GetType())
	}
}

// -------------------------------------- strval -------------------------------------- MARK: strval

func nativeFn_strval(args []values.RuntimeValue, _ runtime.Context) (values.RuntimeValue, phpError.Error) {
	args, err := funcParamValidator.NewValidator("strval").AddParam("$value", []string{"mixed"}, nil).Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	str, err := StrVal(args[0])
	return values.NewStr(str), err
}

func StrVal(runtimeValue values.RuntimeValue) (string, phpError.Error) {
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-string-type

	switch runtimeValue.GetType() {
	case values.ArrayValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-string-type
		// If the source is an array, the conversion is invalid. The result value is the string “Array” and a non-fatal error is produced.
		return "Array", nil
		// TODO lib_strval - array non-fatal error: "Warning: Array to string conversion in /home/user/scripts/code.php on line 2" - only if E_ALL | E_WARNING
	case values.BoolValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-string-type
		// If the source type is bool, then if the source value is FALSE, the result value is the empty string;
		// otherwise, the result value is “1”.
		if !runtimeValue.(*values.Bool).Value {
			return "", nil
		}
		return "1", nil
	case values.FloatValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-string-type
		// If the source type is int or float, then the result value is a string containing the textual representation
		// of the source value (as specified by the library function sprintf).
		return runtimeValue.(*values.Float).ToPhpString(), nil
	case values.IntValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-string-type
		// If the source type is int or float, then the result value is a string containing the textual representation
		// of the source value (as specified by the library function sprintf).
		return fmt.Sprintf("%d", runtimeValue.(*values.Int).Value), nil
	case values.NullValue:
		// Spec: https://phplang.org/spec/08-conversions.html#converting-to-string-type
		// If the source value is NULL, the result value is the empty string.
		return "", nil
	case values.StrValue:
		return runtimeValue.(*values.Str).Value, nil
	case values.VoidValue:
		return "", nil
	default:
		return "", phpError.NewError("lib_strval: Unsupported runtime value %s", runtimeValue.GetType())
	}

	// TODO lib_strval - object
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-string-type
	// If the source is an object, then if that object’s class has a __toString method, the result value is the string returned by that method; otherwise, the conversion is invalid and a fatal error is produced.

	// TODO lib_strval - resource
	// Spec: https://phplang.org/spec/08-conversions.html#converting-to-string-type
	// If the source is a resource, the result value is an implementation-defined string.
}

// -------------------------------------- unserialize -------------------------------------- MARK: unserialize

func nativeFn_unserialize(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.unserialize.php
	args, err := funcParamValidator.NewValidator("unserialize").
		AddParam("$data", []string{"string"}, nil).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}
	// TODO unserialize - "array $options = []"

	return Unserialize(args[0].(*values.Str).Value)
}

func Unserialize(data string) (values.RuntimeValue, phpError.Error) {
	// Null
	if after, ok := strings.CutPrefix(data, "N;"); ok {
		data = after
		if len(data) > 0 {
			return values.NewNull(), phpError.NewWarning("Unexpected extra data")
		}

		return values.NewNull(), nil
	}

	// Boolean
	if after, ok := strings.CutPrefix(data, "b:"); ok {
		data = after
		if len(data) < 2 {
			return values.NewBool(false), phpError.NewWarning("Unserialize: Unexpected end of boolean data")
		}

		var result values.RuntimeValue
		if data[0] == '1' && data[1] == ';' {
			result = values.NewBool(true)
		} else if data[0] == '0' && data[1] == ';' {
			result = values.NewBool(false)
		} else {
			return values.NewBool(false), phpError.NewWarning("Unserialize: Invalid boolean value %q", data[0])
		}
		data = data[2:]

		if len(data) > 0 {
			return result, phpError.NewWarning("Unexpected extra data")
		}

		return result, nil
	}

	// Integer
	if after, ok := strings.CutPrefix(data, "i:"); ok {
		data = after
		semicolonIndex := strings.IndexByte(data, ';')
		if semicolonIndex == -1 {
			return values.NewBool(false), phpError.NewWarning("Unserialize: Unexpected end of integer data")
		}

		intStr := data[:semicolonIndex]
		data = data[semicolonIndex+1:]

		intStr, isNegative := strings.CutPrefix(intStr, "-")
		intValue, err := common.IntegerLiteralToInt64(intStr, false)
		if err != nil {
			return values.NewBool(false), phpError.NewWarning("Unserialize: Invalid integer value %q", intStr)
		}
		if isNegative {
			intValue = -intValue
		}

		result := values.NewInt(intValue)

		if len(data) > 0 {
			return result, phpError.NewWarning("Unexpected extra data")
		}

		return result, nil
	}

	// TODO float

	// TODO string

	// TODO array

	// TODO object

	return values.NewVoid(), phpError.NewError("Unhandeled data '%s'", data)
}

// func unserialize(data string) (values.RuntimeValue, phpError.Error)
// 	// Spec: https://www.php.net/manual/en/function.serialize.php#66147

// 	switch runtimeValue.GetType() {
// 	case values.FloatValue:
// 		// d:<value>;
// 		return fmt.Sprintf("d:%s;", runtimeValue.(*values.Float).ToPhpString()), nil
// 	case values.StrValue:
// 		// s:<size>:"<value>";
// 		// String values are always in double quotes
// 		str := runtimeValue.(*values.Str).Value
// 		return fmt.Sprintf(`s:%d:"%s";`, len(str), str), nil

// 	case values.ArrayValue:
// 		// a:<size>:{<key definition><value definition>(repeated per element)}
// 		array := runtimeValue.(*values.Array)
// 		var result strings.Builder
// 		fmt.Fprintf(&result, "a:%d:{", len(array.Keys))
// 		for _, key := range array.Keys {
// 			// Array key
// 			valueStr, err := Serialize(key)
// 			if err != nil {
// 				return "", err
// 			}
// 			result.WriteString(valueStr)
// 			// Array value
// 			slot, _ := array.GetElement(key)
// 			valueStr, err = Serialize(slot.Value)
// 			if err != nil {
// 				return "", err
// 			}
// 			result.WriteString(valueStr)
// 		}
// 		result.WriteString("}")
// 		return result.String(), nil

// 	case values.ObjectValue:
// 		// O:<strlen(object name)>:"<object name>":<object size>:{<property name definition)><property value definition>(repeated per property)}
// 		object := runtimeValue.(*values.Object)
// 		var result strings.Builder
// 		fmt.Fprintf(&result, `O:%d:"%s":%d:{`,
// 			len(object.Class.GetQualifiedName()),
// 			object.Class.GetQualifiedName(),
// 			len(object.PropertyNames))
// 		for _, property := range object.PropertyNames {
// 			// Property name
// 			// Remove the $ prefix
// 			propertyName := property[1:]
// 			switch object.Class.Properties[property].Visibility {
// 			case "private":
// 				propertyName = "\x00" + object.Class.Name + "\x00" + propertyName
// 			case "protected":
// 				propertyName = "\x00*\x00" + propertyName
// 			}
// 			fmt.Fprintf(&result, `s:%d:"%s";`, len(propertyName), propertyName)

// 			// Property value
// 			value, _ := object.GetProperty(property)
// 			valueStr, err := Serialize(value)
// 			if err != nil {
// 				return "", err
// 			}
// 			result.WriteString(valueStr)
// 		}
// 		result.WriteString("}")
// 		return result.String(), nil

// 	default:
// 		return "", phpError.NewError("Serialize: Unsupported runtime value %s", runtimeValue.GetType())
// 	}
// }

// -------------------------------------- var_dump -------------------------------------- MARK: var_dump

func nativeFn_var_dump(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.var-dump

	args, err := funcParamValidator.NewValidator("var_dump").
		AddParam("$value", []string{"mixed"}, nil).AddVariableLenParam("$values", []string{"mixed"}).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	if err := lib_var_dump_var(context, args[0], 2); err != nil {
		return values.NewVoid(), err
	}

	if len(args) == 2 {
		arrayValues := args[1].(*values.Array)
		for _, key := range arrayValues.Keys {
			argValue, _ := arrayValues.GetElement(key)
			if err := lib_var_dump_var(context, argValue.Value, 2); err != nil {
				return values.NewVoid(), err
			}
		}
	}

	return values.NewVoid(), nil
}

func lib_var_dump_var(context runtime.Context, value values.RuntimeValue, depth int) phpError.Error {
	switch value.GetType() {
	case values.ArrayValue:
		array := value.(*values.Array)
		context.Interpreter.Println(fmt.Sprintf("array(%d) {", len(array.Keys)))
		for _, key := range array.Keys {
			switch key.GetType() {
			case values.IntValue:
				keyValue := key.(*values.Int).Value
				context.Interpreter.Println(fmt.Sprintf("%s[%d]=>", strings.Repeat(" ", depth), keyValue))
			case values.StrValue:
				keyValue := key.(*values.Str).Value
				context.Interpreter.Println(fmt.Sprintf(`%s["%s"]=>`, strings.Repeat(" ", depth), keyValue))
			default:
				return phpError.NewError("lib_var_dump_var: Unsupported array key type %s", key.GetType())
			}
			context.Interpreter.Print(strings.Repeat(" ", depth))
			elementValue, _ := array.GetElement(key)
			if err := lib_var_dump_var(context, elementValue.Value, depth+2); err != nil {
				return err
			}
		}
		context.Interpreter.Println(strings.Repeat(" ", depth-2) + "}")
	case values.BoolValue:
		if value.(*values.Bool).Value {
			context.Interpreter.Println("bool(true)")
		} else {
			context.Interpreter.Println("bool(false)")
		}
	case values.FloatValue:
		strVal, err := StrVal(value)
		if err != nil {
			return err
		}
		context.Interpreter.Println("float(" + strVal + ")")
	case values.IntValue:
		strVal, err := StrVal(value)
		if err != nil {
			return err
		}
		context.Interpreter.Println("int(" + strVal + ")")
	case values.NullValue:
		context.Interpreter.Println("NULL")
	case values.StrValue:
		strVal := value.(*values.Str).Value
		context.Interpreter.Println(fmt.Sprintf(`string(%d) "%s"`, len(strVal), strVal))
	case values.ObjectValue:
		object := value.(*values.Object)
		// TODO var_dump - object: dynamic counter of instaces
		context.Interpreter.Println(fmt.Sprintf("object(%s)#1 (%d) {",
			object.Class.GetQualifiedName(), len(object.Class.PropertieNames),
		))
		for _, propertyName := range object.PropertyNames {
			property := object.Class.Properties[propertyName]
			propertyValue := object.Properties[propertyName]

			context.Interpreter.Println(fmt.Sprintf(`%s["%s":"%s":%s]=>`,
				strings.Repeat(" ", depth), propertyName[1:], object.Class.GetQualifiedName(), property.Visibility,
			))
			context.Interpreter.Print(strings.Repeat(" ", depth))
			if err := lib_var_dump_var(context, propertyValue.Value, depth+2); err != nil {
				return err
			}
		}
		context.Interpreter.Println(strings.Repeat(" ", depth-2) + "}")
	default:
		return phpError.NewError("lib_var_dump_var: Unsupported runtime value %s", value.GetType())
	}
	return nil
}

// -------------------------------------- var_export -------------------------------------- MARK: var_export

func nativeFn_var_export(args []values.RuntimeValue, interpreter runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.var-export

	args, err := funcParamValidator.NewValidator("var_dump").
		AddParam("$value", []string{"mixed"}, nil).
		AddParam("$return", []string{"bool"}, values.NewBool(false)).
		Validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return lib_var_export(args[0], args[1].(*values.Bool).Value, interpreter)
}

func lib_var_export(value values.RuntimeValue, returnValue bool, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	result, err := lib_var_export_var(value, 2)
	if err != nil {
		return values.NewVoid(), err
	}

	if returnValue {
		return values.NewStr(result), nil
	} else {
		context.Interpreter.Print(result)
		return values.NewNull(), nil
	}
}

func lib_var_export_var(value values.RuntimeValue, depth int) (string, phpError.Error) {
	result := ""
	var err phpError.Error
	switch value.GetType() {
	case values.ArrayValue:
		array := value.(*values.Array)
		result = fmt.Sprintf("%sarray (\n", strings.Repeat(" ", depth-2))
		for _, key := range array.Keys {
			keyStr, err := lib_var_export_var(key, depth+2)
			if err != nil {
				return "", err
			}
			elementValue, _ := array.GetElement(key)
			valueStr, err := lib_var_export_var(elementValue.Value, depth+2)
			if err != nil {
				return "", err
			}

			if elementValue.GetType() == values.ArrayValue {
				result += fmt.Sprintf("%s%s => \n%s%s,\n",
					strings.Repeat(" ", depth), keyStr,
					strings.Repeat(" ", depth-2), valueStr,
				)
			} else {
				result += fmt.Sprintf("%s%s => %s,\n", strings.Repeat(" ", depth), keyStr, valueStr)
			}
		}
		result += fmt.Sprintf("%s)", strings.Repeat(" ", depth-2))
	case values.BoolValue:
		if value.(*values.Bool).Value {
			result = "true"
		} else {
			result = "false"
		}
	case values.FloatValue, values.IntValue:
		result, err = StrVal(value)
		if err != nil {
			return "", err
		}
	case values.NullValue:
		result = "NULL"
	case values.StrValue:
		result = "'" + value.(*values.Str).Value + "'"
	default:
		return "", phpError.NewError("lib_var_export: Unsupported runtime value %s", value.GetType())
	}
	return result, nil
}

// TODO debug_​zval_​dump
// TODO get_​defined_​vars
// TODO get_​resource_​id
// TODO get_​resource_​type
// TODO is_​callable
// TODO is_​countable
// TODO is_​iterable
// TODO is_​numeric
// TODO is_​resource
// TODO settype
