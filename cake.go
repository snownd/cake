package cake

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"
)

type Factory struct {
	lock *sync.RWMutex
	// cache  map[int]interface{}
	client *http.Client
}

type buildOptions struct {
	baseUrl string
	timeout time.Duration
	client  *http.Client
}

type BuildOption func(opt *buildOptions)

func WithBaseURL(url string) BuildOption {
	return func(opt *buildOptions) {
		opt.baseUrl = url
	}
}

func NewFactoryWithClient(client *http.Client) *Factory {
	f := &Factory{
		lock: &sync.RWMutex{},
		// cache:  map[int]interface{}{},
		client: client,
	}
	return f
}

func (f *Factory) Build(target interface{}, opts ...BuildOption) (interface{}, error) {
	if target == nil {
		return nil, ErrInvalidBuildTarget
	}
	bopts := &buildOptions{
		timeout: 5 * time.Second,
		client:  f.client,
	}
	for _, apply := range opts {
		apply(bopts)
	}
	t := reflect.TypeOf(target)
	switch t.Kind() {
	case reflect.Struct:
		return f.build(t, bopts)
	case reflect.Ptr:
		return f.build(t.Elem(), bopts)
	default:
		return nil, ErrInvalidBuildTarget
	}
}

func (f *Factory) build(target reflect.Type, opts *buildOptions) (interface{}, error) {
	v := reflect.New(target)
	for i := 0; i < target.NumField(); i++ {
		field := v.Elem().Field(i)
		fieldType := target.Field(i)
		if field.Kind() != reflect.Func {
			return nil, fmt.Errorf("%w with field %s is not func", ErrInvalidBuildTarget, field.Type().Name())
		}
		f, err := makeRequestFunction(field.Type(), fieldType, opts)
		if err != nil {
			return nil, err
		}
		v.Elem().Field(i).Set(f)
	}
	return v.Interface(), nil
}
