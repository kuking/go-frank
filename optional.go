package go_frank

import (
	"errors"
	"reflect"
)

type Optional struct {
	value *interface{}
}

func EmptyOptional() Optional {
	return Optional{nil}
}

func OptionalOf(value interface{}) Optional {
	return Optional{&value}
}

func (o Optional) Get() interface{} {
	if o.value == nil {
		return errors.New("tried to Get an empty option")
	}
	return *o.value
}

func (o Optional) isPresent() bool {
	return o.value != nil
}

func (o Optional) isEmpty() bool {
	return o.value == nil
}

func (o Optional) ifPresent(op interface{}) {
	if o.isPresent() {
		reflect.ValueOf(op).Call([]reflect.Value{reflect.ValueOf(*o.value)})
	}
}

func (o Optional) orElse(other interface{}) interface{} {
	if o.isPresent() {
		return o.Get()
	} else {
		return other
	}
}

func (o Optional) Map(op interface{}) Optional {
	if o.isEmpty() {
		return EmptyOptional()
	}
	return OptionalOf(reflect.ValueOf(op).Call([]reflect.Value{reflect.ValueOf(*o.value)})[0].Interface())
}
