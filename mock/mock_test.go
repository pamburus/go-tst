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

func (m *StringerMock) String() string {
	mock.HandleThisCall(m)

	return ""
}

func TestStringerMock(tt *testing.T) {
	t := tst.Build(tt).
		WithPlugins(mock.NewPlugin()).
		Done()

	m := &StringerMock{}

	mock.Expect(t, mock.Call(m, "String"))
}
