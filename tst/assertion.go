// Package tst provides a lightweight framework for writing unit tests in an assertive style.
package tst

import (
	"errors"
	"fmt"
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

// EqualUsing returns an assertion that passes in case values to be tested are equal to the specified values
// using the equality test function f.
func EqualUsing(f any, values ...any) Assertion {
	return equalUsing{f, values}
}

// NotEqual returns an assertion that passes in case values to be tested using it not equal to the specified values.
// Number of values to test with this assertion must match the number of the specified values.
func NotEqual(values ...any) Assertion {
	return Not(Equal(values...))
}

// BeLessThan returns an assertion that passes in case values to be tested using it are less than corresponding specified values.
// Number of values to test with this assertion must match the number of the specified values.
func BeLessThan(values ...any) Assertion {
	return comparison{values, lt, "less than"}
}

// LessThan returns an assertion that passes in case values to be tested using it are less than corresponding specified values.
func LessThan(values ...any) Assertion {
	return BeLessThan(values...)
}

// BeLessOrEqualThan returns an assertion that passes in case values to be tested using it are less or equal than corresponding specified values.
func BeLessOrEqualThan(values ...any) Assertion {
	return comparison{values, le, "less or equal than"}
}

// LessOrEqualThan returns an assertion that passes in case values to be tested using it are less or equal than corresponding specified values.
func LessOrEqualThan(values ...any) Assertion {
	return BeLessOrEqualThan(values...)
}

// BeGreaterThan returns an assertion that passes in case values to be tested using it are greater than corresponding specified values.
func BeGreaterThan(values ...any) Assertion {
	return comparison{values, gt, "greater than"}
}

// GreaterThan returns an assertion that passes in case values to be tested using it are greater than corresponding specified values.
func GreaterThan(values ...any) Assertion {
	return BeGreaterThan(values...)
}

// BeGreaterOrEqualThan returns an assertion that passes in case values to be tested using it are greater or equal than corresponding specified values.
func BeGreaterOrEqualThan(values ...any) Assertion {
	return comparison{values, ge, "greater or equal than"}
}

// GreaterOrEqualThan returns an assertion that passes in case values to be tested using it are greater or equal than corresponding specified values.
func GreaterOrEqualThan(values ...any) Assertion {
	return BeGreaterOrEqualThan(values...)
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
func HaveLen(n ...any) Assertion {
	return haveLen{n}
}

// HaveField returns an assertion that passes in case all the struct values to be tested have a field with the specified name and the specified assertion passes for the field value.
func HaveField(name string, assertion Assertion) Assertion {
	return haveField{name, assertion}
}

// Field returns an assertion that passes in case all the struct values to be tested have a field with the specified name and the specified assertion passes for the field value.
// It is an alias for HaveField and is provided for better readability when used in [Struct] or [Contain] assertions.
func Field(name string, assertion Assertion) Assertion {
	return haveField{name, assertion}
}

// Contain returns an assertion that passes in case all the array or slice values to be tested contain at least one element matching each of the given assertions.
func Contain(assertion Assertion) Assertion {
	return contain{assertion}
}

// Struct returns an assertion that passes in case all the values to be tested are structs each of them containing at least one field matching any of the expected field assertions.
func Struct(fields ...Assertion) Assertion {
	return beStruct{fields}
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

type equalUsing struct {
	f        any
	expected []any
}

func (a equalUsing) check(actual []any) ([]bool, error) {
	if len(a.expected) != len(actual) && len(a.expected) != 1 {
		return nil, errNumberOfValuesToTestDiffers{len(actual), len(a.expected)}
	}

	expected := func(i int) any {
		if len(a.expected) == 1 {
			return a.expected[0]
		}

		return a.expected[i]
	}

	rf := reflect.ValueOf(a.f)
	if rf.Kind() != reflect.Func {
		return nil, errUnexpectedAssertionType{0, typeOf(a.f), "function"}
	}

	if rf.Type().NumIn() != 2 || rf.Type().NumOut() != 1 {
		return nil, errUnexpectedAssertionType{0, typeOf(a.f), "function with two arguments and one return value"}
	}

	if rf.Type().In(0) != reflect.TypeOf(actual[0]) || rf.Type().In(1) != reflect.TypeOf(expected(0)) {
		return nil, errUnexpectedAssertionType{0, typeOf(a.f), fmt.Sprintf("func(%s, %s) bool", typeOf(actual[0]), typeOf(expected(0)))}
	}

	if rf.Type().Out(0) != reflect.TypeOf(true) {
		return nil, errUnexpectedAssertionType{0, typeOf(a.f), "function with return value of type bool"}
	}

	result := make([]bool, len(actual))
	for i := range actual {
		out := rf.Call([]reflect.Value{reflect.ValueOf(actual[i]), reflect.ValueOf(expected(i))})
		result[i] = out[0].Bool()
	}

	return result, nil
}

func (a equalUsing) description() string {
	return "equal to\n" + indent(1, values(a.expected).description())
}

func (a equalUsing) complexity() int {
	return 1
}

func (a equalUsing) at(i int) Assertion {
	if len(a.expected) == 1 {
		return a
	}

	return equal{[]any{a.expected[i]}}
}

// ---

type comparison struct {
	expected            []any
	expectedResult      func(int) bool
	expectedDescription string
}

func (a comparison) check(actual []any) ([]bool, error) {
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
		actual := reflect.ValueOf(actual[i])
		expected := reflect.ValueOf(expected(i))
		r, err := compare(i, actual, expected)
		if err != nil {
			return nil, err
		}
		result[i] = a.expectedResult(r)
	}

	return result, nil
}

func (a comparison) description() string {
	return "be " + a.expectedDescription + "\n" + indent(1, values(a.expected).description())
}

func (a comparison) complexity() int {
	return 1
}

func (a comparison) at(i int) Assertion {
	if len(a.expected) == 1 {
		return a
	}

	return comparison{[]any{a.expected[i]}, a.expectedResult, a.expectedDescription}
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
	expected []any
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

	test := func(i int, actual int, expected any) (bool, error) {
		switch expected := expected.(type) {
		case Assertion:
			ok, err := expected.check([]any{actual})
			if err != nil {
				return false, err
			}

			return ok[0], nil
		case int:
			return actual == expected, nil
		default:
			return false, errUnexpectedValueType{i, typeOf(actual), typeOf(expected)}
		}
	}

	result := make([]bool, len(actual))

	for i := range actual {
		var err error
		v := reflect.ValueOf(actual[i])

		switch v.Kind() {
		case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array, reflect.String:
			result[i], err = test(i, v.Len(), expected(i))
		case reflect.Ptr:
			if v.Elem().Kind() == reflect.Array {
				result[i], err = test(i, v.Len(), expected(i))
			} else {
				return nil, errUnexpectedValueType{i, typeOf(actual[i]), typeOf(expected(i))}
			}
		default:
			return nil, errUnexpectedValueType{i, typeOf(actual[i]), typeOf(expected(i))}
		}
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (a haveLen) description() string {
	if len(a.expected) == 1 {
		if assertion, ok := a.expected[0].(Assertion); ok {
			return "have length that is expected to " + assertion.description()
		}
	}

	return "have length\n" + indent(1, values(anySlice(a.expected)).description())
}

func (a haveLen) complexity() int {
	return 1
}

func (a haveLen) at(i int) Assertion {
	if len(a.expected) == 1 {
		return a
	}

	return haveLen{[]any{a.expected[i]}}
}

// ---

type haveField struct {
	name      string
	assertion Assertion
}

func (a haveField) check(actual []any) ([]bool, error) {
	result := make([]bool, len(actual))

	for i := range actual {
		v := reflect.ValueOf(actual[i])

		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			field := v.FieldByName(a.name)
			if field.IsValid() {
				r, err := a.assertion.check([]any{field.Interface()})
				if err != nil {
					return nil, err
				}
				result[i] = r[0]
			}
		default:
			return nil, errUnexpectedValueType{i, typeOf(actual[i]), "struct"}
		}
	}

	return result, nil
}

func (a haveField) description() string {
	return fmt.Sprintf("have field %q that is expected to %s", a.name, a.assertion.description())
}

func (a haveField) complexity() int {
	return 1
}

func (a haveField) at(int) Assertion {
	return a
}

// ---

type contain struct {
	assertion Assertion
}

func (a contain) check(actual []any) ([]bool, error) {
	result := make([]bool, len(actual))

	for i := range actual {
		v := reflect.ValueOf(actual[i])

		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			for j := 0; j < v.Len(); j++ {
				ok, err := a.assertion.check([]any{v.Index(j).Interface()})
				if err != nil {
					return nil, err
				}
				if ok[0] {
					result[i] = true

					break
				}
			}
		default:
			return nil, errUnexpectedValueType{i, typeOf(actual[i]), "array or slice"}
		}
	}

	return result, nil
}

func (a contain) description() string {
	return fmt.Sprintf("contain an element that is expected to\n%s", indent(1, a.assertion.description()))
}

func (a contain) complexity() int {
	return 2
}

func (a contain) at(int) Assertion {
	return a
}

// ---

type beStruct struct {
	assertions []Assertion
}

func (a beStruct) check(actual []any) ([]bool, error) {
	result := make([]bool, len(actual))

	for i := range actual {
		v := reflect.ValueOf(actual[i])

		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return nil, errUnexpectedValueType{i, typeOf(actual[i]), "struct"}
		}

		result[i] = true
		for _, assertion := range a.assertions {
			ok, err := assertion.check([]any{actual[i]})
			if err != nil {
				return nil, err
			}
			if !ok[0] {
				result[i] = false

				break
			}
		}
	}

	return result, nil
}

func (a beStruct) description() string {
	sb := strings.Builder{}
	sb.WriteString("be a struct that is expected to\n")
	for i, assertion := range a.assertions {
		fmt.Fprintf(&sb, "%d. ", i+1)
		sb.WriteString(indent(1, assertion.description()))
		sb.WriteRune('\n')
	}

	return sb.String()
}

func (a beStruct) complexity() int {
	return 2
}

func (a beStruct) at(int) Assertion {
	return a
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

func compare(i int, actual, expected reflect.Value) (int, error) {
	fail := func() (int, error) {
		return 0, errUnexpectedValueType{i, typeOf(actual.Interface()), typeOf(expected.Interface())}
	}
	succeed := func(result int) (int, error) {
		return result, nil
	}

	switch actual.Kind() {
	case reflect.Bool:
		if actual.Kind() != expected.Kind() {
			return fail()
		}

		if actual.Bool() == expected.Bool() {
			return succeed(0)
		}

		if actual.Bool() {
			return succeed(1)
		}

		return succeed(-1)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch expected.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			switch {
			case actual.Int() > expected.Int():
				return succeed(1)
			case actual.Int() < expected.Int():
				return succeed(-1)
			default:
				return succeed(0)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if actual.Int() < 0 {
				return succeed(-1)
			}
			switch {
			case uint64(actual.Int()) > expected.Uint():
				return succeed(1)
			case uint64(actual.Int()) < expected.Uint():
				return succeed(-1)
			default:
				return succeed(0)
			}
		default:
			return fail()
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch expected.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			switch {
			case actual.Uint() > expected.Uint():
				return succeed(1)
			case actual.Uint() < expected.Uint():
				return succeed(-1)
			default:
				return succeed(0)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if expected.Int() < 0 {
				return succeed(1)
			}
			switch {
			case actual.Uint() > uint64(expected.Int()):
				return succeed(1)
			case actual.Uint() < uint64(expected.Int()):
				return succeed(-1)
			default:
				return succeed(0)
			}
		default:
			return fail()
		}

	case reflect.Float32, reflect.Float64:
		switch expected.Kind() {
		case reflect.Float32, reflect.Float64:
			switch {
			case actual.Float() > expected.Float():
				return succeed(1)
			case actual.Float() < expected.Float():
				return succeed(-1)
			default:
				return succeed(0)
			}
		default:
			return fail()
		}

	case reflect.String:
		if actual.Kind() != expected.Kind() {
			return fail()
		}

		return succeed(strings.Compare(actual.String(), expected.String()))

	default:
		return fail()
	}
}

func lt(r int) bool {
	return r < 0
}

func le(r int) bool {
	return r <= 0
}

func gt(r int) bool {
	return r > 0
}

func ge(r int) bool {
	return r >= 0
}
