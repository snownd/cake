package cake

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

var _emptyValue reflect.Value
var _nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

type requestTemplate struct {
	ctx    context.Context
	url    string
	body   io.ReadCloser
	header http.Header
	method string
}

type argBuilder func(args []reflect.Value, req *requestTemplate) error

func newRequestTemplate(method string, opts *buildOptions) *requestTemplate {
	return &requestTemplate{
		method: method,
		url:    opts.baseUrl,
		header: make(http.Header),
	}
}

func makeRequestFunction(funcType reflect.Type, defination reflect.StructField, opts *buildOptions) (reflect.Value, error) {
	builders := make([]argBuilder, 0)
	apiDefTagMap := NewTagMap(defination.Tag)
	method, ok := apiDefTagMap[TagMethod]
	if ok {
		method = strings.ToUpper(method)
	} else {
		method = MethodGet
	}
	for i := 0; i < funcType.NumIn(); i++ {
		index := i
		argType := funcType.In(i)
		switch argType.Kind() {
		case reflect.Interface:
			if !IsContext(argType) {
				err := fmt.Errorf("%w, only accept context interface, function %v", ErrInvalidRequestFunction, funcType)
				return _emptyValue, err
			}
			builders = append(builders, func(args []reflect.Value, req *requestTemplate) error {
				// todo
				ctx := args[index].Interface().(context.Context)
				req.ctx = ctx
				return nil
			})
		case reflect.Struct:
			if IsRequestConfig(argType) {
				ab := makeArgBuilderForRequestConfigCached(argType, index, apiDefTagMap[TagURL])
				builders = append(builders, ab)
			}
		case reflect.Ptr:
			if IsRequestConfig(argType.Elem()) {
				ab := makeArgBuilderForRequestConfigCached(argType, index, apiDefTagMap[TagURL])
				builders = append(builders, ab)
			}
		default:
			err := fmt.Errorf("%w, arg types must be one of: %s,%s or %s", ErrInvalidRequestFunction, reflect.Interface, reflect.Struct, reflect.Ptr)
			return _emptyValue, err
		}
	}
	newRequest := newRequestTemplate
	if functionLevelHeaderTemplate, ok := apiDefTagMap[TagHeaders]; ok {
		headersList := strings.Split(functionLevelHeaderTemplate, ";")
		functionLevelHeaders := make(http.Header)
		for _, h := range headersList {
			kv := strings.Split(h, "=")
			// ignore invalid header format
			if len(kv) == 2 {
				functionLevelHeaders[kv[0]] = strings.Split(kv[1], ",")
			}
		}
		newRequest = func(method string, opts *buildOptions) *requestTemplate {
			r := newRequestTemplate(method, opts)
			for k, v := range functionLevelHeaders {
				r.header[k] = v
			}
			return r
		}
	}
	return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
		r := newRequest(method, opts)
		for _, builder := range builders {
			if e := builder(args, r); e != nil {
				// TODO
				panic(e)
			}
		}
		req, err := newHTTPRequest(r)
		if err != nil {
			panic(err)
		}
		res, err := opts.client.Do(req)
		if err != nil {
			// TODO
			panic(err)
		}
		results := make([]reflect.Value, 0, funcType.NumOut())
		defer res.Body.Close()
		if funcType.NumOut() == 0 {
			return results
		}
		var body []byte
		// fmt.Println(res.Header.Get(HeaderContentEncoding))
		switch res.Header.Get(HeaderContentEncoding) {
		case "gzip", "x-gzip":
			reader, e := gzip.NewReader(res.Body)
			defer reader.Close()
			if e == nil {
				body, err = io.ReadAll(reader)
			} else {
				err = e
			}
		case "deflate":
			reader := flate.NewReader(res.Body)
			defer reader.Close()
			body, err = io.ReadAll(reader)
		default:
			body, err = io.ReadAll(res.Body)
		}
		makeResponse(funcType, res.Header.Get(HeaderContentType), &results, body, err)
		return results
	}), nil
}

func newHTTPRequest(r *requestTemplate) (*http.Request, error) {
	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, r.method, r.url, r.body)
	if err != nil {
		return nil, err
	}
	if len(r.header) > 0 {
		for k, v := range r.header {
			req.Header[k] = v
		}
	}
	if r.body != nil {
		if cts := req.Header.Values(HeaderContentType); len(cts) == 0 {
			req.Header.Set(HeaderContentType, "text/plain")
		}
	}
	req.Header[HeaderAccept] = Accept
	req.Header[HeaderUserAgent] = UserAgent
	req.Header[HeaderAcceptEncoding] = AcceptEncoding

	return req, nil
}
