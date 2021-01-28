package api

import (
	"reflect"
)

type Stream interface {

	// Lifecycle
	Feed(elem interface{})
	Close()
	IsClosed() bool
	Wait(waitApproach WaitApproach)

	// Positioning operations
	Reset() uint64
	Skip(int) Stream
	//SkipWhile(cmd interface{}) Stream
	//SkipWhileNA(func(interface{}) bool) Stream
	//Find(cmp interface{}) Stream // can be a value or a function
	//BinaryFind(cmp interface{}) Stream // will skip as much as possible, while possible.
	//Makes sense on network/archive streams or big buffered stream

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
	CloseFile() error
	//PruneUntil(absPos uint64)

	// stats
	//Oldest() uint64
	//Newest() uint64
	//Statistics() map[string]interface{}

	Publish(uri string)
	// Subscribing, wait approach is UntilNoMoreData
	Consume(clientName string) Stream
}

type WaitApproach int64

const (
	UntilClosed       WaitApproach = -1
	UntilNoMoreData   WaitApproach = 0
	WaitingUpto1000ns WaitApproach = 1_000
	WaitingUpto10ms   WaitApproach = 10_000_000
	WaitingUpto1s     WaitApproach = 1_000_000_000
)

// allocation free reducer
type Reducer interface {
	First(interface{})
	Next(interface{})
	Result() interface{}
}
