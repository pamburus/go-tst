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
	"sync/atomic"

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

	mock.init()

	return CallAssertion{
		desc: &callDescriptor{
			mock:    mock.get(),
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

// HandleThisCall handles the current method call of a mock object.
func HandleThisCall[T any](mock AnyMockFor[T], in InputArgs, out OutputArgs) {
	mock.init()
	mock.get().handleCall(context.Background(), 1, callerMethodName(1), in.args, out.args)
}

// HandleCall handles the method call of a mock object.
func HandleCall[T any](mock AnyMockFor[T], method string, in InputArgs, out OutputArgs) {
	mock.init()
	mock.get().handleCall(context.Background(), 1, method, in.args, out.args)
}

// Inputs creates an InputArgs with the given arguments.
func Inputs(args ...any) InputArgs {
	return InputArgs{args: args}
}

// Outputs creates an OutputArgs with the given arguments.
func Outputs(args ...any) OutputArgs {
	return OutputArgs{args: args}
}

// ---

// InputArgs is a list of input arguments in a method call.
type InputArgs struct {
	args []any
}

// String returns a string representation of the input arguments.
func (a InputArgs) String() string {
	return fmt.Sprintf("InputArgs(%v)", a.args)
}

// ---

// OutputArgs is a list of output arguments in a method call.
type OutputArgs struct {
	args []any
}

// String returns a string representation of the output arguments.
func (a OutputArgs) String() string {
	return fmt.Sprintf("OutputArgs(%v)", a.args)
}

// ---

// CallAssertion is an assertion for a call to a mock object.
type CallAssertion struct {
	desc     *callDescriptor
	args     []any
	minCount int
	maxCount int
	after    map[*callDescriptor]struct{}
	do       reflect.Value
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

// Do sets the function to be called when the assertion is satisfied.
func (a CallAssertion) Do(f any) CallAssertion {
	do := reflect.ValueOf(f)
	if do.Kind() != reflect.Func {
		panic("mock: argument must be a function")
	}

	if do.Type().NumIn() != a.desc.method.Type.NumIn() {
		panic(fmt.Sprintf("mock: function for %s.%s must have %d input arguments",
			a.desc.mock.typ,
			a.desc.method.Name,
			a.desc.method.Type.NumIn(),
		))
	}

	if do.Type().NumOut() != a.desc.method.Type.NumOut() {
		panic(fmt.Sprintf("mock: function for %s.%s must have %d output arguments",
			a.desc.mock.typ,
			a.desc.method.Name,
			a.desc.method.Type.NumOut(),
		))
	}

	for i := range do.Type().NumIn() {
		if do.Type().In(i) != a.desc.method.Type.In(i) {
			panic(fmt.Sprintf("mock: function for %s.%s must have input argument %d of type %v",
				a.desc.mock.typ,
				a.desc.method.Name,
				i,
				a.desc.method.Type.In(i),
			))
		}
	}

	for i := range do.Type().NumOut() {
		if do.Type().Out(i) != a.desc.method.Type.Out(i) {
			panic(fmt.Sprintf("mock: function for %s.%s must have output argument %d of type %v",
				a.desc.mock.typ,
				a.desc.method.Name,
				i,
				a.desc.method.Type.Out(i),
			))
		}
	}

	a.do = do

	return a
}

// Return sets the return values for the call.
func (a CallAssertion) Return(values ...any) CallAssertion {
	if a.desc.method.Type.NumIn() != 0 {
		panic("mock: Return must be used only with methods that have no input arguments")
	}

	if a.desc.method.Type.NumOut() != len(values) {
		panic(fmt.Sprintf("mock: %s.%s requires %d return values",
			a.desc.mock.typ,
			a.desc.method.Name,
			a.desc.method.Type.NumOut(),
		))
	}

	for i, value := range values {
		if reflect.TypeOf(value) != a.desc.method.Type.Out(i) {
			panic(fmt.Sprintf("mock: return value %d for %s.%s must be of type %v",
				i,
				a.desc.mock.typ,
				a.desc.method.Name,
				a.desc.method.Type.Out(i),
			))
		}
	}

	a.do = reflect.MakeFunc(a.desc.method.Type, func([]reflect.Value) []reflect.Value {
		result := make([]reflect.Value, len(values))
		for i, value := range values {
			result[i] = reflect.ValueOf(value)
		}

		return result
	})

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
	count     atomic.Int64
	completed atomic.Bool
}

func (c *expectedCall) registerCall() bool {
	var count int64

	for {
		count = c.count.Load()
		if c.assertion.maxCount > 0 && count >= int64(c.assertion.maxCount) {
			return false
		}

		count++
		if c.count.CompareAndSwap(count-1, count) {
			break
		}
	}

	return count >= int64(c.assertion.minCount)
}

func (c *expectedCall) String() string {
	var sb strings.Builder
	sb.WriteString(c.assertion.desc.String())
	if c.assertion.minCount > 0 || c.assertion.maxCount > 0 {
		fmt.Fprintf(&sb, " called %d of %d..%d times", c.count.Load(), c.assertion.minCount, c.assertion.maxCount)
	}

	return sb.String()
}

// ---

type callDescriptor struct {
	mock    *mock
	method  reflect.Method
	lineTag tst.LineTag
}

func (d *callDescriptor) String() string {
	return fmt.Sprintf("%v.%v defined at %s", d.mock.typ, d.method.Name, d.lineTag)
}

// ---

// TODO: use
//
//nolint:unused // later
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
		entry.typ = typ
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
