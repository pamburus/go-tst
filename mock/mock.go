package mock

import (
	"cmp"
	"context"
	"runtime"
	"strings"
	"sync"
)

// ---

type Mock[T any] struct {
	mock
}

func (m *Mock[T]) getMock() *Mock[T] {
	return m
}

// ---

type AnyMock interface {
	get() *mock
}

type AnyMockFor[T any] interface {
	AnyMock
	getMock() *Mock[T]
}

func RegisterThisCall(mock AnyMock, args ...any) {
	mock.get().registerCall(nil, 1, callerMethodName(1), args...)
}

func RegisterThisCallContext(ctx context.Context, mock AnyMock, args ...any) {
	mock.get().registerCall(ctx, 1, callerMethodName(1), args...)
}

func RegisterCall(ref AnyMock, method string, args ...any) {
	ref.get().registerCall(nil, 1, method, args...)
}

func RegisterCallContext(ctx context.Context, ref AnyMock, method string, args ...any) {
	ref.get().registerCall(ctx, 1, method, args...)
}

// ---

type mock struct {
	exectedCalls map[*Controller][]*ExpectedCall
	calls        []*call
	mu           sync.Mutex
}

func (m *mock) get() *mock {
	return m
}

func (m *mock) registerCall(ctx context.Context, skip int, method string, args ...any) {
	ctx = cmp.Or(ctx, context.Background())

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.exectedCalls) == 0 {
		panic("mock: RegisterCall called outside of a test")
	}

	call := &call{m, method, args, callers(skip + 1)}
	m.calls = append(m.calls, call)
}

// ---

func callerMethodName(skip int) string {
	pc, _, _, ok := runtime.Caller(1 + skip)
	if !ok {
		return ""
	}

	method := runtime.FuncForPC(pc).Name()
	if i := strings.LastIndexByte(method, '.'); i >= 0 {
		method = method[i+1:]
	}

	return method
}

// ---

var (
	_ AnyMock         = &mock{}
	_ AnyMockFor[any] = &Mock[any]{}
)
