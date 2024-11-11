package tst

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"
)

// New constructs a new Test based on the t.
func New[T HT[T]](tt T) Test {
	return Build(tt).Done()
}

// Build constructs a new Test based on the tt with advanced options.
func Build[T HT[T]](tt T) TestBuilder[T] {
	return TestBuilder[T]{test[T]{
		core: core{
			TB: tt,
			mu: &sync.Mutex{},
		},
	}}.
		WithContext(context.Background())
}

// Current returns the Test associated with the given context.
func Current(ctx context.Context) Test {
	if t, ok := current(ctx); ok {
		return t
	}

	panic("tst: no test is associated with the given context")
}

// Test transparently wraps an object compatible with [testing.TB]
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

	// AddLineTags adds information about the lines of interest to be displayed in test failure message.
	// Do not add lines that are not relevant to the test failure.
	AddLineTags(tags ...LineTag)

	// SetTimeout sets the timeout for the test.
	SetTimeout(timeout time.Duration)

	// SetDeadline sets the deadline for the test.
	SetDeadline(deadline time.Time)

	// Deadline returns the time when the test will be done.
	Deadline() (deadline time.Time, ok bool)

	sealed()
	get() *core
}

// ---

// TestBuilder is a builder for Test.
type TestBuilder[T HT[T]] struct {
	t test[T]
}

// WithContext sets the base context for the test.
func (b TestBuilder[T]) WithContext(ctx context.Context) TestBuilder[T] {
	if ctx != nil {
		ctx, cancel := context.WithCancelCause(ctx)
		b.t.ctx = ctx
		b.t.Cleanup(func() {
			cancel(errTestIsDone)
		})

		b.t.ctx = setupPlugins(b.t.ctx, b.t.TB, b.t.plugins...)
	}

	return b
}

// WithContextFunc sets a function that constructs context from the base test object.
func (b TestBuilder[T]) WithContextFunc(ctxFunc func(T) context.Context) TestBuilder[T] {
	if ctxFunc != nil {
		b.t.ctxFunc = ctxFunc
		b = b.WithContext(ctxFunc(b.t.TB.(T)))
	}

	return b
}

// WithPlugins adds plugins to the test.
func (b TestBuilder[T]) WithPlugins(plugins ...Plugin) TestBuilder[T] {
	b.t.plugins = append(b.t.plugins, plugins...)
	b.t.ctx = setupPlugins(b.t.ctx, &b.t, plugins...)

	return b
}

// Done returns the constructed Test.
func (b TestBuilder[T]) Done() Test {
	setup(&b.t)

	return &b.t
}

// ---

type test[T HT[T]] struct {
	core
	ctxFunc func(T) context.Context
}

func (t *test[T]) Context() context.Context {
	return t.ctx
}

func (t *test[T]) Run(name string, f func(Test)) bool {
	t.Helper()

	return t.TB.(T).Run(name, func(tt T) {
		tt.Helper()
		f(t.fork(tt))
	})
}

func (t *test[T]) RunContext(name string, f func(context.Context, Test)) bool {
	return t.Run(name, func(child Test) {
		f(child.Context(), child)
	})
}

func (t *test[T]) Expect(values ...any) Expectation {
	return Expectation{&t.core, values, CallerLine(1)}
}

func (t *test[T]) AddLineTags(tags ...LineTag) {
	t.addLineTags(tags...)
}

func (t *test[T]) SetTimeout(timeout time.Duration) {
	ctx, cancel := context.WithTimeoutCause(t.ctx, timeout, errTestTimeout)
	t.ctx = ctx
	t.Cleanup(cancel)
}

func (t *test[T]) SetDeadline(deadline time.Time) {
	ctx, cancel := context.WithDeadlineCause(t.ctx, deadline, errTestDeadlineExceeded)
	t.ctx = ctx
	t.Cleanup(cancel)
}

func (t *test[T]) Deadline() (deadline time.Time, ok bool) {
	d1, ok1 := t.ctx.Deadline()
	var d2 time.Time
	var ok2 bool

	if td, ok := t.TB.(interface{ Deadline() (time.Time, bool) }); ok {
		d2, ok2 = td.Deadline()
	}

	if !ok1 && !ok2 {
		return time.Time{}, false
	}

	if !ok1 {
		return d2, true
	}

	if !ok2 {
		return d1, true
	}

	if d1.Before(d2) {
		return d1, true
	}

	return d2, true
}

func (t *test[T]) fork(tt T) *test[T] {
	tt.Helper()

	ctx := t.ctx
	if t.ctxFunc != nil {
		ctx = t.ctxFunc(tt)
	}

	ctx, cancel := context.WithCancelCause(ctx)
	tt.Cleanup(func() { //nolint:wsl // false positive
		cancel(errTestIsDone)
	})

	fork := &test[T]{
		core{tt, ctx, slices.Clip(t.tags), t.plugins, t.mu},
		t.ctxFunc,
	}
	setup(fork)

	return fork
}

func (t *test[T]) sealed() {}

func (t *test[T]) get() *core {
	return &t.core
}

// ---

type core struct {
	testing.TB
	ctx     context.Context
	tags    []LineTag
	plugins []Plugin
	mu      *sync.Mutex
}

func (t *core) addLineTags(tags ...LineTag) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.tags = append(slices.Clip(t.tags), tags...)
}

// ---

func setup[T HT[T]](t *test[T]) {
	t.Helper()

	t.ctx = ctxWithTest(t.ctx, t)
	t.ctx = setupPlugins(t.ctx, t.TB, t.plugins...)

	t.Cleanup(func() {
		if t.Failed() {
			for _, tag := range t.get().tags {
				t.Helper()
				t.Log("See", tag)
			}
		}
	})
}
