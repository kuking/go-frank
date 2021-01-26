package go_frank

import (
	"reflect"
)

type Stream interface {

	// Lifecycle
	Feed(elem interface{})
	Close()
	IsClosed() bool

	// Positioning operations
	Reset() uint64
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
	ModifyNA(func(interface{})) Stream
	Reduce(op interface{}) Stream
	ReduceNA(reducer Reducer) Stream
	Filter(op interface{}) Stream
	FilterNA(func(interface{}) bool) Stream
	//SkipRight(int) Stream
	//FlatMap() Stream
	//Reverse?

	Sum() Stream
	SumInt64() Stream

	EnsureTypeEx(t reflect.Type, coerce bool, dropIfNotPossible bool) Stream
	EnsureType(t reflect.Type) Stream

	JsonToMap() Stream
	MapToJson() Stream
	CSVasMap(firstRowIsHeader bool, asMap bool) Stream
	MapAsCSV(firstRowIsHeader bool) Stream

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
	//IndexOf(interface()) int // can be a value or a function
	//Distinct() []interface{}
	//EndsWith()

	Publish(uri string)
}

// multi-consumer, multi-producer persistent Stream
type PersistentStream interface {

	// Lifecycle
	Feed(elem interface{})
	Close()
	IsClosed() bool
	//PruneUntil(absPos uint64)

	// stats
	//Oldest() uint64
	//Newest() uint64
	//Statistics() map[string]interface{}

	// Subscribing
	Consume(clientName string) Stream
}

// --------------------------------------------------------------------------------------------------------------------
//
// Builders
//
// --------------------------------------------------------------------------------------------------------------------

// Creates am empty stream with the required capacity in its ring buffer; the stream is not cosed. If used directly with
// a termination function, it will block waiting for the Closing signal. This constructor is meant to be used in a
// multithreading consumer/producer scenarios, not for simple consumers i.e. e. an array of elements (use ArrayStream)
// or for creating a stream with a generator function (see StreamGenerator)
func EmptyStream(capacity int) Stream {
	rb := newRingBufferProvider(capacity)
	return &streamImpl{
		provider: rb,
		pull:     rb.Pull,
	}
}

func ArrayStream(elems interface{}) Stream {
	var s Stream
	slice := reflect.ValueOf(elems)
	if slice.Kind() == reflect.Slice {
		s = EmptyStream(slice.Len() + 1)
		for i := 0; i < slice.Len(); i++ {
			s.Feed(slice.Index(i).Interface())
		}
	} else {
		s = EmptyStream(2)
		if elems != nil {
			s.Feed(elems)
		}
	}
	go s.Close()
	return s
}

func StreamGenerator(generator func() Optional) Stream {
	s := EmptyStream(1024)
	go streamGeneratorFeeder(s, generator)
	return s
}

func streamGeneratorFeeder(s Stream, generator func() Optional) {
	opt := generator()
	for opt.isPresent() {
		s.Feed(opt.Get())
		opt = generator()
	}
	s.Close()
}

func OpenCreatePersistentStream(basePath string, partSize uint64, serialiser StreamSerialiser) (ps PersistentStream, err error) {
	ps, err = mmapStreamOpen(basePath, serialiser)
	if err == nil {
		return
	} else {
		ps, err = mmapStreamCreate(basePath, partSize, serialiser)
		return
	}
}
