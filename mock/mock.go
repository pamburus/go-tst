package mock

import (
	"cmp"
	"context"
	"fmt"
	"reflect"
	"sync"
)

// ---

// Mock is a mock object base.
// Should be embedded into a concrete mock object.
type Mock[T any] struct {
	mock
}

// TODO: use
//
//nolint:unused // later
func (m *Mock[T]) getMock() *Mock[T] {
	return m
}

// ---

// AnyMock is a mock object interface.
type AnyMock interface {
	get() *mock
}

// AnyMockFor is a mock object interface for a specific type.
type AnyMockFor[T any] interface {
	AnyMock
	getMock() *Mock[T]
}

// ---

type mock struct {
	exectedCalls map[*Controller]map[*callDescriptor]*expectedCall
	calls        []*call
	mu           sync.Mutex
	once         sync.Once
	typ          reflect.Type
	methods      map[string]reflect.Method
	init         chan struct{}
}

func (m *mock) get() *mock {
	return m
}

//nolint:unparam // `(*mock).handleCall` - `skip` always receives `1`
func (m *mock) handleCall(ctx context.Context, skip int, method string, args ...any) {
	ctx = cmp.Or(ctx, context.Background())

	// TODO: loop over all expected calls and check if the current call matches any of them
	select {
	case <-ctx.Done():
		panic(context.Cause(ctx))
	case <-m.init:
	}

	if _, ok := m.methods[method]; !ok {
		panic(fmt.Sprintf("mock: type %s has no method %s", m.typ, method))
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.exectedCalls) == 0 {
		panic("mock: RegisterCall called outside of a test")
	}

	call := &call{m, method, args, callers(skip + 1)}
	m.calls = append(m.calls, call)
}

func (m *mock) expectCall(t test, assertion CallAssertion) *expectedCall {
	ctx := t.Context()

	controller := getController(ctx)
	if controller == nil {
		panic(fmt.Sprintf("mock: test %s has no registered mock controller, use mock.Plugin() to register it", t.Name()))
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.exectedCalls == nil {
		m.exectedCalls = make(map[*Controller]map[*callDescriptor]*expectedCall)
	}

	calls := m.exectedCalls[controller]
	if calls == nil {
		calls = make(map[*callDescriptor]*expectedCall)
		m.exectedCalls[controller] = calls
	}

	call, ok := calls[assertion.desc]
	if !ok {
		if ctx.Err() != nil {
			panic("mock: can't add expected calls to a test that has already finished")
		}

		for dep := range assertion.after {
			if _, ok := calls[dep]; !ok {
				panic(fmt.Sprintf("mock: expected call %s should be registered before %s", dep, assertion.desc))
			}
		}

		call = &expectedCall{assertion, false, false}
		calls[assertion.desc] = call
	}

	return call
}

// ---

func getController(ctx context.Context) *Controller {
	if value := ctx.Value(ctxKeyController); value != nil {
		return value.(*Controller)
	}

	return nil
}

// ---

var (
	_ AnyMock         = &mock{}
	_ AnyMockFor[any] = &Mock[any]{}
)
