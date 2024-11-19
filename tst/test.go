package tst

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/pamburus/go-tst/tst/constraints"
)

// New constructs a new [Test] based on the t.
func New[T constraints.HT[T]](tt T) Test {
	return Build(tt).Done()
}

// NewT constructs a new [T] based on the tt.
func NewT[TT constraints.T[TT]](tt TT) T {
	return &ti[TT]{Build(tt).done().tb, tt}
}

// Build constructs a new [Test] based on the tt with advanced options.
func Build[T constraints.HT[T]](tt T) TestBuilder[T] {
	return TestBuilder[T]{test[T]{&tb[T]{
		core: core{
			TB: tt,
			mu: &sync.Mutex{},
		},
	}}}.
		WithContext(context.Background())
}

// CurrentTest returns the [Test] associated with the given context.
func CurrentTest(ctx context.Context) Test {
	if t, ok := ctxGet[Test](ctx); ok {
		return t
	}

	panic("tst: no test is associated with the given context")
}

// CurrentT returns the [T] associated with the given context.
func CurrentT(ctx context.Context) T {
	if t, ok := ctxGet[T](ctx); ok {
		return t
	}

	panic("tst: no test is associated with the given context")
}

// ---

// TB transparently extends [testing.TB] interface with additional methods.
// Adds support for context, expectation building, line tags, timeouts and deadlines.
type TB interface {
	testing.TB

	// Context returns the context associated with the test.
	// Context is done when the test is done.
	Context() context.Context

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

// Test extends [TB] with methods for hierarchical test execution taking [Test].
type Test interface {
	HT[Test]
}

// T extends [TB] with methods for hierarchical test execution taking [T].
type T interface {
	HT[T]
	constraints.T[T]
}

// HT extends [TB] with methods for hierarchical test execution taking provided type parameter T.
type HT[T constraints.HT[T]] interface {
	TB
	constraints.HT[T]

	// RunContext runs sub-test using f feeding it with test's context.
	RunContext(name string, f func(context.Context, T)) bool
}

// ---

// TestBuilder is a builder for Test.
type TestBuilder[T constraints.HT[T]] struct {
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
	b.t.Helper()
	setup(b.t.tb)

	return &b.t
}

func (b TestBuilder[T]) done() *test[T] {
	return &b.t
}

// ---

type test[T constraints.HT[T]] struct {
	*tb[T]
}

func (t *test[T]) Run(name string, f func(Test)) bool {
	t.Helper()

	return t.run(name, func(tt *tb[T]) {
		child := &test[T]{tt}
		child.ctx = ctxPut[Test](tt.ctx, child)
		f(child)
	})
}

func (t *test[T]) RunContext(name string, f func(context.Context, Test)) bool {
	t.Helper()

	return t.Run(name, func(tt Test) {
		f(tt.Context(), tt)
	})
}

// ---

type ti[X constraints.T[X]] struct {
	*tb[X]
	tHelper
}

func (t *ti[X]) Run(name string, f func(T)) bool {
	t.Helper()

	return t.run(name, func(tt *tb[X]) {
		child := &ti[X]{tt, tt}
		child.ctx = ctxPut[T](tt.ctx, child)
		f(child)
	})
}

func (t *ti[X]) RunContext(name string, f func(context.Context, T)) bool {
	t.Helper()

	return t.Run(name, func(tt T) {
		f(tt.Context(), tt)
	})
}

func (t *ti[X]) Parallel() {
	t.Inner().Parallel()
}

// ---

type tHelper interface {
	Helper()
}

// ---

type tb[T constraints.HT[T]] struct {
	core
	ctxFunc func(T) context.Context
}

func (t *tb[T]) Inner() T {
	return t.TB.(T)
}

func (t *tb[T]) Context() context.Context {
	return t.ctx
}

func (t *tb[T]) Expect(values ...any) Expectation {
	return Expectation{&t.core, values, CallerLine(1)}
}

func (t *tb[T]) AddLineTags(tags ...LineTag) {
	t.addLineTags(tags...)
}

func (t *tb[T]) SetTimeout(timeout time.Duration) {
	ctx, cancel := context.WithTimeoutCause(t.ctx, timeout, errTestTimeout)
	t.ctx = ctx
	t.Cleanup(cancel)
}

func (t *tb[T]) SetDeadline(deadline time.Time) {
	ctx, cancel := context.WithDeadlineCause(t.ctx, deadline, errTestDeadlineExceeded)
	t.ctx = ctx
	t.Cleanup(cancel)
}

func (t *tb[T]) Deadline() (deadline time.Time, ok bool) {
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

func (t *tb[T]) run(name string, f func(*tb[T])) bool {
	t.Helper()

	return t.TB.(T).Run(name, func(tt T) {
		tt.Helper()
		f(t.fork(tt))
	})
}

func (t *tb[T]) fork(tt T) *tb[T] {
	tt.Helper()

	ctx := t.ctx
	if t.ctxFunc != nil {
		ctx = t.ctxFunc(tt)
	}

	ctx, cancel := context.WithCancelCause(ctx)
	tt.Cleanup(func() { //nolint:wsl // false positive
		cancel(errTestIsDone)
	})

	fork := &tb[T]{
		core{tt, ctx, slices.Clip(t.tags), slices.Clip(t.plugins), &sync.Mutex{}},
		t.ctxFunc,
	}
	setup(fork)

	return fork
}

func (t *tb[T]) sealed() {}

func (t *tb[T]) get() *core {
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

func setup[T constraints.HT[T]](t *tb[T]) {
	t.Helper()

	t.ctx = ctxPut(t.ctx, t)
	t.ctx = ctxPut(t.ctx, &t.core)
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
