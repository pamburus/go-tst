package tst

import (
	"testing"
)

// New constructs a new Test based on the t.
func New(t *testing.T) Test {
	return Test{T: t}
}

// ---

// Test transparently wraps testing.T object and adds additional methods to make expectations in an assertive way.
type Test struct {
	*testing.T
	tags []LineTag
}

// Run runs sub-test using f.
func (t Test) Run(name string, f func(Test)) {
	t.T.Run(name, func(t *testing.T) {
		f(New(t))
	})
}

// Expect begins expectation building process against the given values.
func (t Test) Expect(values ...any) Expectation {
	return Expectation{&t, values, t.tags}
}

// WithLineTag returns a new Test that adds information about the line where it was called to the error messages.
func (t Test) WithLineTag(tag LineTag) Test {
	t.tags = append(t.tags[:len(t.tags):len(t.tags)], tag)

	return t
}
