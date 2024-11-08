package tst

import (
	"context"
	"testing"
	"time"
)

// New constructs a new Test based on the t.
func New[T HT[T]](tt T) Test {
	return Build(tt)
}

// Test transparently wraps compatible an object compatible with [testing.TB]
// and adds additional methods to make expectations in an assertive way.
type Test interface {
	testing.TB

	// Context returns the context associated with the test.
	// Context is done when the test is done.
	Context() context.Context

	// Run runs sub-test using f.
	Run(name string, f func(Test)) bool

	// RunContext runs sub-test using f feeding it with test's context.
	RunContext(name string, f func(context.Context, Test)) bool

	// Expect begins expectation building process against the given values.
	Expect(values ...any) Expectation

	// WithLineTag returns a new Test that adds information about the line where it was called to the error messages.
	WithLineTag(tag LineTag) Test

	// WithTimeout returns a new Test that adds a timeout to the context.
	WithTimeout(timeout time.Duration) Test

	// WithDeadline returns a new Test that sets deadline to the context.
	WithDeadline(deadline time.Time) Test

	sealed()
}

// Build constructs a new Test based on the tt with advanced options.
func Build[T HT[T]](tt T) TestBuilder[T] {
	return TestBuilder[T]{test[T]{core: core{TB: tt}}}.WithContext(context.Background())
}

// ---

// TestBuilder is a builder for Test.
type TestBuilder[T HT[T]] struct {
	test[T]
}

// WithContext sets the base context for the test.
func (b TestBuilder[T]) WithContext(ctx context.Context) TestBuilder[T] {
	if ctx != nil {
		ctx, cancel := context.WithCancelCause(ctx)
		b.ctx = ctx
		b.Cleanup(func() {
			cancel(errTestIsDone)
		})
		b.ctx = setupPlugins(b.ctx, b.TB, b.plugins...)
	}

	return b
}

// WithContextFunc sets a function that constructs context from the base test object.
func (b TestBuilder[T]) WithContextFunc(ctxFunc func(T) context.Context) TestBuilder[T] {
	if ctxFunc != nil {
		b.ctxFunc = ctxFunc
		b = b.WithContext(ctxFunc(b.TB.(T)))
	}

	return b
}

func (b TestBuilder[T]) WithPlugins(plugins ...Plugin) TestBuilder[T] {
	b.plugins = append(b.plugins, plugins...)
	b.ctx = setupPlugins(b.ctx, b.TB, plugins...)

	return b
}

// Done returns the constructed Test.
func (b TestBuilder[T]) Done() Test {
	return b
}

// ---

type test[T HT[T]] struct {
	core
	ctxFunc func(T) context.Context
}

func (t test[T]) Context() context.Context {
	return t.ctx
}

func (t test[T]) Run(name string, f func(Test)) bool {
	return t.TB.(T).Run(name, func(tt T) {
		f(t.fork(tt))
	})
}

func (t test[T]) RunContext(name string, f func(context.Context, Test)) bool {
	return t.Run(name, func(child Test) {
		f(child.Context(), child)
	})
}

func (t test[T]) Expect(values ...any) Expectation {
	return Expectation{&t.core, values}
}

func (t test[T]) WithLineTag(tag LineTag) Test {
	t.tags = append(t.tags[:len(t.tags):len(t.tags)], tag)

	return t
}

func (t test[T]) WithTimeout(timeout time.Duration) Test {
	ctx, cancel := context.WithTimeoutCause(t.ctx, timeout, errTestTimeout)
	t.ctx = ctx
	t.Cleanup(cancel)

	return t
}

func (t test[T]) WithDeadline(deadline time.Time) Test {
	ctx, cancel := context.WithDeadlineCause(t.ctx, deadline, errTestDeadlineExceeded)
	t.ctx = ctx
	t.Cleanup(cancel)

	return t
}

func (t test[T]) fork(tt T) test[T] {
	ctx := t.ctx
	if t.ctxFunc != nil {
		ctx = t.ctxFunc(tt)
	}
	ctx, cancel := context.WithCancelCause(ctx)
	t.Cleanup(func() {
		cancel(errTestIsDone)
	})

	ctx = setupPlugins(ctx, t.TB, t.plugins...)

	return test[T]{
		core{tt, ctx, t.tags, t.plugins},
		t.ctxFunc,
	}
}

func (t test[T]) sealed() {}

// ---

type core struct {
	testing.TB
	ctx     context.Context
	tags    []LineTag
	plugins []Plugin
}
