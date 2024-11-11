package mock_test

import (
	"fmt"
	"testing"

	"github.com/pamburus/go-tst/mock"
	"github.com/pamburus/go-tst/tst"
)

type StringerMock struct {
	mock.Mock[fmt.Stringer]
}

func (m *StringerMock) String() (result string) {
	mock.HandleThisCall(m, mock.Inputs(), mock.Outputs(&result))

	return
}

func TestStringerMock(tt *testing.T) {
	t := tst.Build(tt).
		WithPlugins(mock.NewPlugin()).
		Done()

	m := &StringerMock{}

	mock.Expect(t,
		mock.Call(m, "String").Return("hello"),
	)
}
