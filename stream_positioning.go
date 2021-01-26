package go_frank

// skips N elements in the stream, it can block
func (s *streamImpl) Skip(n int) Stream {
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
