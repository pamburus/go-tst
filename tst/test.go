package tst

import (
	"testing"
)

// New constructs a new Test based on the t.
func New[T HT[T]](t T) Test {
	return &test[T]{core{TB: t}}
}

// Test transparently wraps compatible an object compatible with [testing.TB]
// and adds additional methods to make expectations in an assertive way.
type Test interface {
	testing.TB

	// Run runs sub-test using f.
	Run(name string, f func(Test)) bool

	// Expect begins expectation building process against the given values.
	Expect(values ...any) Expectation

	// AddLineTags adds information about the lines of interest to be displayed in test failure message.
	// Do not add lines that are not relevant to the test failure.
	AddLineTags(tags ...LineTag)

	sealed()
	get() *core
}

// ---

type test[T HT[T]] struct {
	core
}

func (t *test[T]) Run(name string, f func(Test)) bool {
	t.Helper()

	//nolint:forcetypeassert // it is guaranteed that TB is T
	return t.TB.(T).Run(name, func(tt T) {
		tt.Helper()
		f(t.fork(tt))
	})
}

func (t *test[T]) Expect(values ...any) Expectation {
	return Expectation{&t.core, values, CallerLine(1)}
}

func (t *test[T]) AddLineTags(tags ...LineTag) {
	t.addLineTags(tags...)
}

func (t *test[T]) fork(tt T) *test[T] {
	tt.Helper()

	fork := &test[T]{core{tt, t.tags}}
	setup(fork)

	return fork
}

func (t *test[T]) sealed() {}

func (t *test[T]) get() *core {
	return &t.core
}

// ---

type core struct {
	testing.TB
	tags []LineTag
}

func (c *core) addLineTags(tags ...LineTag) {
	c.tags = append(c.tags, tags...)
}

// ---

func setup(t Test) {
	t.Helper()
	t.Cleanup(func() {
		if t.Failed() {
			for _, tag := range t.get().tags {
				t.Helper()
				t.Log("See", tag)
			}
		}
	})
}
