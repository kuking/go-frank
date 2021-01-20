package go_frank

// skips N elements in the stream, it can block
func (s *streamImpl) Skip(n int) Stream {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	pending := n
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		value, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		if pending > 0 {
			pending--
			return ns.pull(n)
		} else {
			return value, false
		}

	}
	return &ns
}
