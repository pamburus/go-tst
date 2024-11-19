package mock

import (
	"runtime"
)

// TODO: use
//
//nolint:unused // later
func callers(skip int) []uintptr {
	result := make([]uintptr, 64)

	return result[:runtime.Callers(1+skip, result)]
}
