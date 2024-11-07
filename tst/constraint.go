package tst

import "testing"

// ---

// Base is constraint for a hierarchical test base.
type Base[T testing.TB] interface {
	testing.TB
	Run(name string, f func(T)) bool
}
