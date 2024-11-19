package mock

import "testing"

// NewController creates a new mock controller.
func NewController(t testing.TB) *Controller {
	t.Helper()

	controller := &Controller{
		t,
		make(map[*mock]struct{}),
	}
	t.Cleanup(controller.Finish)

	return controller
}

// ---

// Controller is a mock controller.
type Controller struct {
	tb    testing.TB
	mocks map[*mock]struct{}
}

// Checkpoint ensures all pending assertions are satisfied.
func (c *Controller) Checkpoint() {
	// TODO: Implement this.
}

// Finish ensures all assertions are satisfied and forbids adding new assertions.
func (c *Controller) Finish() {
}
