package highlevel

import (
	"github.com/kuking/go-frank/v2/api"
	"math"
)

func (s *streamImpl[T]) Count() int {
	c := s.CountUint64()
	if c > math.MaxInt32 {
		return -1
	}
	return int(c)
}

func (s *streamImpl[T]) CountUint64() (c uint64) {
	c = 0
	found := true
	for found {
		_, found = s.pull()
		if found {
			c++
		}
	}
	return
}

func (s *streamImpl[T]) AllMatch(condition func(T) bool) (allMatches bool) {
	allMatches = true
	for {
		value, found := s.pull()
		if !found {
			return
		}
		allMatches = allMatches && condition(value)
		if !allMatches {
			return
		}
	}
}

func (s *streamImpl[T]) NoneMatch(condition func(T) bool) (noneMatches bool) {
	noneMatches = true
	for {
		value, found := s.pull()
		if !found {
			return
		}
		noneMatches = noneMatches && !condition(value)
		if !noneMatches {
			return
		}
	}
}

func (s *streamImpl[T]) AtLeastOne(condition func(T) bool) (oneMatched bool) {
	for {
		value, found := s.pull()
		if !found {
			return false
		}
		if condition(value) {
			return true
		}
	}
}

func (s *streamImpl[T]) First() api.Optional[T] {
	s.Reset()
	return s.Pull()
}

func (s *streamImpl[T]) Last() api.Optional[T] {
	s.ForEach(func(t T) {})
	s.Prev()
	return s.Pull()
}

func (s *streamImpl[T]) Reduce(reducer func(left T, right T) T) (reduced T) {
	firstDone := false
	s.ForEach(func(t T) {
		if !firstDone {
			reduced = t
			firstDone = true
		} else {
			reduced = reducer(reduced, t)
		}
	})
	return
}

func (s *streamImpl[T]) ReduceId(identity T, reducer func(left T, right T) T) (reduced T) {
	reduced = identity
	s.ForEach(func(t T) {
		reduced = reducer(reduced, t)
	})
	return reduced
}

// -- NumericStream

func (n *numericStreamImpl[N]) Sum() N {
	return n.Reduce(n.sum)
}

func (n *numericStreamImpl[N]) Max() N {
	return n.Reduce(func(l, r N) N {
		if l > r {
			return l
		}
		return r
	})
}

func (n *numericStreamImpl[N]) Min() N {
	return n.Reduce(func(l, r N) N {
		if l < r {
			return l
		}
		return r
	})
}

func (n *numericStreamImpl[N]) Avg() float64 {
	var count float64
	var sum float64
	n.ForEach(func(t N) {
		count += 1
		sum += float64(t)
	})
	if count == 0 {
		return 0
	}
	return sum / count
}
