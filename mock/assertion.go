// Package mock provides a way to create and manage mock objects.
package mock

import "errors"

// Expect expects the given assertions to be true.
func Expect(t test, assertions ...Assertion) {
	t.Helper()

	for _, assertion := range assertions {
		assertion.setup(t)
	}
}

// ---

func On(t test) OnBuilder {
	return OnBuilder{t}
}

// ---

type OnBuilder struct {
	t test
}

func (b OnBuilder) During(f func()) ExpectationBuilder {
	return ExpectationBuilder{b.t, f}
}

// ---

type ExpectationBuilder struct {
	t test
	f func()
}

// Expect expects the given assertions to be true.
func (b ExpectationBuilder) Expect(assertions ...Assertion) {
	b.t.Helper()

	Expect(b.t, assertions...)

	pv := catch(b.f)
	if pv != nil {
		var unexpectedCall errUnexpectedCallError
		if err, ok := pv.(error); ok && errors.As(err, &unexpectedCall) {
			b.t.Error(unexpectedCall)
		} else {
			panic(pv)
		}
	}

	getController(b.t.Context()).Checkpoint()
}

// ---

// Assertion is a single assertion for a mock object or a group of assertions.
type Assertion interface {
	setup(test)
}

// ---

type assertionFunc func(test)

func (f assertionFunc) setup(t test) {
	f(t)
}

// ---

func catch(f func()) (r any) {
	defer func() {
		r = recover()
	}()

	f()

	return nil
}
