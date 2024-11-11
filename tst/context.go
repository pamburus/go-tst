package tst

import "context"

// ---

func ctxWithTest(ctx context.Context, t Test) context.Context {
	return context.WithValue(ctx, ctxKeyTest, t)
}

func current(ctx context.Context) (Test, bool) {
	if value := ctx.Value(ctxKeyTest); value != nil {
		return value.(Test), true
	}

	return nil, false
}

// ---

type ctxKey int

const (
	ctxKeyTest ctxKey = iota
)
