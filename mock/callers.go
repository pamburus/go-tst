package mock

import (
	"runtime"
)

func callers(skip int) []uintptr {
	result := make([]uintptr, 64)
	return result[:runtime.Callers(1+skip, result)]
}
