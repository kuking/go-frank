package base

import (
	"github.com/kuking/go-frank/api"
	"github.com/kuking/go-frank/misc"
	"reflect"
)

// --------------------------------------------------------------------------------------------------------------------
//
// Transformation methods
//
// --------------------------------------------------------------------------------------------------------------------

func (s *StreamImpl) Reduce(op interface{}) api.Stream {
	return s.chain(func() (read interface{}, closed bool) {
		left, closed := s.pull()
		if closed {
			return nil, true
		}
		for {
			right, closed := s.pull()
			if closed {
				return left, false
			}
			fnop := reflect.ValueOf(op)
			left = fnop.Call([]reflect.Value{misc.Reflected(left), misc.Reflected(right)})[0].Interface()
		}
	})
}

func (s *StreamImpl) ReduceNA(reducer api.Reducer) api.Stream {
	return s.chain(func() (read interface{}, closed bool) {
		left, closed := s.pull()
		if closed {
			return nil, true
		}
		reducer.First(left)
		for {
			right, closed := s.pull()
			if closed {
				return reducer.Result(), false
			}
			reducer.Next(right)
		}
	})
}

func (s *StreamImpl) Sum() api.Stream {
	return s.ReduceNA(&IntSumReducer{})
}

func (s *StreamImpl) SumInt64() api.Stream {
	return s.ReduceNA(&Int64SumReducer{})
}

func (s *StreamImpl) Map(op interface{}) api.Stream {
	fnop := reflect.ValueOf(op)
	return s.chain(func() (read interface{}, closed bool) {
		value, closed := s.pull()
		if closed {
			return nil, true
		}
		return fnop.Call([]reflect.Value{misc.Reflected(value)})[0].Interface(), false
	})
}

// MapInt64 requires one allocation per element (2 for the generic Map)
func (s *StreamImpl) MapInt64(op func(int64) int64) api.Stream {
	return s.chain(func() (read interface{}, closed bool) {
		value, closed := s.pull()
		if closed {
			return nil, true
		}
		return op(value.(int64)), false
	})
}

func (s *StreamImpl) Filter(op interface{}) api.Stream {
	fnop := reflect.ValueOf(op)
	return s.chain(func() (read interface{}, closed bool) {
		var elem interface{}
		closed = false
		for !closed {
			elem, closed = s.pull()
			if !closed {
				if !fnop.Call([]reflect.Value{misc.Reflected(elem)})[0].Bool() {
					return elem, false
				}
			}
		}
		return nil, true
	})
}

func (s *StreamImpl) FilterNA(op func(interface{}) bool) api.Stream {
	return s.chain(func() (read interface{}, closed bool) {
		closed = false
		for !closed {
			read, closed = s.pull()
			if !closed {
				if !op(read) {
					return
				}
			}
		}
		return
	})
}

// converts the stream elements to the type provided; if the coerce parameter is set to true, it will try to convert the
// elements to the expected type i.e. a int64 will be converted into a string representation in base10, if this
// conversion is disabled, or the coercion is not possible, it can be either drop (setting the parameter dropIfNotPossible)
// or panic when dropping is disable.
func (s *StreamImpl) EnsureTypeEx(t reflect.Type, coerce bool, dropIfNotPossible bool) api.Stream {
	return nil
}

// Ensure the types trying to coerce and dropping elements that can not be converted
func (s *StreamImpl) EnsureType(t reflect.Type) api.Stream {
	return s.EnsureTypeEx(t, true, true)
}

// Modifies the Stream element in-place, avoids non-allocating operation. Given the root pointer can not be changed,
// this can only be used with struct or containers i.e. maps, etc.
func (s *StreamImpl) ModifyNA(fn func(interface{})) api.Stream {
	return s.chain(func() (read interface{}, closed bool) {
		read, closed = s.pull()
		if !closed {
			fn(read)
		}
		return
	})
}
