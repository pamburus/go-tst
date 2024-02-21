// Package tst provides a lightweight framework for writing unit tests in an assertive style.
package tst

import (
	"errors"
	"reflect"
	"strings"
)

// Not returns an assertion that passes in case the specified assertion do not pass and vise versa.
func Not(assertion Assertion) Assertion {
	return not{assertion}
}

// And returns an assertion that passes in case all of the specified assertions pass.
func And(assertions ...Assertion) Assertion {
	return and{assertions}
}

// Or returns an assertion that passes in case any of the specified assertions pass.
func Or(assertions ...Assertion) Assertion {
	return or{assertions}
}

// Equal returns an assertion that passes in case values to be tested using it equal to the specified values.
// Number of values to test with this assertion must match the number of the specified values.
func Equal(values ...any) Assertion {
	return equal{values}
}

// NotEqual returns an assertion that passes in case values to be tested using it not equal to the specified values.
// Number of values to test with this assertion must match the number of the specified values.
func NotEqual(values ...any) Assertion {
	return Not(Equal(values...))
}

// BeTrue returns an assertion that passes in case all the values to be tested are boolean and equal to true.
func BeTrue() Assertion {
	return boolean{true}
}

// BeFalse returns an assertion that passes in case all the values to be tested are boolean and equal to false.
func BeFalse() Assertion {
	return boolean{false}
}

// BeZero returns an assertion that passes in case all the values to be tested are zero initialized.
func BeZero() Assertion {
	return zero{}
}

// BeNil returns an assertion that passes in case all the values to be tested are nil.
func BeNil() Assertion {
	return nilValue{}
}

// HaveOccurred returns an assertion that passes in case all the values to be tested are non-nil errors.
func HaveOccurred() Assertion {
	return Not(nilError{})
}

// MatchError returns an assertion that passes in case all the values to be tested are errors that pass test `errors.Is(actual, expected)`.
func MatchError(expected error) Assertion {
	if expected == nil {
		return nilError{}
	}

	return matchError{expected}
}

// HaveLen returns an assertion that passes in case all the values to be tested have length equal to the specified value.
func HaveLen(n ...int) Assertion {
	return haveLen{n}
}

// ---

// Assertion is an abstract assertion that can be composite and can be used to build expectations.
type Assertion interface {
	check(actual []any) ([]bool, error)
	description() string
	complexity() int
	at(int) Assertion
}

// ---

type equal struct {
	expected []any
}

func (a equal) check(actual []any) ([]bool, error) {
	if len(a.expected) != len(actual) && len(a.expected) != 1 {
		return nil, errNumberOfValuesToTestDiffers{len(actual), len(a.expected)}
	}

	expected := func(i int) any {
		if len(a.expected) == 1 {
			return a.expected[0]
		}

		return a.expected[i]
	}

	result := make([]bool, len(actual))
	for i := range actual {
		result[i] = reflect.DeepEqual(actual[i], expected(i))
	}

	return result, nil
}

func (a equal) description() string {
	return "equal to\n" + indent(1, values(a.expected).description())
}

func (a equal) complexity() int {
	return 1
}

func (a equal) at(i int) Assertion {
	if len(a.expected) == 1 {
		return a
	}

	return equal{[]any{a.expected[i]}}
}

// ---

type boolean struct {
	expected bool
}

func (a boolean) check(actual []any) ([]bool, error) {
	result := make([]bool, len(actual))

	for i := range actual {
		value, ok := actual[i].(bool)
		if !ok {
			return nil, errUnexpectedValueType{i, typeOf(actual[i]), typeOf(a.expected)}
		}

		result[i] = value == a.expected
	}

	return result, nil
}

func (a boolean) description() string {
	return "be\n" + indent(1, value{a.expected}.description())
}

func (a boolean) complexity() int {
	return 1
}

func (a boolean) at(int) Assertion {
	return a
}

// ---

type zero struct{}

func (a zero) check(actual []any) ([]bool, error) {
	result := make([]bool, len(actual))
	for i := range actual {
		result[i] = reflect.ValueOf(actual[i]).IsZero()
	}

	return result, nil
}

func (a zero) description() string {
	return "be zero"
}

func (a zero) complexity() int {
	return 1
}

func (a zero) at(int) Assertion {
	return a
}

// ---

type nilValue struct{}

func (a nilValue) check(actual []any) ([]bool, error) {
	result := make([]bool, len(actual))
	for i := range actual {
		if actual[i] == nil {
			result[i] = true

			continue
		}

		v := reflect.ValueOf(actual[i])
		switch v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			result[i] = v.IsNil()
		default:
			result[i] = false
		}
	}

	return result, nil
}

func (a nilValue) description() string {
	return "be nil"
}

func (a nilValue) complexity() int {
	return 1
}

func (a nilValue) at(int) Assertion {
	return a
}

// ---

type nilError struct{}

func (a nilError) check(actual []any) ([]bool, error) {
	result := make([]bool, len(actual))

	for i := range actual {
		if actual[i] == nil {
			result[i] = true

			continue
		}

		val, ok := actual[i].(error)
		if !ok {
			return nil, errUnexpectedValueType{i, typeOf(actual[i]), typeOf(val)}
		}

		result[i] = val == nil
	}

	return result, nil
}

func (a nilError) description() string {
	return "be\n" + indentSnippet + "<nil>"
}

func (a nilError) complexity() int {
	return 1
}

func (a nilError) at(int) Assertion {
	return a
}

// ---

type matchError struct {
	expected error
}

func (a matchError) check(actual []any) ([]bool, error) {
	result := make([]bool, len(actual))

	for i := range actual {
		if actual[i] == nil {
			result[i] = false

			continue
		}

		val, ok := actual[i].(error)
		if !ok {
			return nil, errUnexpectedValueType{i, typeOf(actual[i]), typeOf(val)}
		}

		result[i] = errors.Is(val, a.expected)
	}

	return result, nil
}

func (a matchError) description() string {
	return "be non-nil and match error\n" + indent(1, value{a.expected}.description())
}

func (a matchError) complexity() int {
	return 1
}

func (a matchError) at(int) Assertion {
	return a
}

// ---

type haveLen struct {
	expected []int
}

func (a haveLen) check(actual []any) ([]bool, error) {
	if len(a.expected) != len(actual) && len(a.expected) != 1 {
		return nil, errNumberOfValuesToTestDiffers{len(actual), len(a.expected)}
	}

	expected := func(i int) any {
		if len(a.expected) == 1 {
			return a.expected[0]
		}

		return a.expected[i]
	}

	result := make([]bool, len(actual))

	for i := range actual {
		v := reflect.ValueOf(actual[i])
		switch v.Kind() {
		case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array, reflect.String:
			if v.Len() == expected(i) {
				result[i] = true

				continue
			}
		case reflect.Ptr:
			if v.Elem().Kind() == reflect.Array {
				if v.Len() == expected(i) {
					result[i] = true

					continue
				}
			}
		}

		result[i] = false
	}

	return result, nil
}

func (a haveLen) description() string {
	return "have length\n" + indent(1, values(anySlice(a.expected)).description())
}

func (a haveLen) complexity() int {
	return 1
}

func (a haveLen) at(i int) Assertion {
	if len(a.expected) == 1 {
		return a
	}

	return haveLen{[]int{a.expected[i]}}
}

// ---

type not struct {
	assertion Assertion
}

func (a not) check(actual []any) ([]bool, error) {
	result, err := a.assertion.check(actual)
	if err != nil {
		return nil, err
	}

	for i := range result {
		result[i] = !result[i]
	}

	return result, nil
}

func (a not) description() string {
	if inner, ok := a.assertion.(not); ok {
		return inner.assertion.description()
	}

	if a.assertion.complexity() > 1 {
		return "not\n" + indent(1, a.assertion.description())
	}

	return "not " + a.assertion.description()
}

func (a not) complexity() int {
	return 1
}

func (a not) at(i int) Assertion {
	return Not(a.assertion.at(i))
}

// ---

type and struct {
	assertions []Assertion
}

func (a and) check(actual []any) ([]bool, error) {
	return combineAssertionChecks(a.assertions, actual, true, func(x, y bool) bool {
		return x && y
	})
}

func (a and) description() string {
	return combineAssertionDescriptions("and", a.assertions)
}

func (a and) complexity() int {
	return len(a.assertions)
}

func (a and) at(int) Assertion {
	assertions := make([]Assertion, len(a.assertions))
	for i := range a.assertions {
		assertions[i] = a.assertions[i].at(i)
	}

	return and{assertions}
}

// ---

type or struct {
	assertions []Assertion
}

func (a or) check(actual []any) ([]bool, error) {
	return combineAssertionChecks(a.assertions, actual, false, func(x, y bool) bool {
		return x || y
	})
}

func (a or) description() string {
	return combineAssertionDescriptions("or", a.assertions)
}

func (a or) complexity() int {
	return len(a.assertions)
}

func (a or) at(int) Assertion {
	assertions := make([]Assertion, len(a.assertions))
	for i := range a.assertions {
		assertions[i] = a.assertions[i].at(i)
	}

	return or{assertions}
}

// ---

func combineAssertionDescriptions(operator string, assertions []Assertion) string {
	if len(assertions) == 1 {
		return assertions[0].description()
	}

	var sb strings.Builder
	for i, assertion := range assertions {
		if i != 0 {
			sb.WriteRune('\n')
			sb.WriteString(operator)
			sb.WriteRune(' ')
		}
		desc := assertion.description()
		if assertion.complexity() > 1 {
			sb.WriteString("\n")
			sb.WriteString(indent(1, desc))
		} else {
			sb.WriteString(desc)
		}
	}

	return sb.String()
}

func combineAssertionChecks(assertions []Assertion, actual []any, initial bool, operator func(bool, bool) bool) ([]bool, error) {
	if len(assertions) == 0 {
		return nil, errors.New("expected 1 or more assertion in composite assertion")
	}

	result := make([]bool, len(actual))
	for i := range result {
		result[i] = initial

		for j := range assertions {
			ok, err := assertions[j].check(actual)
			if err != nil {
				return nil, err
			}

			result[i] = operator(result[i], ok[i])
		}
	}

	return result, nil
}

func anySlice[T any](values []T) []any {
	result := make([]any, len(values))

	for i, v := range values {
		result[i] = v
	}

	return result
}
