package tst

import (
	"fmt"
	"runtime"
	"strings"
)

// ThisLine returns a LineTag that represents the line where it was called.
func ThisLine() LineTag {
	return callerLine(0)
}

// CallerLine returns a LineTag that represents the line where the call was made
// skipping the specified number of callers above.
func CallerLine(skip int) LineTag {
	return callerLine(skip + 1)
}

// LineTag represents a line in the source code.
type LineTag struct {
	pc uintptr
}

// IsZero returns true if the LineTag is zero.
func (t LineTag) IsZero() bool {
	return t.pc == 0
}

// String returns a string representation of the LineTag.
func (t LineTag) String() string {
	fn := runtime.FuncForPC(t.pc)
	file, line := fn.FileLine(t.pc)

	e := len(file)
	for i := 0; i != 2 && e > 0; i++ {
		e = strings.LastIndexByte(file[:e], '/')
	}
	if e > 0 {
		file = file[e+1:]
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// ---

func callerLine(skip int) LineTag {
	var pcs [1]uintptr
	runtime.Callers(2+skip, pcs[:])

	return LineTag{pcs[0]}
}
