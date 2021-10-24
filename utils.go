package cake

import (
	"container/list"
	"context"
	"reflect"
)

var _requestConfigType = reflect.TypeOf(RequestConfig{})
var _contextPtr *context.Context
var _contextType = reflect.TypeOf(_contextPtr).Elem()
var _errorPtr *error
var _errorType = reflect.TypeOf(_errorPtr).Elem()

func IsRequestConfig(o interface{}) bool {
	return embedsType(o, _requestConfigType)
}

func IsContext(o interface{}) bool {
	t, ok := o.(reflect.Type)
	if !ok {
		t = reflect.TypeOf(o)
	}
	return t == _contextType
}

func IsError(o interface{}) bool {
	t, ok := o.(reflect.Type)
	if !ok {
		t = reflect.TypeOf(o)
	}
	return t == _errorType
}

// embedsType Returns true if t embeds e or if any of the types embedded by t embed e.
func embedsType(i interface{}, e reflect.Type) bool {
	// TODO: this function doesn't consider e being a pointer.

	if i == nil {
		return false
	}
	t, ok := i.(reflect.Type)
	if !ok {
		t = reflect.TypeOf(i)
	}
	types := list.New()
	types.PushBack(t)
	for types.Len() > 0 {
		t := types.Remove(types.Front()).(reflect.Type)
		if t == e {
			return true
		}
		if t.Kind() != reflect.Struct {
			continue
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				types.PushBack(f.Type)
			}
		}
	}
	return false
}
