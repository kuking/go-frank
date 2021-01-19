package go_frank

// Interfaces that facilitate no-allocation stream operations


// allocation free reducer
type Reducer interface {
	First(interface{})
	Next(interface{})
	Result() interface{}
}

// allocation free filter
type Filter interface {
	Filter(interface{}) bool
}
