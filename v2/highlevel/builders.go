package highlevel

import (
	"github.com/kuking/go-frank/v2/api"
	"github.com/kuking/go-frank/v2/dutycycle"
	"github.com/kuking/go-frank/v2/memory"
)

func InMemoryStream[T any](capacity int) api.Stream[T] {
	rb := memory.NewRingBufferProvider[T](capacity, dutycycle.NewDefaultFastSpinThenWait())
	return NewStreamImpl[T](rb, rb.Pull)
}

func AsNumericStream[T api.Numeric](s api.Stream[T]) api.NumericStream[T] {
	return &numericStreamImpl[T]{
		s,
		func(l T, r T) T { return l + r },
	}
}
