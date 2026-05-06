package values

import (
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/config"
	"QIQ/cmd/qiq/phpError"
	"fmt"
	"strconv"
)

type Array struct {
	*abstractValue
	Keys     []RuntimeValue
	Elements map[string]*Slot
	// Keeping track of next key
	nextKey    int64
	nextKeySet bool
}

func NewArray() *Array {
	return &Array{
		abstractValue: newAbstractValue(ArrayValue),
		Keys:          []RuntimeValue{},
		Elements:      map[string]*Slot{},
	}
}

func NewArraySlot() *Slot { return NewSlot(NewArray()) }

func NewArrayFromMap(elements map[RuntimeValue]RuntimeValue) *Array {
	array := NewArray()
	for key, value := range elements {
		if err := array.SetElement(key, value); err != nil && config.IsDevMode {
			fmt.Println("NewArrayFromMap: " + err.Error())
		}
	}
	return array
}

func NewArrayFromSlice(elements []RuntimeValue) *Array {
	array := NewArray()
	for _, value := range elements {
		if err := array.SetElement(nil, value); err != nil && config.IsDevMode {
			fmt.Println("NewArrayFromSlice: " + err.Error())
		}
	}
	return array
}

func (array *Array) SetElement(key RuntimeValue, value RuntimeValue) phpError.Error {
	key, err := array.getNextKey(key)
	if err != nil {
		return err
	}

	mapKey, found, err := array.GetMapKey(key, false)
	if err != nil {
		return err
	}

	if !found {
		array.Keys = append(array.Keys, key)
	}
	array.Elements[mapKey] = NewSlot(value)

	return nil
}

func (array *Array) GetMapKey(key RuntimeValue, shouldConvertKey bool) (string, bool, phpError.Error) {
	if shouldConvertKey {
		var err phpError.Error
		key, err = convertKey(key)
		if err != nil {
			return "", false, err
		}
	}

	mapKey, err := keyToMapKey(key)
	if err != nil {
		return "", false, err
	}

	_, found := array.Elements[mapKey]

	return mapKey, found, nil
}

func (array *Array) IsEmpty() bool { return len(array.Elements) == 0 }

func (array *Array) Contains(key RuntimeValue) bool {
	_, found, err := array.GetMapKey(key, true)
	if err != nil && config.IsDevMode {
		fmt.Println("Array.Contains: " + err.Error())
		return false
	}
	return found
}

func (array *Array) GetElement(key RuntimeValue) (*Slot, bool) {
	mapKey, found, err := array.GetMapKey(key, true)
	if err != nil || !found {
		return nil, false
	}
	return array.Elements[mapKey], true
}

func keyToMapKey(key RuntimeValue) (string, phpError.Error) {
	if key.GetType() == IntValue {
		return strconv.FormatInt(key.(*Int).Value, 10), nil
	}
	if key.GetType() == StrValue {
		return key.(*Str).Value, nil
	}
	return "", phpError.NewError("Array.keyToMapKey: Unsupported key type %s", ToPhpType(key))
}

func convertKey(key RuntimeValue) (RuntimeValue, phpError.Error) {
	if key == nil {
		return NewVoid(), nil
	}

	if key.GetType() == ArrayValue {
		// Spec: https://www.php.net/manual/en/language.types.array.php
		// Arrays and objects can not be used as keys. Doing so will result in a warning: Illegal offset type.
		return NewVoid(), phpError.NewWarning("Illegal offset type %s", ToPhpType(key))
	}

	if key.GetType() == BoolValue {
		// Spec: https://www.php.net/manual/en/language.types.array.php
		// Bools are cast to ints, too, i.e. the key true will actually be stored under 1 and the key false under 0.
		if key.(*Bool).Value {
			return NewInt(1), nil
		}
		return NewInt(0), nil
	}

	if key.GetType() == FloatValue {
		// Spec: https://www.php.net/manual/en/language.types.array.php
		// Floats are also cast to ints, which means that the fractional part will be truncated. E.g. the key 8.7 will actually be stored under 8.
		return NewInt(int64(key.(*Float).Value)), nil
	}

	if key.GetType() == StrValue && common.IsDecimalLiteral(key.(*Str).Value, false) {
		// Spec: https://www.php.net/manual/en/language.types.array.php
		// Strings containing valid decimal ints, unless the number is preceded by a + sign, will be cast to the int type.
		// E.g. the key "8" will actually be stored under 8. On the other hand "08" will not be cast, as it isn't a valid decimal integer.
		return NewInt(common.DecimalLiteralToInt64(key.(*Str).Value, false)), nil
	}

	if key.GetType() == NullValue {
		// Spec: https://www.php.net/manual/en/language.types.array.php
		// Null will be cast to the empty string, i.e. the key null will actually be stored under "".
		return NewStr(""), nil
	}

	return key, nil
}

func (array *Array) getNextKey(key RuntimeValue) (RuntimeValue, phpError.Error) {
	// If a key is passed
	if key != nil {
		var err phpError.Error
		key, err = convertKey(key)
		if err != nil {
			return key, err
		}

		// If the passed key is an integer after the convertion
		if key.GetType() == IntValue {
			keyValue := key.(*Int).Value
			// If no key is stored yet or the passed key is greater than nextKey
			if !array.nextKeySet ||
				(array.nextKeySet && keyValue > array.nextKey) {
				// Store value + 1 as next key
				array.nextKey = keyValue + 1
				array.nextKeySet = true
			}
		}

		return key, nil
	}

	// If no key is passed and no nextKey is set yet
	if !array.nextKeySet {
		// lastFoundInt is -1 because at the end of the loop it will be increased by one
		var lastFoundInt int64 = -1
		found := false
		// Iterate all keys and search for integer values
		for i := len(array.Keys) - 1; i >= 0; i-- {
			if array.Keys[i].GetType() != IntValue {
				continue
			}

			foundInt := array.Keys[i].(*Int).Value
			if !found {
				lastFoundInt = foundInt
				found = true
				continue
			}

			if foundInt > lastFoundInt {
				lastFoundInt = foundInt
			}
		}
		array.nextKey = lastFoundInt + 1
		array.nextKeySet = true
	}

	key = NewInt(array.nextKey)
	array.nextKey++
	return key, nil
}
