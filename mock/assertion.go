package mock

func Expect(t T, assertions ...Assertion) {
	for _, assertion := range assertions {
		assertion.setup(t)
	}
}

type Assertion interface {
	setup(T)
}

// ---

type assertionFunc func(T)

func (f assertionFunc) setup(t T) {
	f(t)
}
