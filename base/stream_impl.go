package base

import "github.com/kuking/go-frank/api"

type StreamImpl struct {
	provider StreamProvider
	pull     func() (read interface{}, closed bool)
}

func NewStreamImpl(provider StreamProvider, pullFn func() (read interface{}, closed bool)) *StreamImpl {
	return &StreamImpl{
		provider: provider,
		pull:     pullFn,
	}
}

func (s *StreamImpl) chain(pullFn func() (read interface{}, closed bool)) *StreamImpl {
	return &StreamImpl{
		provider: s.provider,
		pull:     pullFn,
	}
}

func (s *StreamImpl) Feed(elem interface{}) {
	s.provider.Feed(elem)
}

func (s *StreamImpl) Close() {
	s.provider.Close()
}

func (s *StreamImpl) IsClosed() bool {
	return s.provider.IsClosed()
}

// This is not mean to be used as standard API but on specific cases (i.e. to implement Binary Search). Blocks until
// an element is in the Stream, or the Stream is Closed()
func (s *StreamImpl) Pull() api.Optional {
	elem, closed := s.pull()
	if closed {
		return api.EmptyOptional()
	} else {
		return api.OptionalOf(elem)
	}
}

func (s *StreamImpl) Reset() uint64 {
	return s.provider.Reset()
}

func (s *StreamImpl) CurrAbsPos() uint64 {
	return s.provider.CurrAbsPos()
}

func (s *StreamImpl) PeekLimit() uint64 {
	return s.provider.PeekLimit()
}

func (s *StreamImpl) Peek(absPos uint64) interface{} {
	return s.provider.Peek(absPos)
}

func (s *StreamImpl) Wait(waitApproach api.WaitApproach) api.Stream {
	s.provider.Wait(waitApproach)
	return s
}
