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

// HaveNotOccurred returns an assertion that passes in case all the values to be tested are nil errors.
func HaveNotOccurred() Assertion {
	return nilError{}
}

// MatchError returns an assertion that passes in case all the values to be tested are errors that pass test `errors.Is(actual, expected)`.
func MatchError(expected error) Assertion {
	if expected == nil {
		return HaveNotOccurred()
	}

	return matchError{expected}
}

// ---

// Assertion is an abstract assertion that can be composite and can be used to build expectations.
type Assertion interface {
	check(actual []any) (bool, error)
	description() string
	complexity() int
}

// ---

type equal struct {
	expected []any
}

func (a equal) check(actual []any) (bool, error) {
	if len(a.expected) != len(actual) {
		return false, errNumberOfValuesToTestDiffers{len(actual), len(a.expected)}
	}

	return reflect.DeepEqual(actual, a.expected), nil
}

func (a equal) description() string {
	return "equal to\n" + indent(1, values(a.expected).description())
}

func (a equal) complexity() int {
	return 1
}

// ---

type boolean struct {
	expected bool
}

func (a boolean) check(actual []any) (bool, error) {
	for i := range actual {
		value, ok := actual[i].(bool)
		if !ok {
			return false, errUnexpectedValueType{i, typeOf(actual[i]), typeOf(a.expected)}
		}

		if value != a.expected {
			return false, nil
		}
	}

	return true, nil
}

func (a boolean) description() string {
	return "be\n" + indent(1, value{a.expected}.description())
}

func (a boolean) complexity() int {
	return 1
}

// ---

type zero struct{}

func (a zero) check(actual []any) (bool, error) {
	for i := range actual {
		if !reflect.ValueOf(actual[i]).IsZero() {
			return false, nil
		}
	}

	return true, nil
}

func (a zero) description() string {
	return "be zero"
}

func (a zero) complexity() int {
	return 1
}

// ---

type nilError struct{}

func (a nilError) check(actual []any) (bool, error) {
	for i := range actual {
		if actual[i] == nil {
			continue
		}

		val, ok := actual[i].(error)
		if !ok {
			return false, errUnexpectedValueType{i, typeOf(actual[i]), typeOf(val)}
		}

		if val != nil {
			return false, nil
		}
	}

	return true, nil
}

func (a nilError) description() string {
	return "be\n" + indentSnippet + "<nil>"
}

func (a nilError) complexity() int {
	return 1
}

// ---

type matchError struct {
	expected error
}

func (a matchError) check(actual []any) (bool, error) {
	for i := range actual {
		if actual[i] == nil {
			return false, nil
		}

		val, ok := actual[i].(error)
		if !ok {
			return false, errUnexpectedValueType{i, typeOf(actual[i]), typeOf(val)}
		}

		if !errors.Is(val, a.expected) {
			return false, nil
		}
	}

	return true, nil
}

func (a matchError) description() string {
	return "be non-nil and match error\n" + indent(1, value{a.expected}.description())
}

func (a matchError) complexity() int {
	return 1
}

// ---

type not struct {
	assertion Assertion
}

func (a not) check(actual []any) (bool, error) {
	ok, err := a.assertion.check(actual)
	if err != nil {
		return false, err
	}

	return !ok, nil
}

func (a not) description() string {
	if a.assertion.complexity() > 1 {
		return "not\n" + indent(1, a.assertion.description())
	}

	return "not " + a.assertion.description()
}

func (a not) complexity() int {
	return 1
}

// ---

type and struct {
	assertions []Assertion
}

func (a and) check(actual []any) (bool, error) {
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

// ---

type or struct {
	assertions []Assertion
}

func (a or) check(actual []any) (bool, error) {
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

func combineAssertionChecks(assertions []Assertion, actual []any, initial bool, operator func(bool, bool) bool) (bool, error) {
	if len(assertions) == 0 {
		return false, errors.New("expected 1 or more assertion in composite assertion")
	}

	result := initial

	for i := range assertions {
		ok, err := assertions[i].check(actual)
		if err != nil {
			return false, err
		}

		result = operator(result, ok)
	}

	return result, nil
}
