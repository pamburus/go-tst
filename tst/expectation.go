package tst

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ---

// Expectation is an expectation builder that have associated values to be tested against assertions.
type Expectation struct {
	t      *core
	actual []any
	tag    LineTag
}

// To tests that the associated values conform all of the given assertions.
func (e Expectation) To(assertions ...Assertion) {
	e.t.Helper()

	assertion := func(i int) Assertion {
		return assertions[i]
	}

	if len(e.actual) != len(assertions) {
		if len(assertions) != 1 || len(e.actual) <= 1 {
			e.log(msg("number of values to test", value{len(e.actual)}, expDesc("be", len(assertions))))
			e.fail()
		}

		assertion = func(int) Assertion {
			return assertions[0]
		}
	}

	fail := func(i int, assertion Assertion) {
		e.t.Helper()

		what := ""
		if len(e.actual) != 1 {
			what = fmt.Sprintf("value #%d", i+1)
		}

		e.log(msg(what, value{e.actual[i]}, assertion))
		e.t.Fail()
	}

	if len(assertions) == 1 && len(e.actual) > 1 {
		ok := e.check(assertions[0], e.actual)
		for i := range e.actual {
			if !ok[i] {
				fail(i, assertions[0].at(i))
			}
		}
	} else {
		for i := range e.actual {
			if !e.check(assertion(i), []any{e.actual[i]})[0] {
				fail(i, assertion(i))
			}
		}
	}

	if e.t.Failed() {
		e.fail()
	}
}

// ToNot tests that the associated values do not conform all of the given assertions.
func (e Expectation) ToNot(assertions ...Assertion) {
	e.t.Helper()

	for i := range assertions {
		assertions[i] = Not(assertions[i])
	}

	e.To(assertions...)
}

// ToEqual tests that the associated values are equal to the specified expected values.
//
// The number of associated values must be exactly the same as the number of specified expected values.
func (e Expectation) ToEqual(expected ...any) {
	e.t.Helper()

	e.To(Equal(expected...))
}

// ToNotEqual tests that the associated values are not equal to the specified expected values.
//
// The number of associated values must be exactly the same as the number of specified expected values.
func (e Expectation) ToNotEqual(expected ...any) {
	e.t.Helper()

	e.To(NotEqual(expected...))
}

// ToBeTrue tests that all of the associated values are boolean values and are true.
func (e Expectation) ToBeTrue() {
	e.t.Helper()

	e.To(BeTrue())
}

// ToBeFalse tests that all of the associated values are boolean values and are false.
func (e Expectation) ToBeFalse() {
	e.t.Helper()

	e.To(BeFalse())
}

// ToSucceed tests that the last of the associated values is a nil error
// and returns a SuccessExpectation that allows to add assertions to check other values.
//
// The associated values are assumed to be return values from a function call returning a error
// so the last value should be an error and all other values are treated as a result.
func (e Expectation) ToSucceed() SuccessExpectation {
	e.t.Helper()

	if len(e.actual) == 0 {
		e.log(msg("number of values to test", value{len(e.actual)}, expDescText("be", "non-zero")))
		e.fail()
	}

	last := e.actual[len(e.actual)-1]
	if last == nil {
		return SuccessExpectation{e}
	}

	actual, ok := last.(error)
	if !ok {
		e.log(msg("last value to test", value{last}, expDescText("be", "an error")))
		e.fail()
	}

	if actual == nil {
		return SuccessExpectation{e}
	}

	e.log(msg("error", value{actual}, expDesc("be", nil)))
	e.fail()

	return SuccessExpectation{}
}

// ToFail builds expectation for a non-nil error value that is expected be the last in the list of values
// assuming these values are return values from a function call.
//
// All other values are ignored in this expectation.
func (e Expectation) ToFail() {
	if len(e.actual) == 0 {
		e.log(msg("number of values to test", value{len(e.actual)}, expDescText("be", "non-zero")))
		e.fail()
	}

	last := e.actual[len(e.actual)-1]
	if last != nil {
		_, ok := last.(error)
		if !ok {
			e.log(msg("last value to test", value{last}, expDescText("be", "an error")))
			e.fail()
		}

		return
	}

	e.t.Helper()
	e.log(msg("error", value{last}, expDescText("be", "non-nil error")))
	e.fail()
}

// ToFailWith builds expectation for an error value that is expected be the last in the list of values
// assuming these values are return values from a function call.
//
// All other values are ignored in this expectation.
func (e Expectation) ToFailWith(err error) {
	e.t.Helper()

	if len(e.actual) == 0 {
		e.log(msg("number of values to test", value{len(e.actual)}, expDescText("be", "non-zero")))
		e.fail()
	}

	last := e.actual[len(e.actual)-1]
	if last == nil {
		return
	}

	actual, ok := last.(error)
	if !ok {
		e.log(msg("last value to test", value{last}, expDescText("be", "an error")))
		e.fail()
	}

	if actual != nil && errors.Is(actual, err) {
		return
	}

	e.t.Helper()
	e.log(msg("error", value{actual}, expDesc("be like", err)))
	e.fail()
}

func (e Expectation) check(assertion Assertion, actual []any) []bool {
	e.t.Helper()

	ok, err := assertion.check(actual)
	if err != nil {
		var ee errNumberOfValuesToTestDiffersError
		if errors.As(err, &ee) {
			e.log(msg("number of values to test", value{ee.actual}, expDesc("be", ee.expected)))
		} else {
			e.log(err)
		}

		e.fail()
	}

	return ok
}

func (e Expectation) log(args ...any) {
	e.t.Helper()
	e.t.Log(args...)
}

func (e Expectation) fail() {
	e.t.addLineTags(e.tag)
	e.t.FailNow()
}

// ---

// SuccessExpectation is an expectation build that can be used
// to additionally test remaining result values of a function call after checking its error.
type SuccessExpectation struct {
	e Expectation
}

// AndResult returns an expectation builder for the associated values except for the last one.
func (e SuccessExpectation) AndResult() Expectation {
	return Expectation{e.e.t, e.e.actual[:len(e.e.actual)-1], e.e.tag}
}

// ---

func expDesc(what string, expected any) expTextDesc {
	return expDescText(what, value{expected}.description())
}

func expDescText(what, text string) expTextDesc {
	return expTextDesc{what, text}
}

// ---

type expTextDesc struct {
	what string
	text string
}

func (e expTextDesc) description() string {
	return fmt.Sprintf("%s\n%s", e.what, indent(1, e.text))
}

// ---

func msg(what string, actual, expected describable) string {
	return fmt.Sprintf("\nExpected %s\n%s\nto %s", what, indent(1, actual.description()), expected.description())
}

// ---

//nolint:unparam // `indent` - `n` always receives `1`
func indent(n int, text string) string {
	var sb strings.Builder

	for _, line := range strings.Split(text, "\n") {
		for range n {
			sb.WriteString(indentSnippet)
		}

		sb.WriteString(line)
		sb.WriteRune('\n')
	}

	return strings.TrimRight(sb.String(), "\n")
}

// ---

type value struct {
	v any
}

func (v value) description() string {
	comment := ""
	length := -1

	switch vv := v.v.(type) {
	case nil:
		return "<nil>"
	case []byte:
		comment = fmt.Sprintf(" | %q", vv)
		length = len(vv)
	case fmt.Stringer:
		s := vv.String()
		comment = fmt.Sprintf(" | [%d] %s", len(s), s)
	default:
		rv := reflect.ValueOf(vv)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map, reflect.String, reflect.Chan:
			length = rv.Len()
		}
	}

	ls := ""
	if length != -1 {
		ls = fmt.Sprintf("[%d] ", length)
	}

	return fmt.Sprintf("<%T>: %s%#v%s", v.v, ls, v.v, comment)
}

// ---

type values []any

func (v values) description() string {
	if len(v) == 1 {
		return value{v[0]}.description()
	}

	var sb strings.Builder
	for i := range v {
		fmt.Fprintf(&sb, "[#%d] %s\n", i+1, value{v[i]}.description())
	}

	return strings.TrimRight(sb.String(), "\n")
}

// ---

type describable interface {
	description() string
}

// ---

const indentSnippet = "    "
