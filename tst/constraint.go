package tst

import "testing"

// ---

// HT is constraint for a any hierarchical test compatible with [testing.TB].
type HT[T testing.TB] interface {
	testing.TB
	Run(name string, f func(T)) bool
}
