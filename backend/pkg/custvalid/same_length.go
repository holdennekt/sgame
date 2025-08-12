package custvalid

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

const SameLength = "same_length"

func ValidateSameLength(fl validator.FieldLevel) bool {
	sliceVal := fl.Field()
	if sliceVal.Kind() != reflect.Slice && sliceVal.Kind() != reflect.Array {
		return false
	}

	fieldName := fl.Param()
	if fieldName == "" {
		return false
	}

	first := sliceVal.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}
	field := first.FieldByName(fieldName)
	if field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
		return false
	}
	expectedLength := field.Len()

	for i := 1; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		field := elem.FieldByName(fieldName)
		if field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
			return false
		}
		if field.Len() != expectedLength {
			return false
		}
	}
	return true
}
