package tst

import (
	"context"
	"testing"
)

type Plugin interface {
	Configure(context.Context, testing.TB) context.Context
}

// ---

func setupPlugins(ctx context.Context, t testing.TB, plugins ...Plugin) context.Context {
	for _, plugin := range plugins {
		ctx = plugin.Configure(ctx, t)
	}

	return ctx
}
