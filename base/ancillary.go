package base

// Memory efficient reducer adding all the elements of the stream into a int64 (assuming elements are int64)
// can process 175M int64/second/thread
type Int64SumReducer struct {
	v int64
}

func (r *Int64SumReducer) First(a interface{}) {
	r.v = a.(int64)
}

func (r *Int64SumReducer) Next(a interface{}) {
	r.v += a.(int64)
}

func (r *Int64SumReducer) Result() interface{} {
	return r.v
}

// Memory efficient int reducer
type IntSumReducer struct {
	v int
}

func (r *IntSumReducer) First(a interface{}) {
	r.v = a.(int)
}

func (r *IntSumReducer) Next(a interface{}) {
	r.v += a.(int)
}

func (r *IntSumReducer) Result() interface{} {
	return r.v
}
