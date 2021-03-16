package api

import (
	"reflect"
)

type Stream interface {

	// Lifecycle
	Feed(elem interface{})
	Close()
	IsClosed() bool
	TimeOut(waitTimeOut WaitTimeOut) Stream
	WaitDuty(duty WaitDuty) Stream

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

	EnsureTypeEx(t reflect.Kind, coerce bool, dropIfNotPossible bool) Stream
	EnsureType(t reflect.Kind) Stream

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

	Publish(uri string) error
	PublishClose(uri string) error
}

// multi-consumer, multi-producer persistent Stream
type PersistentStream interface {

	// Lifecycle
	Feed(elem interface{})
	Close()
	IsClosed() bool
	CloseFile() error
	Delete() error
	//PruneUntil(absPos uint64)

	// stats
	//Oldest() uint64
	//Newest() uint64
	//Statistics() map[string]interface{}

	Publish(uri string)
	// Subscribing, wait time-out is UntilNoMoreData
	Consume(subscriberName string) Stream
}

type WaitTimeOut int64

const (
	UntilClosed       WaitTimeOut = -1
	UntilNoMoreData   WaitTimeOut = 0
	WaitingUpto1000ns WaitTimeOut = 1_000
	WaitingUpto10ms   WaitTimeOut = 10_000_000
	WaitingUpto1s     WaitTimeOut = 1_000_000_000
)

// Strategy specifying how to wait and process for new data. Different strategies are a trade-off between latency and
// CPU usage. i.e. busy wait will have the lowest latency, at the expense of putting one CPU at 100% usage.
// Always-wait, will impose a few ms latency due to clock-checking, wait and context-switching.
// An hybrid solution it usually a good trade-off, fast-spin for a few thousand times, and then if no activity is
// detected the thread is put to sleep and the context is switched to another process/goroutine; each time waiting a
// bit longer until certain upper limit.
type WaitDuty interface {
	// callback notifying a loop during wait has occurred, if something has been processed since the last call the
	// parameter `hasProcessed` should be set to true; the call might put the processor to wait or the goroutine to
	// yield, depending on its configured behaviour.
	Loop() (waitedNs int64)
	Reset()
}

// allocation free reducer
type Reducer interface {
	First(interface{})
	Next(interface{})
	Result() interface{}
}
