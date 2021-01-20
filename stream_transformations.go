package go_frank

import "reflect"

// --------------------------------------------------------------------------------------------------------------------
//
// Transformation methods
//
// --------------------------------------------------------------------------------------------------------------------

func (s *streamImpl) Reduce(op interface{}) Stream {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}

	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		left, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		for {
			right, closed := ns.prev.pull(ns.prev)
			if closed {
				return left, false
			}
			rLeft := reflect.ValueOf(left)
			rRight := reflect.ValueOf(right)
			fnop := reflect.ValueOf(op)
			left = fnop.Call([]reflect.Value{rLeft, rRight})[0].Interface()
		}
	}
	return &ns
}

func (s *streamImpl) ReduceNA(reducer Reducer) Stream {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		left, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		reducer.First(left)
		for {
			right, closed := ns.prev.pull(ns.prev)
			if closed {
				return reducer.Result(), false
			}
			reducer.Next(right)
		}
	}
	return &ns
}

func (s *streamImpl) Sum() Stream {
	return s.ReduceNA(&IntSumReducer{})
}

func (s *streamImpl) SumInt64() Stream {
	return s.ReduceNA(&Int64SumReducer{})
}

func (s *streamImpl) Map(op interface{}) Stream {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	fnop := reflect.ValueOf(op)
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		value, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		return fnop.Call([]reflect.Value{reflect.ValueOf(value)})[0].Interface(), false
	}
	return &ns
}

// MapInt64 requires one allocation per element (2 for the generic Map)
func (s *streamImpl) MapInt64(op func(int64) int64) Stream {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		value, closed := ns.prev.pull(ns.prev)
		if closed {
			return nil, true
		}
		return op(value.(int64)), false
	}
	return &ns
}

func (s *streamImpl) Filter(op interface{}) Stream {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	fnop := reflect.ValueOf(op)
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		var elem interface{}
		closed = false
		for !closed {
			elem, closed = ns.prev.pull(ns.prev)
			if !closed {
				filtered := fnop.Call([]reflect.Value{reflect.ValueOf(elem)})[0].Bool()
				if !filtered {
					return elem, false
				}
			}
		}
		return nil, true
	}
	return &ns
}

func (s *streamImpl) FilterNA(op func(interface{}) bool) Stream {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		closed = false
		for !closed {
			read, closed = ns.prev.pull(ns.prev)
			if !closed {
				if !op(read) {
					return
				}
			}
		}
		return
	}
	return &ns
}

// converts the stream elements to the type provided; if the coerce parameter is set to true, it will try to convert the
// elements to the expected type i.e. a int64 will be converted into a string representation in base10, if this
// conversion is disabled, or the coercion is not possible, it can be either drop (setting the parameter dropIfNotPossible)
// or panic when dropping is disable.
func (s *streamImpl) EnsureTypeEx(t reflect.Type, coerce bool, dropIfNotPossible bool) Stream {
	return nil
}

// Ensure the types trying to coerce and dropping elements that can not be converted
func (s *streamImpl) EnsureType(t reflect.Type) Stream {
	return s.EnsureTypeEx(t, true, true)
}

// Modifies the Stream element in-place, avoids non-allocating operation
func (s *streamImpl) ModifyNA(fn func(interface{})) Stream {
	ns := streamImpl{
		closed: 0,
		prev:   s,
	}
	ns.pull = func(n *streamImpl) (read interface{}, closed bool) {
		read, closed = ns.prev.pull(ns.prev)
		if !closed {
			fn(read)
		}
		return
	}
	return &ns
}
