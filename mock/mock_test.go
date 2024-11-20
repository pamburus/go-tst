package mock_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/pamburus/go-tst/mock"
	. "github.com/pamburus/go-tst/tst"
)

func TestSortMock(tt *testing.T) {
	t := NewT(tt).
		WithPlugins(mock.NewPlugin())

	m := &sortMock{}

	t.Run("T1", func(t T) {
		t.Parallel()
		// Example data: []int{3, 1, 4}
		mock.On(t).During(func() {
			sort.Sort(m)
		}).Expect(mock.InOrder(
			mock.Call(m, "Len").Return(3),
			mock.Call(m, "Less", 1, 0).Return(true),
			mock.Call(m, "Swap", 1, 0),
			mock.Call(m, "Less", 2, 1).Return(false),
		))
	})

	t.Run("T2", func(t T) {
		t.Parallel()
		// Example data: []int{3, 1, 4}
		mock.On(t).During(func() {
			sort.Sort(m)
		}).Expect(mock.InOrder(
			mock.Call(m, "Len").Return(3),
			mock.Call(m, "Less", 1, 0).Return(true),
			mock.Call(m, "Swap", 1, 0),
			mock.Call(m, "Less", 2, 1).Return(false),
		))
	})
}

func TestStringerMock(tt *testing.T) {
	t := NewT(tt).
		WithPlugins(mock.NewPlugin())

	m := &stringerMock{}

	mock.On(t).During(func() {
		t.Expect(m.String()).To(Equal("hello"))
	}).Expect(
		mock.Call(m, "String").Return("hello"),
	)

	mock.On(t).During(func() {
		t.Expect(m.String()).To(Equal("bye"))
	}).Expect(
		mock.Call(m, "String").Return("bye"),
	)

	mock.On(t).During(func() {
		t.Expect(m.String()).To(Equal("hello"))
		t.Expect(m.String()).To(Equal("bye"))
	}).Expect(mock.InOrder(
		mock.Call(m, "String").Return("hello"),
		mock.Call(m, "String").Return("bye"),
	))

	t.Expect(m.String).To(PanicAndPanicValueTo(
		MatchErrorBy(mock.IsUnexpectedCallError, "IsUnexpectedCallError"),
	))
}

func TestContextDescriptorMock(tt *testing.T) {
	t := NewT(tt).
		WithPlugins(mock.NewPlugin())

	m := &contextDescriptorMock{}

	t.RunContext("T1", func(ctx context.Context, t T) {
		t.Parallel()

		mock.On(t).During(func() {
			t.Expect(m.Describe(ctx)).To(Equal("hello"))
			t.Expect(m.Describe(ctx)).To(Equal("friend"))
		}).Expect(
			mock.Call(m, "Describe", ctx).Return("hello"),
			mock.Call(m, "Describe", ctx).Return("friend"),
		)
	})

	t.RunContext("T2", func(ctx context.Context, t T) {
		t.Parallel()

		mock.On(t).During(func() {
			t.Expect(m.Describe(ctx)).To(Equal("bye"))
			t.Expect(m.Describe(ctx)).To(Equal("stranger"))
		}).Expect(
			mock.Call(m, "Describe", ctx).Return("bye"),
			mock.Call(m, "Describe", ctx).Return("stranger"),
		)
	})

	// t.RunContext("T3", func(ctx context.Context, t T) {
	// 	t.Parallel()

	// 	mock.Expect(t,
	// 		mock.Call(m, "Describe", ctx).Return("bye"),
	// 		mock.Call(m, "Describe", ctx).Return("stranger"),
	// 	)
	// 	t.Expect(m.Describe(ctx)).To(Equal("bye"))
	// 	t.Expect(m.Describe(ctx)).To(Equal("stranger"))
	// })
}

// ---

var _ fmt.Stringer = &stringerMock{}

type stringerMock struct {
	mock.Mock[fmt.Stringer]
}

func (m *stringerMock) String() (result string) {
	mock.HandleThisCall(m, mock.Inputs(), mock.Outputs(&result))
	return
}

// ---

var _ sort.Interface = &sortMock{}

type sortMock struct {
	mock.Mock[sort.Interface]
}

func (m *sortMock) Len() (result int) {
	mock.HandleThisCall(m, mock.Inputs(), mock.Outputs(&result))
	return
}

func (m *sortMock) Less(i, j int) (result bool) {
	mock.HandleThisCall(m, mock.Inputs(i, j), mock.Outputs(&result))
	return
}

func (m *sortMock) Swap(i, j int) {
	mock.HandleThisCall(m, mock.Inputs(i, j), mock.Outputs())
}

// ---

type contextDescriptor interface {
	Describe(context.Context) string
}

var _ contextDescriptor = &contextDescriptorMock{}

type contextDescriptorMock struct {
	mock.Mock[contextDescriptor]
}

func (m *contextDescriptorMock) Describe(ctx context.Context) (result string) {
	mock.HandleThisCall(m, mock.Inputs(ctx), mock.Outputs(&result))
	return
}
