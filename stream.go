package go_frank

import (
	"math"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

var noElementMarker interface{} = "nope"

type Stream interface {
	// Feeds an element to the frank stream, blocking until space is available. In order to have space, a stream
	// terminator has to be pulling the stream.
	Feed(elem interface{})
	Close()

	// Transformations
	Map(op interface{}) *Stream
	Reduce(op interface{}) *Stream
	Filter(op interface{}) *Stream
	//Skip(int) *Stream
	//SkipRight(int) *Stream
	//DropWhile(op interface{}) *Stream
	//Find(interface{}) *Stream // can be a value or a function
	//FlatMap() *Stream
	// Reverse?
	//Sum() *Stream
	//Sort

	// Status
	IsClosed() bool

	// Terminating methods
	First() (interface{}, bool)
	Last() (interface{}, bool)
	AsArray() []interface{}
	Count() int  //LENGTH?
	CountUint64() uint64  //LENGTH?
	//AllMatch(interface{}) bool
	//NoneMatch(interface{}) bool
	//AnyMatch(interface{}) bool
	//IsEmpty() bool
	//IndexOf(interface()) int // can be a value or a function
	//Distinct??
	//EndsWith()
}

type streamImpl struct {
	elem   *interface{}
	closed int32
	pull   func(s *streamImpl) (read interface{}, closed bool)
	prev   *streamImpl
}

// Creates a EmptyStream empty stream
func EmptyStream() (stream *streamImpl) {
	stream = &streamImpl{
		elem:   &noElementMarker,
		closed: 0,
		pull:   headPull,
		prev:   nil,
	}
	return
}

func ArrayStream(elems []interface{}) (stream *streamImpl) {
	st := EmptyStream()
	idx := 0
	count := len(elems)
	st.pull = func(n *streamImpl) (read interface{}, closed bool) {
		if idx < count {
			idx++
			return elems[idx-1], false
		}
		return 0, true
	}
	return st
}

func (s *streamImpl) arrayFeederCloser(elems []interface{}) {
	for _, elem := range elems {
		s.Feed(elem)
	}
	s.Close()
}

func (s *streamImpl) Feed(elem interface{}) {
	for i := 0; ; i++ {
		var unsafeP = (*unsafe.Pointer)(unsafe.Pointer(&s.elem))
		if atomic.CompareAndSwapPointer(unsafeP, unsafe.Pointer(&noElementMarker), unsafe.Pointer(&elem)) {
			return
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Nanosecond) // notice nanos vs micros
	}
}

func headPull(s *streamImpl) (read interface{}, closed bool) {
	for i := 0; ; i++ {
		var unsafeP = (*unsafe.Pointer)(unsafe.Pointer(&s.elem))
		old := atomic.SwapPointer(unsafeP, unsafe.Pointer(&noElementMarker))
		var r = (*interface{})(old)
		if r != &noElementMarker { /// does equals string? FIX XXX
			return *r, false
		}
		runtime.Gosched()
		time.Sleep(time.Duration(i) * time.Microsecond) // notice micros vs nanos
		if s.IsClosed() {
			return nil, true
		}
	}
}

func (s *streamImpl) Close() {
	for i := 0; ; i++ {
		var unsafeP = (*unsafe.Pointer)(unsafe.Pointer(&s.elem))
		var loaded = atomic.LoadPointer(unsafeP)
		var r = (*interface{})(loaded)
		if r == &noElementMarker {
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

// --------------------------------------------------------------------------------------------------------------------
//
// Transformation methods
//
// --------------------------------------------------------------------------------------------------------------------

func (s *streamImpl) Reduce(op interface{}) *streamImpl {
	ns := streamImpl{
		elem:   nil,
		closed: 0,
		prev:   s,
	}
	fnop := reflect.ValueOf(op)
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
			left = fnop.Call([]reflect.Value{rLeft, rRight})[0].Interface()
		}
	}
	return &ns
}

func (s *streamImpl) Sum() *streamImpl {
	return s.Reduce(func(l, r int) int { return l + r })
}

func (s *streamImpl) Map(op interface{}) *streamImpl {
	ns := streamImpl{
		elem:   nil,
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

func (s *streamImpl) Filter(op interface{}) *streamImpl {
	ns := streamImpl{
		elem:   nil,
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

// --------------------------------------------------------------------------------------------------------------------
//
// Terminating methods
//
// --------------------------------------------------------------------------------------------------------------------

func (s *streamImpl) First() (first interface{}, ok bool) {
	read, closed := s.pull(s)
	if closed {
		return nil, false
	}
	return read, true
}

func (s *streamImpl) Last() (last interface{}, ok bool) {
	read, closed := s.pull(s)
	if closed {
		return nil, false
	}
	for {
		lastRead := read
		read, closed = s.pull(s)
		if closed {
			return lastRead, true
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
