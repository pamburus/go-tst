package mock

import (
	"sync"
	"sync/atomic"
	"testing"
)

// NewController creates a new mock controller.
func NewController(t testing.TB) *Controller {
	t.Helper()

	controller := &Controller{
		t,
		sync.Mutex{},
		make(map[*mock]struct{}),
		atomic.Bool{},
	}
	t.Cleanup(controller.Finish)

	return controller
}

// ---

// Controller is a mock controller.
type Controller struct {
	tb    testing.TB
	mu    sync.Mutex
	mocks map[*mock]struct{}
	done  atomic.Bool
}

// Checkpoint ensures all pending assertions are satisfied.
func (c *Controller) Checkpoint() {
	c.tb.Helper()

	for mock := range c.mocks {
		mock.mu.Lock()
		defer mock.mu.Unlock()

		for _, call := range mock.exectedCalls[c] {
			if !call.assertion.countConstraint().Contains(call.count.Load()) {
				c.tb.Errorf("mock: missing call %s", call)
			}
		}

		delete(mock.exectedCalls, c)
	}
}

// Finish ensures all assertions are satisfied and forbids adding new assertions.
func (c *Controller) Finish() {
	c.tb.Helper()
	c.done.Store(true)
	c.Checkpoint()
}
