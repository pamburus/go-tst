package tst_test

import (
	"testing"

	"github.com/pamburus/go-tst/tst"
)

func TestThisLine(t *testing.T) {
	line := tst.ThisLine().String()
	expected := "tst/linetag_test.go:10"
	if line != expected {
		t.Fatalf("Expected ThisLine() to return %s, but got %s", expected, line)
	}
}

func TestCallerLine(t *testing.T) {
	line := tst.CallerLine(0).String()
	expected := "tst/linetag_test.go:18"
	if line != expected {
		t.Fatalf("Expected CallerLine(0) to return %s, but got %s", expected, line)
	}
}

func TestCallerLineSkip1(t *testing.T) {
	f := func(skip int) tst.LineTag {
		return tst.CallerLine(skip + 1)
	}

	line := f(0).String()
	expected := "tst/linetag_test.go:30"
	if line != expected {
		t.Fatalf("Expected CallerLine(1) to return %s, but got %s", expected, line)
	}
}
