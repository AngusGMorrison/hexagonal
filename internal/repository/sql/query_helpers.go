package sql

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// SliceToInArg takes a slice of strings, ints or uints and attempts to convert
// it to a comma-separated string argument for the SQL IN operator.
func SliceToInArg(slice any) (string, error) {
	value := reflect.ValueOf(slice)
	sliceKind := value.Kind()
	if sliceKind != reflect.Slice {
		return "", SliceKindError{kind: sliceKind}
	}

	strs := make([]string, value.Len())

	switch elemKind := value.Elem().Kind(); elemKind {
	case reflect.String:
		for i := 0; i < value.Len(); i++ {
			strs[i] = value.Index(i).String()
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for i := 0; i < value.Len(); i++ {
			i64 := value.Index(i).Int()
			strs[i] = strconv.FormatInt(i64, 10)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		for i := 0; i < value.Len(); i++ {
			u64 := value.Index(i).Uint()
			strs[i] = strconv.FormatUint(u64, 10)
		}
	default:
		return "", ElemKindError{kind: elemKind}
	}

	return strings.Join(strs, ", "), nil
}

type SliceKindError struct {
	kind reflect.Kind
}

func (ske SliceKindError) Error() string {
	return fmt.Sprintf("argument must be a slice; received kind %s", ske.kind)
}

type ElemKindError struct {
	kind reflect.Kind
}

func (eke ElemKindError) Error() string {
	return fmt.Sprintf("slices of kind %s are not supported", eke.kind)
}
