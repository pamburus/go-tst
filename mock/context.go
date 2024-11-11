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
func (b ContextBinding) HandleThisCall(mock AnyMock, in InputArgs, out OutputArgs) {
	mock.get().handleCall(b.ctx, 1, callerMethodName(1), in.args, out.args)
}

// HandleCall handles the method call of a mock object.
func (b ContextBinding) HandleCall(ref AnyMock, method string, in InputArgs, out OutputArgs) {
	ref.get().handleCall(b.ctx, 1, method, in.args, out.args)
}
