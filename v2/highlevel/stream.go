package highlevel

import "github.com/kuking/go-frank/v2/api"

type streamImpl[T any] struct {
	provider api.Provider[T]
	greedy   bool
	pull     func() (read T, found bool)
}

type numericStreamImpl[N api.Numeric] struct {
	api.Stream[N]
	sum func(N, N) N
}

func NewStreamImpl[T any](provider api.Provider[T], pullFn func() (read T, found bool)) api.Stream[T] {
	return &streamImpl[T]{
		provider: provider,
		greedy:   false,
		pull:     pullFn,
	}
}

func (s *streamImpl[T]) chain(pullFn func() (read T, found bool)) *streamImpl[T] {
	return &streamImpl[T]{
		provider: s.provider,
		greedy:   true,
		pull:     pullFn,
	}
}

func (s *streamImpl[T]) Feed(elem T) {
	s.provider.Feed(elem)
}

func (s *streamImpl[T]) Close() {
	s.provider.Close()
}

func (s *streamImpl[T]) IsClosed() bool {
	return s.provider.IsClosed()
}

func (s *streamImpl[T]) Pull() api.Optional[T] {
	if elem, found := s.pull(); found {
		return api.Value(elem)
	}
	return api.Empty[T]()
}

func (s *streamImpl[T]) PullNA() (elem T, found bool) {
	return s.provider.Pull()
}

func (s *streamImpl[T]) Prev() (moved bool) {
	if s.greedy { // greedy streams can't move backwards
		return false
	}
	return s.provider.Prev()
}

func (s *streamImpl[T]) Reset() (position uint64) {
	return s.provider.Reset()
}

func (s *streamImpl[T]) WithTimeOut(waitTimeOut api.WaitTimeOut) api.Stream[T] {
	s.provider.WithWaitTimeOut(waitTimeOut)
	return s
}

func (s *streamImpl[T]) WithWaitDuty(waitDuty api.WaitDuty) api.Stream[T] {
	s.provider.WithWaitDuty(waitDuty)
	return s
}
