package mock

import "testing"

func NewController(t testing.TB) *Controller {
	controller := &Controller{
		t,
		make(map[*mock]struct{}),
	}
	t.Cleanup(controller.Finish)

	return controller
}

type Controller struct {
	t     testing.TB
	mocks map[*mock]struct{}
}

func (c *Controller) Checkpoint() {
}

func (c *Controller) Finish() {
}

func (c *Controller) Suspend() {
}
