package base

import (
	"github.com/kuking/go-frank/v1/api"
)

// skips N elements in the stream, it can block
func (s *StreamImpl) Skip(n int) api.Stream {
	pending := n
	return s.chain(func() (read interface{}, closed bool) {
		for pending > 0 {
			read, closed = s.pull()
			if closed {
				return nil, closed
			}
			pending--
		}
		return s.pull()
	})
}
