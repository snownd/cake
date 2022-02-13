package cake

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

const Version = "0.4.1"

type Factory struct {
	lock               *sync.RWMutex
	globalBuildOptions []BuildOption
	client             *http.Client
}

type buildOptions struct {
	baseUrl     string
	client      *http.Client
	encoders    map[string]BodyEncoder
	requestMws  []RequestMiddleware
	contentType string
}

type BuildOption func(opt *buildOptions)

func WithBaseURL(url string) BuildOption {
	return func(opt *buildOptions) {
		opt.baseUrl = url
	}
}

func WithEncoder(contentType string, encoder BodyEncoder) BuildOption {
	return func(opt *buildOptions) {
		opt.encoders[contentType] = encoder
	}
}

// WithRequestMiddleware can be called multiple times
func WithRequestMiddleware(mw RequestMiddleware) BuildOption {
	return func(opt *buildOptions) {
		opt.requestMws = append(opt.requestMws, mw)
	}
}

func WithDefaultContenType(ct string) BuildOption {
	return func(opt *buildOptions) {
		opt.contentType = ct
	}
}

func New(opts ...BuildOption) *Factory {
	client := &http.Client{
		Transport: createTransport(),
	}
	return NewFactoryWithClient(client, opts...)
}

func NewFactoryWithClient(client *http.Client, opts ...BuildOption) *Factory {
	f := &Factory{
		lock:               &sync.RWMutex{},
		globalBuildOptions: opts,
		client:             client,
	}
	return f
}

func (f *Factory) Build(target interface{}, opts ...BuildOption) (interface{}, error) {
	if target == nil {
		return nil, ErrInvalidBuildTarget
	}
	bopts := &buildOptions{
		client:      f.client,
		contentType: ContentTypeText,
		encoders: map[string]BodyEncoder{
			ContentTypeJson: jsonEncoder,
			ContentTypeText: textEncoder,
		},
	}
	for _, apply := range f.globalBuildOptions {
		apply(bopts)
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
func (f *Factory) Close() {
	f.client.CloseIdleConnections()
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
