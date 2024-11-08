package tst

import "testing"

// ---

// HT is constraint for a any hierarchical test compatible with [testing.TB].
type HT[T testing.TB] interface {
	testing.TB

	// Run runs a subtest with the given name and function.
	Run(name string, f func(T)) bool
}
