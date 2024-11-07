package tst

import (
	"testing"
)

// New constructs a new Test based on the t.
func New[T Base[T]](t T) Test {
	return test[T]{core{TB: t}, t.Run}
}

// Test transparently wraps compatible an object compatible with [testing.TB]
// and adds additional methods to make expectations in an assertive way.
type Test interface {
	testing.TB

	// Run runs sub-test using f.
	Run(name string, f func(Test)) bool

	// Expect begins expectation building process against the given values.
	Expect(values ...any) Expectation

	// WithLineTag returns a new Test that adds information about the line where it was called to the error messages.
	WithLineTag(tag LineTag) Test

	sealed()
}

// ---

type test[T Base[T]] struct {
	core
	run func(name string, f func(T)) bool
}

func (t test[T]) Run(name string, f func(Test)) bool {
	return t.run(name, func(t T) {
		f(New(t))
	})
}

func (t test[T]) Expect(values ...any) Expectation {
	return Expectation{&t.core, values}
}

func (t test[T]) WithLineTag(tag LineTag) Test {
	t.tags = append(t.tags[:len(t.tags):len(t.tags)], tag)

	return t
}

func (t test[T]) sealed() {}

// ---

type core struct {
	testing.TB
	tags []LineTag
}
