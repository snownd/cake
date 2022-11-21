package cake

import (
	"compress/gzip"
	"compress/zlib"
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
	body   io.Reader
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
	url, ok := apiDefTagMap[TagURL]
	if !ok {
		return _emptyValue, fmt.Errorf("%w with no url tag", ErrInvalidRequestFunction)
	}
	if funcType.NumIn() == 0 {
		builders = append(builders, func(args []reflect.Value, req *requestTemplate) error {
			req.url = req.url + url
			return nil
		})
	} else {
		for i := 0; i < funcType.NumIn(); i++ {
			index := i
			argType := funcType.In(i)
			switch argType.Kind() {
			case reflect.Interface:
				if !IsContext(argType) {
					err := fmt.Errorf("%w, only accept context interface, function %v", ErrInvalidRequestFunction, funcType)
					return _emptyValue, err
				}
				if funcType.NumIn() == 1 {
					builders = append(builders, func(args []reflect.Value, req *requestTemplate) error {
						req.url = req.url + url
						ctx := args[index].Interface().(context.Context)
						req.ctx = ctx
						return nil
					})
				} else {
					builders = append(builders, func(args []reflect.Value, req *requestTemplate) error {
						ctx := args[index].Interface().(context.Context)
						req.ctx = ctx
						return nil
					})
				}

			case reflect.Struct:
				if IsRequestConfig(argType) {
					ab := makeArgBuilderForRequestConfigCached(argType, index, url, opts)
					builders = append(builders, ab)
				}
			case reflect.Ptr:
				if IsRequestConfig(argType.Elem()) {
					ab := makeArgBuilderForRequestConfigCached(argType, index, url, opts)
					builders = append(builders, ab)
				}
			default:
				err := fmt.Errorf("%w, arg types must be one of: %s,%s or %s", ErrInvalidRequestFunction, reflect.Interface, reflect.Struct, reflect.Ptr)
				return _emptyValue, err
			}
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
	lastIndex := len(opts.requestMws)
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
		rc := &RequestContext{
			Request:  req,
			handlers: make([]RequestHandler, len(opts.requestMws)+1),
		}
		for i, mw := range opts.requestMws {
			rc.handlers[i] = mw
		}
		results := make([]reflect.Value, 0, funcType.NumOut())
		h := newRequestRunner(opts.client)
		rc.handlers[lastIndex] = h
		err = rc.handlers[0](rc)
		if err != nil {
			makeResponse(funcType, "", &results, nil, newRequestError(err, req, rc.Response))
			return results
		}
		res := rc.Response
		defer res.Body.Close()
		if funcType.NumOut() == 0 {
			return results
		}
		if res.StatusCode/100 != 2 {
			makeResponse(funcType, "", &results, nil, NewRequestError(req, res))
			return results
		}
		var body io.Reader
		// fmt.Println(res.Header.Get(HeaderContentEncoding))
		switch res.Header.Get(HeaderContentEncoding) {
		case "gzip", "x-gzip":
			reader, e := gzip.NewReader(res.Body)
			defer reader.Close()
			if e == nil {
				body = reader
			} else {
				err = e
			}
		case "deflate":
			reader, e := zlib.NewReader(res.Body)
			defer reader.Close()
			if e == nil {
				body = reader
			} else {
				err = e
			}
		default:
			body = res.Body
		}
		makeResponse(funcType, GetContentType(res.Header), &results, body, err)
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

func newRequestRunner(client *http.Client) RequestHandler {
	return func(c *RequestContext) error {
		res, err := client.Do(c.Request)
		if err != nil {
			return err
		}
		c.Response = res
		return c.Next()
	}
}
