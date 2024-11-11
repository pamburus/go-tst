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

func setupPlugins(ctx context.Context, tb testing.TB, plugins ...Plugin) context.Context {
	tb.Helper()

	for _, plugin := range plugins {
		ctx = plugin.Configure(ctx, tb)
	}

	return ctx
}
