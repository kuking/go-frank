package misc

import "reflect"

func Reflected(val interface{}) reflect.Value {
	if val == nil {
		return reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()) //XXX: error?
	} else {
		return reflect.ValueOf(val)
	}
}

func AsUint32Bool(value bool) uint32 {
	if value {
		return 1
	}
	return 0
}

func Uint32Bool(value uint32) bool {
	return value > 0
}
