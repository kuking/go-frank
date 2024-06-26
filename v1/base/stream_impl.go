package base

import (
	api2 "github.com/kuking/go-frank/v1/api"
)

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
func (s *StreamImpl) Pull() api2.Optional {
	elem, closed := s.pull()
	if closed {
		return api2.EmptyOptional()
	} else {
		return api2.OptionalOf(elem)
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

func (s *StreamImpl) TimeOut(waitTimeOut api2.WaitTimeOut) api2.Stream {
	s.provider.WaitTimeOut(waitTimeOut)
	return s
}

func (s *StreamImpl) WaitDuty(waitDuty api2.WaitDuty) api2.Stream {
	s.provider.WaitDuty(waitDuty)
	return s
}
