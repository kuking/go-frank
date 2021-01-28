package base

import (
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/ringbuffer"
	"reflect"
)

// --------------------------------------------------------------------------------------------------------------------
//
// Builders
//
// --------------------------------------------------------------------------------------------------------------------

// Creates am empty api.Stream with the required capacity in its ring buffer; the api.Stream is not cosed. If used directly with
// a termination function, it will block waiting for the Closing signal. This constructor is meant to be used in a
// multithreading consumer/producer scenarios, not for simple consumers i.e. e. an array of elements (use Arrayapi.Stream)
// or for creating a api.Stream with a generator function (see api.StreamGenerator). Default blocking approach is UntilClosed.
func EmptyStream(capacity int) api.Stream {
	rb := ringbuffer.NewRingBufferProvider(capacity)
	//return StreamImpl{
	//	provider: rb,
	//	pull:     rb.Pull,
	//}
	return NewStreamImpl(rb, rb.Pull)
}

func ArrayStream(elems interface{}) api.Stream {
	var s api.Stream
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

func StreamGenerator(generator func() api.Optional) api.Stream {
	s := EmptyStream(1024)
	go StreamGeneratorFeeder(s, generator)
	return s
}

func StreamGeneratorFeeder(s api.Stream, generator func() api.Optional) {
	opt := generator()
	for opt.IsPresent() {
		s.Feed(opt.Get())
		opt = generator()
	}
	s.Close()
}
