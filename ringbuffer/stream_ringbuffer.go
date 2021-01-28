package ringbuffer

import (
	"github.com/kuking/go-frank/api"
	"runtime"
	"sync/atomic"
	"time"
)

type ringBufferProvider struct {
	ringBuffer   []interface{}
	ringRead     uint64
	ringWrite    uint64
	ringWAlloc   uint64
	closedFlag   int32
	WaitApproach api.WaitApproach
}

func NewRingBufferProvider(capacity int) *ringBufferProvider {
	return &ringBufferProvider{
		ringBuffer:   make([]interface{}, capacity),
		ringRead:     0,
		ringWrite:    0,
		ringWAlloc:   0,
		closedFlag:   0,
		WaitApproach: api.UntilClosed,
	}
}

// Feeds an element into the stream
func (r *ringBufferProvider) Feed(elem interface{}) {
	rbs := uint64(len(r.ringBuffer))
	for i := 0; ; i++ {
		ringRead := atomic.LoadUint64(&r.ringRead)
		ringWrite := atomic.LoadUint64(&r.ringWrite)
		ringWAlloc := atomic.LoadUint64(&r.ringWAlloc)
		ringWriteNextVal := ringWrite + 1
		if ringWrite == ringWAlloc && ringWriteNextVal-uint64(len(r.ringBuffer)) != ringRead {
			if atomic.CompareAndSwapUint64(&r.ringWAlloc, ringWAlloc, ringWriteNextVal) {
				r.ringBuffer[ringWrite%rbs] = elem
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

func (r *ringBufferProvider) Pull() (read interface{}, closed bool) {
	var totalNsWait int64
	rbs := uint64(len(r.ringBuffer))
	for i := 0; ; i++ {
		ringRead := atomic.LoadUint64(&r.ringRead)
		ringWrite := atomic.LoadUint64(&r.ringWrite)
		ringWAlloc := atomic.LoadUint64(&r.ringWAlloc)
		if ringRead != ringWrite && ringWrite == ringWAlloc {
			ringReadNextVal := ringRead + 1
			val := r.ringBuffer[ringRead%rbs]
			if atomic.CompareAndSwapUint64(&r.ringRead, ringRead, ringReadNextVal) {
				return val, false
			}
		}
		otherThreadWriting := ringWrite != ringWAlloc
		if ringRead == ringWAlloc && !otherThreadWriting && r.IsClosed() {
			return nil, true
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond)
		totalNsWait += int64(i)
		if r.WaitApproach == api.UntilClosed {
			// just continue
		} else if !otherThreadWriting && totalNsWait > int64(r.WaitApproach) {
			return nil, true
		}
	}
}

// Closes the stream
func (r *ringBufferProvider) Close() {
	for i := 0; ; i++ {
		ringRead := atomic.LoadUint64(&r.ringRead)
		ringWrite := atomic.LoadUint64(&r.ringWrite)
		if ringRead == ringWrite {
			atomic.StoreInt32(&r.closedFlag, 1)
			return
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Microsecond) // notice micros
	}
}

func (r *ringBufferProvider) IsClosed() bool {
	return atomic.LoadInt32(&r.closedFlag) != 0
}

// Resets the stream position to zero, not always possible.
// hack, needs more assertions around resetting non-resettable buffers
func (r *ringBufferProvider) Reset() uint64 {
	if atomic.LoadUint64(&r.ringWrite) > uint64(len(r.ringBuffer)) {
		return atomic.LoadUint64(&r.ringRead)
	}
	atomic.StoreUint64(&r.ringRead, 0)
	if atomic.LoadInt32(&r.closedFlag) != 0 {
		atomic.StoreInt32(&r.closedFlag, 0)
		go r.Close()
	}
	return 0
}

func (r *ringBufferProvider) CurrAbsPos() uint64 {
	return atomic.LoadUint64(&r.ringRead)
}

func (r *ringBufferProvider) PeekLimit() uint64 {
	return atomic.LoadUint64(&r.ringWrite)
}

func (r *ringBufferProvider) Peek(absPos uint64) interface{} {
	rbs := uint64(len(r.ringBuffer))
	ringRead := atomic.LoadUint64(&r.ringRead)
	ringWrite := atomic.LoadUint64(&r.ringWrite)
	if absPos < ringRead || absPos > ringWrite {
		return nil
	}
	return r.ringBuffer[absPos%rbs]
}

func (r *ringBufferProvider) Wait(waitApproach api.WaitApproach) {
	r.WaitApproach = waitApproach
}
