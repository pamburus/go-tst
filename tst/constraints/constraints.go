// Package constraints provides constraints for type parameters in [tst](github.com/pamburus/go-tst/tst) package.
package constraints

import "testing"

// HT is constraint for a any hierarchical test compatible with [testing.TB].
type HT[TT testing.TB] interface {
	testing.TB

	// Run runs a subtest with the given name and function.
	Run(name string, f func(TT)) bool
}

// T is constraint for a any test compatible with [testing.T].
type T[TT HT[TT]] interface {
	HT[TT]
	Helper
	Parallel()
}

// B is constraint for a any test with context compatible with [testing.B].
type B[TT HT[TT]] interface {
	HT[TT]
	Helper
}

// Helper is constraint for a test helper types.
type Helper interface {
	Helper()
}
