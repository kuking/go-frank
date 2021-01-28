package misc

import "reflect"

func Reflected(val interface{}) reflect.Value {
	if val == nil {
		return reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()) //XXX: error?
	} else {
		return reflect.ValueOf(val)
	}
}
