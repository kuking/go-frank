package v2

import (
	"github.com/kuking/go-frank/v2/api"
	"github.com/kuking/go-frank/v2/highlevel"
)

func InMemoryStream[T any](capacity int) api.Stream[T] {
	return highlevel.InMemoryStream[T](capacity)
}

func AsNumericStream[T api.Numeric](s api.Stream[T]) api.NumericStream[T] {
	return highlevel.AsNumericStream(s)
}
