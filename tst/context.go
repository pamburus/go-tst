package tst

import (
	"context"
	"reflect"
)

// ---

func ctxPut[T any](ctx context.Context, t T) context.Context {
	return context.WithValue(ctx, ctxTypeKey{reflect.TypeFor[T]()}, t)
}

func ctxGet[T any](ctx context.Context) (T, bool) {
	var zero T
	if value := ctx.Value(ctxTypeKey{reflect.TypeFor[T]()}); value != nil {
		return value.(T), true
	}

	return zero, false
}

// ---

type ctxTypeKey struct {
	t reflect.Type
}
