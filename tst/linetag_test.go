package tst_test

import (
	"strings"
	"testing"

	"github.com/pamburus/go-tst/tst"
)

func TestThisLine(t *testing.T) {
	line, expected := tst.ThisLine().String(), "/tst/linetag_test.go:11"
	if line != strings.TrimPrefix(expected, "/") && !strings.HasSuffix(line, expected) {
		t.Fatalf("Expected ThisLine() = %q to end with %q", line, expected)
	}
}

func TestCallerLine(t *testing.T) {
	line, expected := tst.CallerLine(0).String(), "/tst/linetag_test.go:18"
	if line != strings.TrimPrefix(expected, "/") && !strings.HasSuffix(line, expected) {
		t.Fatalf("Expected CallerLine(0) = %q to end with %q", line, expected)
	}
}

func TestCallerLineSkip1(t *testing.T) {
	f := func(skip int) tst.LineTag {
		return tst.CallerLine(skip + 1)
	}

	line, expected := f(0).String(), "/tst/linetag_test.go:29"
	if line != strings.TrimPrefix(expected, "/") && !strings.HasSuffix(line, expected) {
		t.Fatalf("Expected CallerLine(1) = %q to end with %q", line, expected)
	}
}
