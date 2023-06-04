package tst

import (
	"testing"
)

// New construct a new Test based on the t.
func New(t *testing.T) Test {
	return Test{t}
}

// ---

// Test transparently wraps testing.T object and adds additional methods to make expectations in an assertive way.
type Test struct {
	*testing.T
}

// Run runs sub-test using f.
func (t Test) Run(name string, f func(Test)) {
	t.T.Run(name, func(t *testing.T) {
		f(New(t))
	})
}

// Expect begins expectation building process against the given values.
func (t Test) Expect(values ...any) Expectation {
	return Expectation{&t, values}
}
