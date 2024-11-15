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

	// Example data: []int{3, 1, 4}
	mock.Expect(t, mock.InOrder(
		mock.Call(m, "Len").Return(3),
		mock.Call(m, "Less", 1, 0).Return(true),
		mock.Call(m, "Swap", 1, 0),
		mock.Call(m, "Less", 2, 1).Return(false),
	))

	sort.Sort(m)
}

func TestStringerMock(tt *testing.T) {
	t := Build(tt).
		WithPlugins(mock.NewPlugin()).
		Done()

	m := &stringerMock{}

	mock.Expect(t,
		mock.Call(m, "String").Return("hello"),
	)

	t.Expect(m.String()).To(Equal("hello"))
	t.Expect(m.String).To(PanicWith(
		String(ContainingSubstring("mock: unexpected call to fmt.Stringer.String")),
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
