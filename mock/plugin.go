package mock

import (
	"context"
	"testing"
)

// NewPlugin creates a new plug-in.
func NewPlugin() Plugin {
	return Plugin{}
}

// ---

// Plugin is a plug-in for [Test].
type Plugin struct{}

// Configure configures the plug-in.
func (Plugin) Configure(ctx context.Context, t testing.TB) context.Context {
	t.Helper()

	if controller := getController(ctx); controller != nil {
		if controller.tb.Name() == t.Name() {
			return ctx
		}

		controller.Checkpoint()
	}

	return context.WithValue(ctx, ctxKeyController, NewController(t))
}

// ---

type test interface {
	testing.TB
	Context() context.Context
}

// ---

type ctxKey int

const (
	ctxKeyController ctxKey = iota
)
