package tst

import (
	"fmt"
	"reflect"
)

// ---

type errNumberOfValuesToTestDiffers struct {
	actual   int
	expected int
}

func (e errNumberOfValuesToTestDiffers) Error() string {
	return fmt.Sprintf("number of values to test is %d but expected to be %d", e.actual, e.expected)
}

// ---

type errUnexpectedValueType struct {
	index    int
	actual   string
	expected string
}

func (e errUnexpectedValueType) Error() string {
	return fmt.Sprintf("value to test #%d is expected to have type <%s> but it has type <%s>", e.index, e.expected, e.actual)
}

// ---

func typeOf[V any](v V) string {
	if any(v) == nil {
		return reflect.TypeOf(&v).Elem().String()
	}

	return reflect.TypeOf(v).String()
}
