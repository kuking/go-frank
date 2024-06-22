package base

import (
	"github.com/kuking/go-frank/v1/api"
	"github.com/kuking/go-frank/v1/ringbuffer"
	"reflect"
)

// --------------------------------------------------------------------------------------------------------------------
//
// # Builders
//
// --------------------------------------------------------------------------------------------------------------------
func EmptyStream(capacity int) api.Stream {
	rb := ringbuffer.NewRingBufferProvider(capacity, NewDefaultFastSpinThenWait())
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
