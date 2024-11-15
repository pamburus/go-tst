package mock_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/pamburus/go-tst/mock"
	. "github.com/pamburus/go-tst/tst"
)

func TestSortMock(tt *testing.T) {
	t := Build(tt).
		WithPlugins(mock.NewPlugin()).
		Done()

	m := &sortMock{}

	t.Run("T1", func(t Test) {
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

	t.Run("T2", func(t Test) {
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
	t := Build(tt).
		WithPlugins(mock.NewPlugin()).
		Done()

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
