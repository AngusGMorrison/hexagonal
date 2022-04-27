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
	sliceType := reflect.TypeOf(slice)
	sliceKind := sliceType.Kind()
	if sliceKind != reflect.Slice {
		return "", SliceKindError{kind: sliceKind}
	}

	sliceValue := reflect.ValueOf(slice)
	strs := make([]string, sliceValue.Len())

	switch elemKind := sliceType.Elem().Kind(); elemKind {
	case reflect.String:
		for i := 0; i < sliceValue.Len(); i++ {
			strs[i] = sliceValue.Index(i).String()
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for i := 0; i < sliceValue.Len(); i++ {
			i64 := sliceValue.Index(i).Int()
			strs[i] = strconv.FormatInt(i64, 10)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		for i := 0; i < sliceValue.Len(); i++ {
			u64 := sliceValue.Index(i).Uint()
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
