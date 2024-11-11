package tst

import (
	"context"
	"testing"
)

// Plugin is a plug-in for [Test].
type Plugin interface {
	Configure(context.Context, testing.TB) context.Context
}

// ---

func setupPlugins(ctx context.Context, t testing.TB, plugins ...Plugin) context.Context {
	t.Helper()

	for _, plugin := range plugins {
		ctx = plugin.Configure(ctx, t)
	}

	return ctx
}
