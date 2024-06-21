package memory

import (
	"github.com/kuking/go-frank/v2/api"
	"runtime"
	"sync/atomic"
	"time"
)

type ringBufferProvider[T any] struct {
	ringBuffer  []T
	ringRead    uint64
	ringWrite   uint64
	ringWAlloc  uint64
	closedFlag  int32
	waitTimeOut api.WaitTimeOut
	waitDuty    api.WaitDuty
	cap         uint64
	zero        T
}

func NewRingBufferProvider[T any](capacity int, waitDuty api.WaitDuty) *ringBufferProvider[T] {
	return &ringBufferProvider[T]{
		ringBuffer:  make([]T, capacity),
		ringRead:    0,
		ringWrite:   0,
		ringWAlloc:  0,
		closedFlag:  0,
		cap:         uint64(capacity),
		waitTimeOut: api.UntilClosed,
		waitDuty:    waitDuty,
	}
}

// Feeds an element into the stream
func (r *ringBufferProvider[T]) Feed(elem T) {
	for i := 0; ; i++ {
		ringRead := atomic.LoadUint64(&r.ringRead)
		ringWrite := atomic.LoadUint64(&r.ringWrite)
		ringWAlloc := atomic.LoadUint64(&r.ringWAlloc)
		ringWriteNextVal := ringWrite + 1
		if ringWrite == ringWAlloc && ringWriteNextVal-r.cap != ringRead {
			if atomic.CompareAndSwapUint64(&r.ringWAlloc, ringWAlloc, ringWriteNextVal) {
				r.ringBuffer[ringWrite%r.cap] = elem
				if !atomic.CompareAndSwapUint64(&r.ringWrite, ringWrite, ringWriteNextVal) {
					panic("failed to commit allocated write in ring-buffer")
				}
				return
			}
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond) // notice nanos vs micros
	}
}

func (r *ringBufferProvider[T]) Pull() (read T, found bool) {
	var totalNsWait int64
	r.waitDuty.Reset()
	for i := 0; ; i++ {
		ringRead := atomic.LoadUint64(&r.ringRead)
		ringWrite := atomic.LoadUint64(&r.ringWrite)
		ringWAlloc := atomic.LoadUint64(&r.ringWAlloc)
		if ringRead != ringWrite && ringWrite == ringWAlloc {
			ringReadNextVal := ringRead + 1
			val := r.ringBuffer[ringRead%r.cap]
			if atomic.CompareAndSwapUint64(&r.ringRead, ringRead, ringReadNextVal) {
				return val, true
			}
		}
		otherThreadWriting := ringWrite != ringWAlloc
		if ringRead == ringWAlloc && !otherThreadWriting && r.IsClosed() {
			return r.zero, false
		}
		totalNsWait += r.waitDuty.Loop()
		if r.waitTimeOut == api.UntilClosed {
			// just continue
		} else if !otherThreadWriting && totalNsWait > int64(r.waitTimeOut) {
			return r.zero, false
		}
	}
}

// Close the stream
func (r *ringBufferProvider[T]) Close() {
	atomic.StoreInt32(&r.closedFlag, 1)
}

func (r *ringBufferProvider[T]) IsClosed() bool {
	return atomic.LoadInt32(&r.closedFlag) != 0
}

// Reset the stream position to zero, not always possible.
func (r *ringBufferProvider[T]) Reset() uint64 {
	if atomic.LoadUint64(&r.ringWrite) > r.cap {
		return atomic.LoadUint64(&r.ringRead)
	}
	atomic.StoreUint64(&r.ringRead, 0)
	if atomic.LoadInt32(&r.closedFlag) != 0 {
		atomic.StoreInt32(&r.closedFlag, 0)
		defer r.Close()
	}
	return 0
}

func (r *ringBufferProvider[T]) Prev() (moved bool) {
	rbs := uint64(len(r.ringBuffer))
	for {
		ringRead := atomic.LoadUint64(&r.ringRead)
		ringWrite := atomic.LoadUint64(&r.ringWrite)
		if ringRead == 0 || ringWrite-rbs == ringRead {
			return false
		}
		ringReadNextVal := ringRead - 1
		if atomic.CompareAndSwapUint64(&r.ringRead, ringRead, ringReadNextVal) {
			return true
		}
		runtime.Gosched()
	}
}

func (r *ringBufferProvider[T]) WritePos() uint64 {
	return atomic.LoadUint64(&r.ringWrite)
}

func (r *ringBufferProvider[T]) ReadPos() uint64 {
	return atomic.LoadUint64(&r.ringRead)
}

func (r *ringBufferProvider[T]) PeekLimit() uint64 {
	return atomic.LoadUint64(&r.ringWrite)
}

func (r *ringBufferProvider[T]) Peek(absPos uint64) interface{} {
	rbs := uint64(len(r.ringBuffer))
	ringRead := atomic.LoadUint64(&r.ringRead)
	ringWrite := atomic.LoadUint64(&r.ringWrite)
	if absPos < ringRead || absPos > ringWrite {
		return nil
	}
	return r.ringBuffer[absPos%rbs]
}

func (r *ringBufferProvider[T]) WithWaitTimeOut(waitTimeOut api.WaitTimeOut) {
	r.waitTimeOut = waitTimeOut
}

func (r *ringBufferProvider[T]) WithWaitDuty(waitDuty api.WaitDuty) {
	r.waitDuty = waitDuty
}
