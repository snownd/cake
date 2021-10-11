package cake

import (
	"container/list"
	"context"
	"reflect"
	"strconv"
)

type cakeConfigSentinel interface {
	cakeConfigSentinel()
}

var _requestConfigType = reflect.TypeOf(RequestConfig{})
var _contextPtr *context.Context
var _contextType = reflect.TypeOf(_contextPtr).Elem()
var _errorPtr *error
var _errorType = reflect.TypeOf(_errorPtr).Elem()

type RequestConfig struct{ cakeConfigSentinel }

type TagMap = map[string]string

func NewTagMap(tag reflect.StructTag) TagMap {
	tm := make(map[string]string)
	for tag != "" {
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		value, err := strconv.Unquote(qvalue)
		if err != nil {
			// ignore invalid tag
			continue
		}
		tm[name] = value
		tag = tag[i+1:]
	}
	return tm
}

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
