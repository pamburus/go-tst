package mock

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/pamburus/go-tst/tst"
)

// Call creates a CallAssertion for the given mock and method.
func Call[T any](mock AnyMockFor[T], method string, args ...any) CallAssertion {
	te := typeEntryFor[T]()

	methodType, ok := te.methods[method]
	if !ok {
		panic(fmt.Sprintf("mock: method %v not found on type %v", method, te.typ))
	}

	if len(args) > methodType.Type.NumIn() {
		panic(fmt.Sprintf("mock: too many arguments for method %v", method))
	}

	if !reflect.TypeOf(mock).Implements(te.typ) {
		panic(fmt.Sprintf("mock: %v does not implement %v", reflect.TypeOf(mock), te.typ))
	}

	for i := range args {
		if _, ok := args[i].(tst.Assertion); ok {
			continue
		}

		args[i] = tst.Equal(args[i])
	}

	m := mock.get()
	m.once.Do(func() {
		m.typ = te.typ
		m.methods = te.methods
		close(m.init)
	})

	return CallAssertion{
		desc: &callDescriptor{
			mock:    m,
			method:  methodType,
			lineTag: tst.CallerLine(1),
		},
		args: args,
	}
}

// InOrder creates an assertion that the given calls are made in the specified order.
func InOrder(calls ...CallAssertion) Assertion {
	for i := 1; i < len(calls); i++ {
		calls[i].After(calls[i-1])
	}

	return assertionFunc(func(t test) {
		for _, call := range calls {
			call.setup(t)
		}
	})
}

// HandleThisCall handles the current method call of the mock object.
func HandleThisCall[T any](mock AnyMockFor[T], args ...any) {
	mock.get().handleCall(context.Background(), 1, callerMethodName(1), args...)
}

// HandleCall handles the method call of the mock object.
func HandleCall[T any](mock AnyMockFor[T], method string, args ...any) {
	mock.get().handleCall(context.Background(), 1, method, args...)
}

// ---

// CallAssertion is an assertion for a call to a mock object.
type CallAssertion struct {
	desc     *callDescriptor
	args     []any
	minCount int
	maxCount int
	after    map[*callDescriptor]struct{}
}

// After sets the order of the current call after the other call.
func (a CallAssertion) After(other CallAssertion) {
	if a.after == nil {
		a.after = make(map[*callDescriptor]struct{})
	}

	a.after[other.desc] = struct{}{}
}

// Before sets the order of the current call before the other call.
func (a CallAssertion) Before(other CallAssertion) {
	other.After(a)
}

// Times sets the number of times the call is expected to be made.
func (a CallAssertion) Times(count int) CallAssertion {
	a.minCount = count
	a.maxCount = count

	return a
}

// AtLeast sets the minimum number of times the call is expected to be made.
func (a CallAssertion) AtLeast(count int) CallAssertion {
	a.minCount = count

	return a
}

// AtMost sets the maximum number of times the call is expected to be made.
func (a CallAssertion) AtMost(count int) CallAssertion {
	a.maxCount = count

	return a
}

func (a CallAssertion) setup(t test) {
	if a.desc == nil {
		panic("mock: ExpectedCall must be created by Call function")
	}

	a.desc.mock.expectCall(t, a)
}

// TODO: use
//
//nolint:unused // later
func (a CallAssertion) clone() CallAssertion {
	a.args = slices.Clone(a.args)
	a.after = maps.Clone(a.after)

	return a
}

// ---

type expectedCall struct {
	assertion CallAssertion
	satisfied bool
	completed bool
}

// ---

type callDescriptor struct {
	mock    *mock
	method  reflect.Method
	lineTag tst.LineTag
}

func (d *callDescriptor) String() string {
	return fmt.Sprintf("%v.%v defined at %s", d.mock.typ.Name(), d.method.Name, d.lineTag)
}

// ---

type call struct {
	mock    *mock
	method  string
	args    []any
	callers []uintptr
}

// ---

type typeEntry struct {
	typ     reflect.Type
	once    sync.Once
	methods map[string]reflect.Method
}

// ---

func typeEntryFor[T any]() *typeEntry {
	typ := reflect.TypeFor[T]()
	if typ.Kind() != reflect.Interface {
		panic(fmt.Sprintf("mock: type %v is not an interface type", typ))
	}

	entryItem, _ := typeRegistry.LoadOrStore(typ, &typeEntry{typ: typ})

	entry := entryItem.(*typeEntry)
	entry.once.Do(func() {
		entry.methods = make(map[string]reflect.Method)

		for i := range typ.NumMethod() {
			method := typ.Method(i)
			entry.methods[method.Name] = method
		}
	})

	return entry
}

// TODO: use
//
//nolint:unused // later
func methodFor[T any](name string) reflect.Method {
	te := typeEntryFor[T]()

	method, ok := te.methods[name]
	if !ok {
		panic(fmt.Sprintf("mock: %v does not have method %v", te.typ, name))
	}

	return method
}

// TODO: use
//
//nolint:unused // later
func callerMethodFor[T any](skip int) reflect.Method {
	te := typeEntryFor[T]()
	name := callerMethodName(skip + 1)

	method, ok := te.methods[name]
	if !ok {
		panic(fmt.Sprintf("mock: %v does not have method %v", te.typ, name))
	}

	return method
}

func callerMethodName(skip int) string {
	pc, _, _, ok := runtime.Caller(1 + skip)
	if !ok {
		return ""
	}

	method := runtime.FuncForPC(pc).Name()
	if i := strings.LastIndexByte(method, '.'); i >= 0 {
		method = method[i+1:]
	}

	return method
}

// ---

var typeRegistry sync.Map