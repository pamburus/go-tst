// Package mock provides a way to create and manage mock objects.
package mock

// Expect expects the given assertions to be true.
func Expect(t test, assertions ...Assertion) {
	for _, assertion := range assertions {
		assertion.setup(t)
	}
}

// Assertion is a single assertion for a mock object or a group of assertions.
type Assertion interface {
	setup(test)
}

// ---

type assertionFunc func(test)

func (f assertionFunc) setup(t test) {
	f(t)
}
