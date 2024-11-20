// Package mock provides a way to create and manage mock objects.
package mock

import "sync"

// Expect expects the given assertions to be true.
func Expect(t test, assertions ...Assertion) {
	t.Helper()

	expect(t, unlocked, assertions...)
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

	ctrl := mustGetController(b.t.Context())

	ctrl.AtomicCheck(b.t, b.f, assertions...)
}

// ---

// Assertion is a single assertion for a mock object or a group of assertions.
type Assertion interface {
	setup(test, lockStatus)
	lockers() map[sync.Locker]struct{}
}

// ---

type assertionFunc struct {
	f func(test, lockStatus)
	l func() map[sync.Locker]struct{}
}

func (f assertionFunc) setup(t test, l lockStatus) {
	f.f(t, l)
}

func (f assertionFunc) lockers() map[sync.Locker]struct{} {
	return f.l()
}

// ---

func expect(t test, lock lockStatus, assertions ...Assertion) {
	t.Helper()

	for _, assertion := range assertions {
		assertion.setup(t, lock)
	}
}

func catch(f func()) (r any) {
	defer func() {
		r = recover()
	}()

	f()

	return nil
}

// ---

type lockStatus int

const (
	unlocked lockStatus = iota
	locked
)
