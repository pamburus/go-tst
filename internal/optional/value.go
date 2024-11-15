package optional

func New[T any](value T, valid bool) Value[T] {
	return Value[T]{value, valid}
}

func Some[T any](value T) Value[T] {
	return Value[T]{value, true}
}

func None[T any]() Value[T] {
	return Value[T]{}
}

func Map[T, U any](v Value[T], f func(T) U) Value[U] {
	if v.valid {
		return Some(f(v.inner))
	}

	return Value[U]{}
}

// ---

type Value[T any] struct {
	inner T
	valid bool
}

func (v Value[T]) Unwrap() (T, bool) {
	return v.inner, v.valid
}

func (v Value[T]) IsSome() bool {
	return v.valid
}

func (v Value[T]) IsNone() bool {
	return !v.valid
}

func (v Value[T]) OrZero() T {
	return v.inner
}

func (v Value[T]) OrSome(defaultValue T) T {
	if v.valid {
		return v.inner
	}

	return defaultValue
}

func (v *Value[T]) Set(value T) {
	v.inner = value
	v.valid = true
}

func (v *Value[T]) Reset() {
	var zero T
	v.inner = zero
	v.valid = false
}
