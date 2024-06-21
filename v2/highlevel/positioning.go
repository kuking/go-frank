package highlevel

import (
	"github.com/kuking/go-frank/v2/api"
)

func (s *streamImpl[T]) Skip(n int) api.Stream[T] {
	pending := n
	return s.chain(func() (read T, found bool) {
		for pending > 0 {
			read, found = s.pull()
			if !found {
				return read, found
			}
			pending--
		}
		return s.pull()
	})
}

func (s *streamImpl[T]) SkipWhile(matching func(T) bool) api.Stream[T] {
	return s.chain(func() (read T, found bool) {
		for {
			read, found = s.pull()
			if !found {
				return read, found
			}
			if !matching(read) {
				return read, found
			}
		}
	})
}

func (s *streamImpl[T]) FindFirst(matching func(T) bool) api.Stream[T] {
	foundFirst := false
	return s.chain(func() (read T, found bool) {
		if foundFirst {
			return s.pull()
		}
		for {
			read, found = s.pull()
			if !found {
				return read, found
			}
			if matching(read) {
				foundFirst = true
				return read, found
			}
		}
	})
}

func (s *streamImpl[T]) Filter(matching func(T) bool) api.Stream[T] {
	return s.chain(func() (read T, found bool) {
		for {
			read, found = s.pull()
			if !found {
				return read, found
			}
			if matching(read) {
				return read, found
			}
		}
	})
}
