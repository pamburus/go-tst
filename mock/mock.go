package mock

import (
	"cmp"
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
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

func (m *Mock[T]) init() {
	m.once.Do(func() {
		te := typeEntryFor[T]()
		m.typ = te.typ
		m.methods = te.methods
	})
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
	init()
}

// ---

type mock struct {
	exectedCalls map[*Controller]map[*callDescriptor]*expectedCall
	calls        []*call //nolint:unused // later // TODO: use
	mu           sync.Mutex
	once         sync.Once
	typ          reflect.Type
	methods      map[string]reflect.Method
}

func (m *mock) get() *mock {
	return m
}

//nolint:unparam // later
func (m *mock) handleCall(ctx context.Context, skip int, methodName string, in []any, out []any) {
	ctx = cmp.Or(ctx, context.Background())
	_ = skip // TODO: use skip

	// TODO: loop over all expected calls and check if the current call matches any of them
	select {
	case <-ctx.Done():
		panic(context.Cause(ctx))
	default:
	}

	method, ok := m.methods[methodName]
	if !ok {
		panic(fmt.Sprintf("mock: type %v has no method %s", m.typ, methodName))
	}

	if len(in) != method.Type.NumIn() {
		panic(fmt.Sprintf("mock: method %s.%s requires %d input parameters, got %d",
			m.typ,
			method.Name,
			method.Type.NumIn(),
			len(in),
		))
	}

	if len(out) != method.Type.NumOut() {
		panic(fmt.Sprintf("mock: method %s.%s requires %d output parameters, got %d",
			m.typ,
			method.Name,
			method.Type.NumOut(),
			len(out),
		))
	}

	for i, arg := range in {
		if reflect.TypeOf(arg) != method.Type.In(i) {
			panic(fmt.Sprintf("mock: method %s.%s requires input parameter #%d to be of type %v, got %v",
				m.typ,
				method.Name,
				i,
				method.Type.In(i),
				reflect.TypeOf(arg),
			))
		}
	}

	for i, arg := range out {
		if reflect.TypeOf(arg) != reflect.PointerTo(method.Type.Out(i)) {
			panic(fmt.Sprintf("mock: method %s.%s requires output parameter #%d to be of type %v, got %v",
				m.typ,
				method.Name,
				i,
				method.Type.Out(i),
				reflect.TypeOf(arg),
			))
		}
	}

	mc := getController(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		matchedCall  *expectedCall
		relatedCalls []*expectedCall
	)

	// if len(m.exectedCalls) == 0 {
	// 	panic("mock: HandleCall called outside of a test")
	// }

	for controller, calls := range m.exectedCalls {
		if mc != nil && mc != controller {
			continue
		}

		if controller.done.Load() {
			continue
		}

		for desc, call := range calls {
			if desc.method.Name != method.Name {
				continue
			}

			relatedCalls = append(relatedCalls, call)

			if !reflect.DeepEqual(call.assertion.args, in) {
				continue
			}

			if !call.registerCall() {
				continue
			}

			matchedCall = call

			break
		}
	}

	if matchedCall == nil {
		var sb strings.Builder
		fmt.Fprintf(&sb, "mock: unexpected call to %s.%s", m.typ, method.Name)
		for _, call := range relatedCalls {
			_, _ = fmt.Fprintf(&sb, "\n(*) See %s", call)
		}

		panic(sb.String())
	}

	inValues := make([]reflect.Value, len(in))
	for i, arg := range in {
		inValues[i] = reflect.ValueOf(arg)
	}

	outValues := matchedCall.assertion.do.Call(inValues)
	for i, arg := range out {
		reflect.ValueOf(arg).Elem().Set(outValues[i])
	}
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

		call = &expectedCall{assertion, atomic.Int64{}, atomic.Bool{}}
		calls[assertion.desc] = call
	}

	controller.mu.Lock()
	defer controller.mu.Unlock()

	controller.mocks[m] = struct{}{}

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
