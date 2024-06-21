package api

type Stream[T any] interface {

	/*
		Lifecycle
	*/
	Feed(elem T)
	Pull() Optional[T]
	Prev() (moved bool)
	PullNA() (elem T, found bool)

	/*
		Managing State
	*/
	Close()
	IsClosed() bool
	WithTimeOut(waitTimeOut WaitTimeOut) Stream[T]
	WithWaitDuty(waitDuty WaitDuty) Stream[T]

	/*
		Positioning
	*/
	Reset() (position uint64)
	Skip(count int) Stream[T]
	SkipWhile(matching func(T) bool) Stream[T]
	FindFirst(matching func(T) bool) Stream[T]
	Filter(matching func(T) bool) Stream[T]

	/*
		Miscellaneous
	*/
	ForEach(op func(t T))

	// Transformations

	//Map(op interface{}) Stream[T]
	//FlatMap() Stream
	//Reverse?

	//JsonToMap() Stream[map[string]interface{}]
	//MapToJson() Stream[string]
	//CSVasMap(firstRowIsHeader bool, asMap bool) Stream[T]
	//MapAsCSV(firstRowIsHeader bool) Stream[T]

	/*
		Terminating Functions
	*/
	Count() int
	CountUint64() uint64
	AllMatch(condition func(T) bool) (allMatches bool)
	NoneMatch(condition func(T) bool) (noneMatches bool)
	AtLeastOne(condition func(T) bool) (oneMatched bool)
	First() Optional[T]
	Last() Optional[T]
	Reduce(reducer func(left T, right T) T) (reduced T)
	ReduceId(identity T, reducer func(left T, right T) T) (reduced T)
}

type Numeric interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

type NumericStream[N Numeric] interface {
	Stream[N]
	Sum() N
	Max() N
	Min() N
	Avg() float64
}

type MappedStream[T any, R any] interface {
	Stream[T]
	Map(mapper func(T) R)
}
