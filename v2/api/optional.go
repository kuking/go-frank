package api

import "fmt"

type Optional[T any] interface {
	IsPresent() bool
	IsEmpty() bool
	Get() (value T, found bool)
	OrElse(other T) T
	OrPanic() T
	IfPresent(op func(T)) (found bool)
	Map(op func(Optional[T]) Optional[any]) Optional[any]
}

type optImpl[T any] struct {
	value *T
}

func Empty[T any]() Optional[T] {
	return &optImpl[T]{value: nil}
}

func Value[T any](value T) Optional[T] {
	return &optImpl[T]{value: &value}
}

func (o *optImpl[T]) IsPresent() bool {
	return o.value != nil
}

func (o *optImpl[T]) IsEmpty() bool {
	return o.value == nil
}

func (o *optImpl[T]) Get() (value T, found bool) {
	if o.value == nil {
		var zero T
		return zero, false
	} else {
		return *o.value, true
	}
}

func (o *optImpl[T]) OrElse(other T) T {
	if o.value == nil {
		return other
	}
	return *o.value
}

func (o *optImpl[T]) OrPanic() T {
	if o.value == nil {
		panic("empty optional")
	}
	return *o.value
}

func (o *optImpl[T]) IfPresent(op func(T)) (found bool) {
	if o.IsPresent() {
		op(*o.value)
		return true
	}
	return false
}

func (o *optImpl[T]) Map(op func(Optional[T]) Optional[any]) Optional[any] {
	return op(o)
}

func (o *optImpl[T]) String() string {
	if o.IsEmpty() {
		return "<empty>"
	} else {
		return fmt.Sprint(*o.value)
	}
}
