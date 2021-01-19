package go_frank

import (
	"math"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
)

type Stream interface {
	// Lifecycle

	// Feeds an element into the stream
	Feed(elem interface{})

	// Closes the stream
	Close()

	// Resets the stream position to zero, not always possible.
	Reset()

	// Transformations
	Map(op interface{}) *Stream
	MapInt(func(int) int) *Stream

	Reduce(op interface{}) *Stream

	// Non-Allocation Reducer
	ReduceNA(reducer Reducer) *Stream

	Filter(op interface{}) *Stream
	FilterNA(filter Filter) *Stream
	//Skip(int) *Stream
	//SkipRight(int) *Stream
	//DropWhile(op interface{}) *Stream
	//Find(interface{}) *Stream // can be a value or a function
	//FlatMap() *Stream
	// Reverse?

	// non-allocation sum int64
	Sum() *Stream
	SumInt64() *Stream

	//Sort

	// Status
	IsClosed() bool

	// Terminating methods
	First() Optional
	Last() Optional
	AsArray() []interface{}
	Count() int          //LENGTH?
	CountUint64() uint64 //LENGTH?
	//AllMatch(interface{}) bool
	//NoneMatch(interface{}) bool
	//AnyMatch(interface{}) bool
	//IsEmpty() bool
	//IndexOf(interface()) int // can be a value or a function
	//Distinct??
	//EndsWith()
}

// --------------------------------------------------------------------------------------------------------------------
//
// Ring buffer implementation and builders
//
// --------------------------------------------------------------------------------------------------------------------

type streamImpl struct {
	ringBuffer []interface{}
	ringRead   int32
	ringWrite  int32
	closed     int32
	pull       func(s *streamImpl) (read interface{}, closed bool)
	prev       *streamImpl
}

// Creates a EmptyStream empty stream
func EmptyStream(capacity int) (stream *streamImpl) {
	stream = &streamImpl{
		ringBuffer: make([]interface{}, capacity),
		ringRead:   0,
		ringWrite:  0,
		closed:     0,
		pull:       ringPull,
		prev:       nil,
	}
	return
}

func ArrayStream(elems []interface{}) (stream *streamImpl) {
	st := EmptyStream(len(elems) + 1)
	for _, elem := range elems {
		st.Feed(elem)
	}
	go st.Close()
	return st
}

func (s *streamImpl) Feed(elem interface{}) {
	for i := 0; ; i++ {
		ringRead := atomic.LoadInt32(&s.ringRead)
		ringWrite := atomic.LoadInt32(&s.ringWrite)
		ringWriteNext := (ringWrite + 1) % int32(len(s.ringBuffer))
		if ringWriteNext != ringRead {
			if atomic.CompareAndSwapInt32(&s.ringWrite, ringWrite, ringWriteNext) {
				s.ringBuffer[ringWrite] = elem
				return
			}
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond) // notice nanos vs micros
	}
}

func ringPull(s *streamImpl) (read interface{}, closer bool) {
	for i := 0; ; i++ {
		ringRead := atomic.LoadInt32(&s.ringRead)
		ringWrite := atomic.LoadInt32(&s.ringWrite)
		if ringRead != ringWrite {
			ringReadNext := (ringRead + 1) % int32(len(s.ringBuffer))
			if atomic.CompareAndSwapInt32(&s.ringRead, ringRead, ringReadNext) {
				return s.ringBuffer[ringRead], false
			}
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond) // notice nanos vs micros
		if s.IsClosed() {
			return nil, true
		}
	}
}

func (s *streamImpl) Close() {
	for i := 0; ; i++ {
		ringRead := atomic.LoadInt32(&s.ringRead)
		ringWrite := atomic.LoadInt32(&s.ringWrite)
		if ringRead == ringWrite {
			atomic.StoreInt32(&s.closed, 1)
			return
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Microsecond) // notice micros
	}
}

func (s *streamImpl) IsClosed() bool {
	return atomic.LoadInt32(&s.closed) != 0
}

// hack, needs more assertions around resetting non-resettable buffers
func (s *streamImpl) Reset() {
	atomic.StoreInt32(&s.ringRead, 0)
	if atomic.LoadInt32(&s.closed) != 0 {
		atomic.StoreInt32(&s.closed, 0)
		go s.Close()
	}
}

// --------------------------------------------------------------------------------------------------------------------
//
// Transformation methods
//
// --------------------------------------------------------------------------------------------------------------------

func (s *streamImpl) Reduce(op interface{}) *streamImpl {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}

	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		left, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		for {
			right, closed := ns.prev.pull(ns.prev)
			if closed {
				return left, false
			}
			rLeft := reflect.ValueOf(left)
			rRight := reflect.ValueOf(right)
			fnop := reflect.ValueOf(op)
			left = fnop.Call([]reflect.Value{rLeft, rRight})[0].Interface()
		}
	}
	return &ns
}

func (s *streamImpl) ReduceNA(reducer Reducer) *streamImpl {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		left, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		reducer.First(left)
		for {
			right, closed := ns.prev.pull(ns.prev)
			if closed {
				return reducer.Result(), false
			}
			reducer.Next(right)
		}
	}
	return &ns
}

func (s *streamImpl) Sum() *streamImpl {
	return s.ReduceNA(&IntSumReducer{})
}

func (s *streamImpl) SumInt64() *streamImpl {
	return s.ReduceNA(&Int64SumReducer{})
}

func (s *streamImpl) Map(op interface{}) *streamImpl {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	fnop := reflect.ValueOf(op)
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		value, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		return fnop.Call([]reflect.Value{reflect.ValueOf(value)})[0].Interface(), false
	}
	return &ns
}

// MapInt64 requires one allocation per element (2 for the generic Map)
func (s *streamImpl) MapInt64(op func(int64) int64) *streamImpl {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		value, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		return op(value.(int64)), false
	}
	return &ns
}

func (s *streamImpl) Filter(op interface{}) *streamImpl {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	fnop := reflect.ValueOf(op)
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		var elem interface{}
		closed = false
		for !closed {
			elem, closed = ns.prev.pull(ns.prev)
			if !closed {
				filtered := fnop.Call([]reflect.Value{reflect.ValueOf(elem)})[0].Bool()
				if !filtered {
					return elem, false
				}
			}
		}
		return nil, true
	}
	return &ns
}

func (s *streamImpl) FilterNA(filter Filter) *streamImpl {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		var elem interface{}
		closed = false
		for !closed {
			elem, closed = ns.prev.pull(ns.prev)
			if !closed {
				if !filter.Filter(elem) {
					return elem, false
				}
			}
		}
		return nil, true
	}
	return &ns
}

// --------------------------------------------------------------------------------------------------------------------
//
// Terminating methods
//
// --------------------------------------------------------------------------------------------------------------------

func (s *streamImpl) First() Optional {
	read, closed := s.pull(s)
	if closed {
		return EmptyOptional()
	}

	return OptionalOf(read)
}

func (s *streamImpl) Last() Optional {
	read, closed := s.pull(s)
	if closed {
		return EmptyOptional()
	}
	for {
		lastRead := read
		read, closed = s.pull(s)
		if closed {
			return OptionalOf(lastRead)
		}
	}
}

func (s *streamImpl) Count() int {
	c := s.CountUint64()
	if c > math.MaxInt32 {
		return -1
	}
	return int(c)
}

func (s *streamImpl) CountUint64() (c uint64) {
	c = 0
	closed := false
	for !closed {
		_, closed = s.pull(s)
		if !closed {
			c++
		}
	}
	return
}

func (s *streamImpl) AsArray() (result []interface{}) {
	result = make([]interface{}, 0)
	var read interface{}
	closed := false
	for !closed {
		read, closed = s.pull(s)
		if !closed {
			result = append(result, read)
		}
	}
	return result
}
