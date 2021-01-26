package go_frank

type streamProvider interface {
	Feed(elem interface{})
	Close()
	IsClosed() bool
	Pull() (elem interface{}, closed bool)
	Reset() uint64
	CurrAbsPos() uint64
	PeekLimit() uint64
	Peek(absPos uint64) interface{}
}

type streamImpl struct {
	provider streamProvider
	pull     func() (read interface{}, closed bool)
}

func (s *streamImpl) chain(pullFn func() (read interface{}, closed bool)) *streamImpl {
	return &streamImpl{
		provider: s.provider,
		pull:     pullFn,
	}
}

func (s *streamImpl) Feed(elem interface{}) {
	s.provider.Feed(elem)
}

func (s *streamImpl) Close() {
	s.provider.Close()
}

func (s *streamImpl) IsClosed() bool {
	return s.provider.IsClosed()
}

// This is not mean to be used as standard API but on specific cases (i.e. to implement Binary Search). Blocks until
// an element is in the Stream, or the Stream is Closed()
func (s *streamImpl) Pull() Optional {
	elem, closed := s.pull()
	if closed {
		return EmptyOptional()
	} else {
		return OptionalOf(elem)
	}
}

func (s *streamImpl) Reset() uint64 {
	return s.provider.Reset()
}

func (s *streamImpl) CurrAbsPos() uint64 {
	return s.provider.CurrAbsPos()
}

func (s *streamImpl) PeekLimit() uint64 {
	return s.provider.PeekLimit()
}

func (s *streamImpl) Peek(absPos uint64) interface{} {
	return s.provider.Peek(absPos)
}
