package tst

import (
	"context"
	"testing"
)

type Plugin interface {
	Configure(context.Context, testing.TB) context.Context
}
