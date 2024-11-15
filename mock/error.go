package mock

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func IsUnexpectedCall(err error) bool {
	var e errUnexpectedCallError

	return errors.As(err, &e)
}

// ---

type errUnexpectedCallError struct {
	typ          reflect.Type
	method       reflect.Method
	args         []any
	relatedCalls []*expectedCall
}

func (e errUnexpectedCallError) Error() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "mock: unexpected call %s.%s", e.typ, e.method.Name)

	sb.WriteByte('(')
	for i, arg := range e.args {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%#v", arg))
	}
	sb.WriteByte(')')

	for _, call := range e.relatedCalls {
		_, _ = fmt.Fprintf(&sb, "\n(*) See %s", call)
	}

	return sb.String()
}
