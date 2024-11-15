package intrange

import (
	"fmt"

	"github.com/pamburus/go-tst/internal/constraints"
	"github.com/pamburus/go-tst/internal/optional"
)

func WithMin[T constraints.Integer](value T) PartialRange[T] {
	return Partial(optional.Some(value), optional.None[T]())
}

func WithMax[T constraints.Integer](value T) PartialRange[T] {
	return Partial(optional.None[T](), optional.Some(value))
}

func Partial[T constraints.Integer](min, max optional.Value[T]) PartialRange[T] {
	return PartialRange[T]{min, max}
}

func EmptyPartial[T constraints.Integer]() PartialRange[T] {
	return PartialRange[T]{}
}

// ---

type PartialRange[T constraints.Integer] struct {
	min optional.Value[T]
	max optional.Value[T]
}

func (r PartialRange[T]) Min() optional.Value[T] {
	return r.min
}

func (r PartialRange[T]) Max() optional.Value[T] {
	return r.max
}

func (r PartialRange[T]) WithMin(min T) PartialRange[T] {
	return PartialRange[T]{optional.Some(min), r.max}
}

func (r PartialRange[T]) WithMax(max T) PartialRange[T] {
	return PartialRange[T]{r.min, optional.Some(max)}
}

func (r PartialRange[T]) WithoutMin() PartialRange[T] {
	return PartialRange[T]{optional.None[T](), r.max}
}

func (r PartialRange[T]) WithoutMax() PartialRange[T] {
	return PartialRange[T]{r.min, optional.None[T]()}
}

func (r PartialRange[T]) Contains(value T) bool {
	return optional.Map(r.min, le(value)).OrSome(true) && optional.Map(r.max, ge(value)).OrSome(true)
}

func (r PartialRange[T]) Overlaps(other PartialRange[T]) bool {
	return (r.min.IsNone() || other.max.IsNone() || r.min.OrZero() <= other.max.OrZero()) &&
		(r.max.IsNone() || other.min.IsNone() || r.max.OrZero() >= other.min.OrZero())
}

func (r PartialRange[T]) Equal(other PartialRange[T]) bool {
	return r.min == other.min && r.max == other.max
}

func (r PartialRange[T]) IsEmpty() bool {
	return r.min.IsNone() && r.max.IsNone()
}

func (r PartialRange[T]) String() string {
	switch {
	case r.min == r.max:
		if r.min.IsNone() {
			return ".."
		}
		return fmt.Sprintf("%d", r.min.OrZero())
	case r.min.IsNone():
		return fmt.Sprintf("..%d", r.max.OrZero())
	case r.max.IsNone():
		return fmt.Sprintf("%d..", r.min.OrZero())
	default:
		return fmt.Sprintf("%d..%d", r.min.OrZero(), r.max.OrZero())
	}
}

// ---

func le[T constraints.Integer](b T) func(T) bool {
	return func(a T) bool {
		return a <= b
	}
}

func ge[T constraints.Integer](b T) func(T) bool {
	return func(a T) bool {
		return a >= b
	}
}
