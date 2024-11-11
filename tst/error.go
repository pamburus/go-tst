package tst

import (
	"errors"
	"fmt"
	"reflect"
)

// IsTestIsDone reports whether the error is an error that indicates that the test is done.
func IsTestIsDone(err error) bool {
	return errors.Is(err, errTestIsDone)
}

// IsTestTimeout reports whether the error is an error that indicates that the test execution timed out.
func IsTestTimeout(err error) bool {
	return errors.Is(err, errTestTimeout)
}

// IsTestDeadlineExceeded reports whether the error is an error that indicates that the test execution deadline exceeded.
func IsTestDeadlineExceeded(err error) bool {
	return errors.Is(err, errTestDeadlineExceeded)
}

// ---

type errNumberOfValuesToTestDiffersError struct {
	actual   int
	expected int
}

func (e errNumberOfValuesToTestDiffersError) Error() string {
	return fmt.Sprintf("number of values to test is %d but expected to be %d", e.actual, e.expected)
}

// ---

type errUnexpectedValueTypeError struct {
	index    int
	actual   string
	expected string
}

func (e errUnexpectedValueTypeError) Error() string {
	return fmt.Sprintf("value to test #%d is expected to have type <%s> but it has type <%s>", e.index+1, e.expected, e.actual)
}

// ---

type errUnexpectedAssertionTypeError struct {
	index    int
	actual   string
	expected string
}

func (e errUnexpectedAssertionTypeError) Error() string {
	return fmt.Sprintf("value in assertion #%d is expected to have type <%s> but it has type <%s>", e.index+1, e.expected, e.actual)
}

// ---

var (
	errTestIsDone           = errors.New("test is done")
	errTestTimeout          = errors.New("test execution timed out")
	errTestDeadlineExceeded = errors.New("test execution deadline exceeded")
)

// ---

func typeOf[V any](v V) string {
	if any(v) == nil {
		return reflect.TypeOf(&v).Elem().String()
	}

	return reflect.TypeOf(v).String()
}
