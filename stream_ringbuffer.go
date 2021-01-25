package go_frank

import (
	"runtime"
	"sync/atomic"
	"time"
)

type streamImpl struct {
	ringBuffer []interface{}
	ringRead   uint64
	ringWrite  uint64
	ringWAlloc uint64
	closed     int32
	pull       func(s *streamImpl) (read interface{}, closed bool)
	prev       *streamImpl
}

func (s *streamImpl) chain(pullFn func(s *streamImpl) (read interface{}, closed bool)) *streamImpl {
	return &streamImpl{prev: s, pull: pullFn}
}

// Feeds an element into the stream
func (s *streamImpl) Feed(elem interface{}) {
	head := s
	for head.prev != nil {
		head = head.prev
	}
	rbs := uint64(len(head.ringBuffer))
	for i := 0; ; i++ {
		ringRead := atomic.LoadUint64(&head.ringRead)
		ringWrite := atomic.LoadUint64(&head.ringWrite)
		ringWAlloc := atomic.LoadUint64(&head.ringWAlloc)
		ringWriteNextVal := ringWrite + 1
		if ringWrite == ringWAlloc && ringWriteNextVal-uint64(len(head.ringBuffer)) != ringRead {
			if atomic.CompareAndSwapUint64(&head.ringWAlloc, ringWAlloc, ringWriteNextVal) {
				head.ringBuffer[ringWrite%rbs] = elem
				if !atomic.CompareAndSwapUint64(&head.ringWrite, ringWrite, ringWriteNextVal) {
					panic("failed to commit allocated write in ring-buffer")
				}
				return
			}
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond) // notice nanos vs micros
	}
}

func ringPull(s *streamImpl) (read interface{}, closer bool) {
	rbs := uint64(len(s.ringBuffer))
	for i := 0; ; i++ {
		ringRead := atomic.LoadUint64(&s.ringRead)
		ringWrite := atomic.LoadUint64(&s.ringWrite)
		ringWAlloc := atomic.LoadUint64(&s.ringWAlloc)
		if ringRead != ringWrite && ringWrite == ringWAlloc {
			ringReadNextVal := ringRead + 1
			val := s.ringBuffer[ringRead%rbs]
			if atomic.CompareAndSwapUint64(&s.ringRead, ringRead, ringReadNextVal) {
				return val, false
			}
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond) // notice nanos vs micros
		if s.IsClosed() {
			return nil, true
		}
	}
}

// Closes the stream
func (s *streamImpl) Close() {
	head := s
	for head.prev != nil {
		head = head.prev
	}
	for i := 0; ; i++ {
		ringRead := atomic.LoadUint64(&head.ringRead)
		ringWrite := atomic.LoadUint64(&head.ringWrite)
		if ringRead == ringWrite {
			atomic.StoreInt32(&head.closed, 1)
			return
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Microsecond) // notice micros
	}
}

func (s *streamImpl) IsClosed() bool {
	head := s
	for head.prev != nil {
		head = head.prev
	}
	return atomic.LoadInt32(&head.closed) != 0
}

// Resets the stream position to zero, not always possible.
// hack, needs more assertions around resetting non-resettable buffers
func (s *streamImpl) Reset() uint64 {
	if s.prev == nil {
		if atomic.LoadUint64(&s.ringWrite) > uint64(len(s.ringBuffer)) {
			return atomic.LoadUint64(&s.ringRead)
		}
		atomic.StoreUint64(&s.ringRead, 0)
		if atomic.LoadInt32(&s.closed) != 0 {
			atomic.StoreInt32(&s.closed, 0)
			go s.Close()
		}
		return 0
	}
	return 0
}

func (s *streamImpl) CurrAbsPos() uint64 {
	if s.prev != nil {
		return 0
	}
	return atomic.LoadUint64(&s.ringRead)
}

func (s *streamImpl) PeekLimit() uint64 {
	if s.prev != nil {
		return 0
	}
	return atomic.LoadUint64(&s.ringWrite)
}

func (s *streamImpl) Peek(absPos uint64) interface{} {
	if s.prev != nil {
		return nil
	}
	rbs := uint64(len(s.ringBuffer))
	ringRead := atomic.LoadUint64(&s.ringRead)
	ringWrite := atomic.LoadUint64(&s.ringWrite)
	if absPos < ringRead || absPos > ringWrite {
		return nil
	}
	return s.ringBuffer[absPos%rbs]
}

// This is not mean to be used as standard API but on specific cases (i.e. to implement Binary Search). Blocks until
// an element is in the Stream, or the Stream is Closed()
func (s *streamImpl) Pull() Optional {
	if s.prev != nil {
		return EmptyOptional()
	}
	value, closed := s.pull(s)
	if closed {
		return EmptyOptional()
	}
	return OptionalOf(value)
}
