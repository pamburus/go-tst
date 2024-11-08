package mock

import "context"

func WithContext(ctx context.Context) ContextBinding {
	return ContextBinding{ctx}
}

// ---

type ContextBinding struct {
	ctx context.Context
}

func (b ContextBinding) HandleThisCall(mock AnyMock, args ...any) {
	mock.get().handleCall(b.ctx, 1, callerMethodName(1), args...)
}

func (b ContextBinding) HandleCall(ref AnyMock, method string, args ...any) {
	ref.get().handleCall(b.ctx, 1, method, args...)
}
