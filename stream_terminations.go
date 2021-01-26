package go_frank

import (
	"math"
	"reflect"
)

// --------------------------------------------------------------------------------------------------------------------
//
// Termination methods
//
// --------------------------------------------------------------------------------------------------------------------

func (s *streamImpl) First() Optional {
	return s.Pull()
}

func (s *streamImpl) Last() Optional {
	read, closed := s.pull()
	if closed {
		return EmptyOptional()
	}
	for {
		lastRead := read
		read, closed = s.pull()
		if closed {
			return OptionalOf(lastRead)
		}
	}
}

func (s *streamImpl) IsEmpty() bool {
	return s.First().IsEmpty()
}

func (s *streamImpl) Count() int {
	c := s.CountUint64()
	if c > math.MaxInt32 {
		return -1
	}
	return int(c)
}

func (s *streamImpl) CountUint64() (c uint64) {
	c = 0
	closed := false
	for !closed {
		_, closed = s.pull()
		if !closed {
			c++
		}
	}
	return
}

func (s *streamImpl) AsArray() (result []interface{}) {
	result = make([]interface{}, 0)
	var read interface{}
	closed := false
	for !closed {
		read, closed = s.pull()
		if !closed {
			result = append(result, read)
		}
	}
	return result
}

func (s *streamImpl) AllMatch(op interface{}) bool {
	val, closed := s.pull()
	for !closed {
		if !reflect.ValueOf(op).Call([]reflect.Value{reflect.ValueOf(val)})[0].Bool() {
			return false
		}
		val, closed = s.pull()
	}
	return true
}

func (s *streamImpl) NoneMatch(op interface{}) bool {
	return !s.AllMatch(op)
}

func (s *streamImpl) AtLeastOne(op interface{}) bool {
	val, closed := s.pull()
	for !closed {
		if reflect.ValueOf(op).Call([]reflect.Value{reflect.ValueOf(val)})[0].Bool() {
			return true
		}
		val, closed = s.pull()
	}
	return false
}

func (s *streamImpl) ForEach(op interface{}) {
	val, closed := s.pull()
	if closed {
		return
	}
	for !closed {
		reflect.ValueOf(op).Call([]reflect.Value{reflect.ValueOf(val)})
		val, closed = s.pull()
	}
}
