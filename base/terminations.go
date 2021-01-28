package base

import (
	"github.com/kuking/go-frank/api"
	"math"
	"reflect"
)

// --------------------------------------------------------------------------------------------------------------------
//
// Termination methods
//
// --------------------------------------------------------------------------------------------------------------------

func (s *StreamImpl) First() api.Optional {
	return s.Pull()
}

func (s *StreamImpl) Last() api.Optional {
	read, closed := s.pull()
	if closed {
		return api.EmptyOptional()
	}
	for {
		lastRead := read
		read, closed = s.pull()
		if closed {
			return api.OptionalOf(lastRead)
		}
	}
}

func (s *StreamImpl) IsEmpty() bool {
	return s.First().IsEmpty()
}

func (s *StreamImpl) Count() int {
	c := s.CountUint64()
	if c > math.MaxInt32 {
		return -1
	}
	return int(c)
}

func (s *StreamImpl) CountUint64() (c uint64) {
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

func (s *StreamImpl) AsArray() (result []interface{}) {
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

func (s *StreamImpl) AllMatch(op interface{}) bool {
	val, closed := s.pull()
	for !closed {
		if !reflect.ValueOf(op).Call([]reflect.Value{reflect.ValueOf(val)})[0].Bool() {
			return false
		}
		val, closed = s.pull()
	}
	return true
}

func (s *StreamImpl) NoneMatch(op interface{}) bool {
	return !s.AllMatch(op)
}

func (s *StreamImpl) AtLeastOne(op interface{}) bool {
	val, closed := s.pull()
	for !closed {
		if reflect.ValueOf(op).Call([]reflect.Value{reflect.ValueOf(val)})[0].Bool() {
			return true
		}
		val, closed = s.pull()
	}
	return false
}

func (s *StreamImpl) ForEach(op interface{}) {
	val, closed := s.pull()
	if closed {
		return
	}
	for !closed {
		reflect.ValueOf(op).Call([]reflect.Value{reflect.ValueOf(val)})
		val, closed = s.pull()
	}
}
