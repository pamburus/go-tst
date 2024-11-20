package mock

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/TheCount/go-multilocker/multilocker"
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

func (c *Controller) AtomicCheck(t test, f func(), assertions ...Assertion) {
	c.tb.Helper()

	lockers := make(map[sync.Locker]struct{}, 1+len(assertions))
	lockers[&c.mu] = struct{}{}
	for _, a := range assertions {
		for l := range a.lockers() {
			lockers[l] = struct{}{}
		}
	}

	ml := multilocker.New(keys(lockers)...)
	ml.Lock()
	defer ml.Unlock()

	expect(t, locked, assertions...)

	defer func() {
		if pv := recover(); pv != nil {
			if err, ok := pv.(error); ok && IsUnexpectedCallError(err) {
				t.Error(err)
			} else {
				panic(pv)
			}
		}
	}()

	f()

	c.checkpoint()
}

// Checkpoint ensures all pending assertions are satisfied.
func (c *Controller) Checkpoint() {
	c.tb.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.checkpoint()
}

// Finish ensures all assertions are satisfied and forbids adding new assertions.
func (c *Controller) Finish() {
	c.tb.Helper()
	c.done.Store(true)
	c.Checkpoint()
}

func (c *Controller) checkpoint() {
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

func (c *Controller) addMock(m *mock, lock lockStatus) {
	c.tb.Helper()

	if c.done.Load() {
		panic("mock: can't add expected calls to a test that has already finished")
	}

	if lock == unlocked {
		c.mu.Lock()
		defer c.mu.Unlock()
	}

	c.mocks[m] = struct{}{}
}
