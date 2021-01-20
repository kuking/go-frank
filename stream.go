package go_frank

import (
	"os"
	"reflect"
)

type Stream interface {

	// Lifecycle
	Feed(elem interface{})
	Close()
	IsClosed() bool
	Reset()

	// Positioning operations

	Skip(int) Stream

	//SkipWhile(cmd interface{}) Stream
	//SkipWhileNA(func(interface{}) bool) Stream
	//Find(cmp interface{}) Stream // can be a value or a function
	//BinaryFind(cmp interface{}) Stream // will skip as much as possible, while possible.
	//Makes sense on network/archive streams or big buffered stream

	// Positioning helpers for Stream heads, non-head streams will return 0
	CurrAbsPos() uint64
	PeekLimit() uint64
	Peek(absPos uint64) interface{}
	Pull() Optional

	// Transformations

	Map(op interface{}) Stream
	MapInt64(func(int64) int64) Stream

	Reduce(op interface{}) Stream

	// Non-Allocation Reducer
	ReduceNA(reducer Reducer) Stream

	Filter(op interface{}) Stream
	FilterNA(func(interface{}) bool) Stream
	//SkipRight(int) Stream
	//FlatMap() Stream
	// Reverse?

	// non-allocation sum int64
	Sum() Stream
	SumInt64() Stream

	EnsureTypeEx(t reflect.Type, coerce bool, dropIfNotPossible bool) Stream
	EnsureType(t reflect.Type) Stream

	//JsonToMap() Stream
	//MapToJson() Stream

	CSVasMap(firstRowIsColumnName bool, asMap bool) Stream
	MapAsCSV(firstRowHasColumnName bool) Stream

	//Sort

	// Terminating methods
	First() Optional
	Last() Optional
	IsEmpty() bool
	AsArray() []interface{}
	Count() int          //LENGTH?
	CountUint64() uint64 //LENGTH?
	AllMatch(op interface{}) bool
	NoneMatch(interface{}) bool
	AtLeastOne(interface{}) bool
	ForEach(op interface{})
	ToTextFile(f os.File)
	//IndexOf(interface()) int // can be a value or a function
	//Distinct() []interface{}
	//EndsWith()

	// Publish(uri string)

}

// --------------------------------------------------------------------------------------------------------------------
//
// Ring buffer implementation and builders
//
// --------------------------------------------------------------------------------------------------------------------

type streamImpl struct {
	ringBuffer []interface{}
	ringRead   uint64
	ringWrite  uint64
	ringWAlloc uint64
	closed     int32
	pull       func(s *streamImpl) (read interface{}, closed bool)
	prev       *streamImpl
}

// Creates am empty stream with the required capacity in its ring buffer; the stream is not cosed. If used directly with
// a termination function, it will block waiting for the Closing signal. This constructor is meant to be used in a
// multithreading consumer/producer scenarios, not for simple consumers i.e. e. an array of elements (use ArrayStream)
// or for creating a stream with a generator function (see StreamGenerator)
func EmptyStream(capacity int) Stream {
	return &streamImpl{
		ringBuffer: make([]interface{}, capacity),
		ringRead:   0,
		ringWrite:  0,
		ringWAlloc: 0,
		closed:     0,
		pull:       ringPull,
		prev:       nil,
	}
}

func ArrayStream(elems []interface{}) Stream {
	st := EmptyStream(len(elems) + 1)
	for _, elem := range elems {
		st.Feed(elem)
	}
	go st.Close()
	return st
}

func StreamGenerator(generator func() Optional) Stream {
	s := EmptyStream(1024)
	go streamGeneratorFeeder(s, generator)
	return s
}

func streamGeneratorFeeder(s Stream, generator func() Optional) {
	opt := generator()
	for ; opt.isPresent(); {
		s.Feed(opt.Get())
		opt = generator()
	}
	s.Close()
}
