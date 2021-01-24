package go_frank

// skips N elements in the stream, it can block
func (s *streamImpl) Skip(n int) Stream {
	pending := n
	return s.chain(func(ns *streamImpl) (read interface{}, closed bool) {
		value, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		if pending > 0 {
			pending--
			return ns.pull(ns)
		} else {
			return value, false
		}
	})
}
