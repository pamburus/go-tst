package mock

import (
	"context"
	"testing"
)

func NewPlugin() Plugin {
	return Plugin{}
}

// ---

type T interface {
	testing.TB
	Context() context.Context
}

// ---

type Plugin struct{}

func (Plugin) Configure(ctx context.Context, t testing.TB) context.Context {
	if controller := getController(ctx); controller != nil {
		if controller.t.Name() == t.Name() {
			return ctx
		}

		controller.Suspend()
	}

	return context.WithValue(ctx, ctxKeyController, NewController(t))
}

// ---

type ctxKey int

const (
	ctxKeyController ctxKey = iota
)
