package tst

import (
	"fmt"
	"reflect"
	"strings"
)

func Panic() Assertion {
	return panicAssertion{}
}

func PanicWith(values ...any) Assertion {
	return panicAssertion{values}
}

type PanicAssertion interface {
	Assertion
	WithMessage(string) PanicAssertion
}

// ---

type panicAssertion struct {
	targets []any
}

func (a panicAssertion) check(actual []any) ([]bool, error) {
	result := make([]bool, len(actual))

	for i := range actual {
		v := reflect.ValueOf(actual[i])

		for v.Kind() != reflect.Func {
			return nil, errUnexpectedValueTypeError{i, typeOf(actual[i]), "func"}
		}
		if v.Type().NumIn() != 0 {
			return nil, errUnexpectedAssertionTypeError{i, typeOf(actual[i]), "func()"}
		}

		var assertion Assertion
		if pv, ok := at(a.targets, i); !ok || pv == nil {
			assertion = Not(BeNil())
		} else if pv, ok := pv.(Assertion); ok {
			assertion = pv
		} else {
			assertion = Equal(pv)
		}

		pv := catchReflect(v)
		if pv == nil {
			tpv, _ := at(a.targets, i)
			result[i] = tpv == nil
		} else {
			res, err := assertion.check([]any{pv})
			if err != nil {
				return nil, err
			}

			result[i] = res[0]
		}
	}

	return result, nil
}

func (a panicAssertion) description() string {
	if len(a.targets) == 0 {
		return "panic"
	}

	var b strings.Builder
	b.WriteString("panic with\n")

	for i, target := range a.targets {
		val := indented(1, value{target})
		if len(a.targets) > 1 {
			fmt.Fprintf(&b, "[#%d] %s\n", i+1, val)
		} else {
			fmt.Fprintf(&b, "%s\n", val)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func (a panicAssertion) complexity() int {
	return 1
}

func (a panicAssertion) at(int) Assertion {
	return a
}

func (a panicAssertion) String() string {
	return a.description()
}

// ---

func catchReflect(f reflect.Value) (r any) {
	return catch(func() {
		_ = f.Call(nil)
	})
}

func catch(f func()) (r any) {
	defer func() {
		r = recover()
	}()

	f()

	return nil
}

func at[T any](values []T, i int) (T, bool) {
	var zero T
	if i < 0 || i >= len(values) {
		return zero, false
	}

	return values[i], true
}
