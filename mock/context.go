package mock

import "context"

// WithContext creates a new context binding
// that is needed to call alternative functions taking context.
func WithContext(ctx context.Context) ContextBinding {
	return ContextBinding{ctx}
}

// ---

// ContextBinding is a context binding.
type ContextBinding struct {
	ctx context.Context
}

// HandleThisCall handles the current method call of a mock object.
func (b ContextBinding) HandleThisCall(mock AnyMock, args ...any) {
	mock.get().handleCall(b.ctx, 1, callerMethodName(1), args...)
}

// HandleCall handles the method call of a mock object.
func (b ContextBinding) HandleCall(ref AnyMock, method string, args ...any) {
	ref.get().handleCall(b.ctx, 1, method, args...)
}
