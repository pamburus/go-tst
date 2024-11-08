package mock

type ExpectedCall struct {
	mock      *mock
	method    string
	args      []any
	count     int
	minCount  int
	maxCount  int
	satisfied bool
}

// ---

type call struct {
	mock    *mock
	method  string
	args    []any
	callers []uintptr
}
