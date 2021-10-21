package cake

import (
	"context"
	"io"
	"reflect"
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

type BodyEncoder interface {
	ContentType() string
	EncodeBody(body interface{}) (io.Reader, error)
}

type TagMap = map[string]string
