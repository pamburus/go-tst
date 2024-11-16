package tst

import (
	"context"
	"testing"
	"time"

	"github.com/pamburus/go-tst/tst/constraints"
)

// New constructs a new Test based on the t.
func New[T constraints.HT[T]](tt T) Test {
	return Build(tt).Done()
}

// NewT constructs a new T based on the tt.
func NewT[TT constraints.T[TT]](tt TT) T {
	return &ti[TT]{Build(tt).done().tb, tt}
}

// Build constructs a new Test based on the tt with advanced options.
func Build[T constraints.HT[T]](tt T) TestBuilder[T] {
	return TestBuilder[T]{test[T]{&tb[T]{core: core{TB: tt}}}}.
		WithContext(context.Background())
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

// Done returns the constructed Test.
func (b TestBuilder[T]) Done() Test {
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
	return t.run(name, func(tt *tb[T]) {
		f(&test[T]{tt})
	})
}

func (t *test[T]) RunContext(name string, f func(context.Context, Test)) bool {
	return t.runContext(name, func(ctx context.Context, tt *tb[T]) {
		f(ctx, &test[T]{tt})
	})
}

// ---

type ti[X constraints.T[X]] struct {
	*tb[X]
	tHelper
}

func (t *ti[X]) Run(name string, f func(T)) bool {
	return t.run(name, func(tt *tb[X]) {
		f(&ti[X]{tt, tt})
	})
}

func (t *ti[X]) RunContext(name string, f func(context.Context, T)) bool {
	return t.runContext(name, func(ctx context.Context, tt *tb[X]) {
		f(ctx, &ti[X]{tt, tt})
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

func (t *tb[T]) runContext(name string, f func(context.Context, *tb[T])) bool {
	return t.run(name, func(child *tb[T]) {
		f(child.Context(), child)
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

	fork := &tb[T]{core{tt, ctx, t.tags}, t.ctxFunc}
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
	ctx  context.Context
	tags []LineTag
}

func (c *core) addLineTags(tags ...LineTag) {
	c.tags = append(c.tags, tags...)
}

// ---

func setup[T constraints.HT[T]](t *tb[T]) {
	t.Helper()
	t.Cleanup(func() {
		if t.Failed() {
			for _, tag := range t.get().tags {
				t.Helper()
				t.Log("See", tag)
			}
		}
	})
}
