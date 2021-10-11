package cake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type argBuilder func(args []reflect.Value, req *http.Request) error

var _emptyValue reflect.Value
var _nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

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
			builders = append(builders, func(args []reflect.Value, req *http.Request) error {
				// todo
				//	ctx := args[index].Interface().(context.Context)
				return nil
			})
		case reflect.Struct:
			if IsRequestConfig(argType) {
				ab := makeArgBuilderForRequestConfig(argType, index, apiDefTagMap[TagURL])
				builders = append(builders, ab)
			}
		case reflect.Ptr:
			if IsRequestConfig(argType.Elem()) {
				ab := makeArgBuilderForRequestConfig(argType, index, apiDefTagMap[TagURL])
				builders = append(builders, ab)
			}
		default:
			err := fmt.Errorf("%w, arg types must be one of: %s,%s or %s", ErrInvalidRequestFunction, reflect.Interface, reflect.Struct, reflect.Ptr)
			return _emptyValue, err
		}
	}
	return reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
		req, err := newBaseRequest(opts, method)
		if err != nil {
			panic(err)
		}
		for _, builder := range builders {
			if e := builder(args, req); e != nil {
				// TODO
				panic(err)
			}
		}
		res, err := opts.client.Do(req)
		if err != nil {
			// TODO
			panic(err)
		}
		results := make([]reflect.Value, 0, funcType.NumOut())
		defer res.Body.Close()
		if res.ContentLength == 0 || funcType.NumOut() == 0 {
			return results
		}
		contentType := res.Header.Get("Content-Type")
		if contentType == "application/json" {
			makeJSONResponse(funcType, &results, res.Body)
		}
		return results
	}), nil
}

func makeArgBuilderForRequestConfig(t reflect.Type, index int, url string) argBuilder {
	urlLayers := strings.Split(url, "/")
	urlParams := make(map[string]int)
	for i, l := range urlLayers {
		if strings.HasPrefix(l, ":") {
			urlParams[l[1:]] = i
		}
	}
	isPtr := t.Kind() == reflect.Ptr
	return func(args []reflect.Value, req *http.Request) error {
		layers := make([]string, len(urlLayers))
		copy(layers, urlLayers)
		config := args[index]

		if isPtr {
			config = config.Elem()
		}
		querys := make([]string, 0)
		for i := 0; i < config.NumField(); i++ {
			field := config.Field(i)
			fieldType := t.Elem().Field(i)
			tagmap := NewTagMap(fieldType.Tag)
			for tagName, tagValue := range tagmap {
				switch tagName {
				case APIFuncArgTagParam:
					index, ok := urlParams[tagValue]
					if ok {
						layers[index] = field.String()
					}
				case APIFuncArgTagHeader:
					key := tagValue
					if tagValue == "-" {
						key = fieldType.Name
					}
					req.Header.Set(key, field.String())
				case APIFuncArgTagHeaders:
					kind := fieldType.Type.Kind()
					headers := field
					if kind == reflect.Ptr {
						headers = field.Elem()
					}
					if headers.Kind() == reflect.Struct {
						for h := 0; h < headers.NumField(); h++ {
							header := headers.Field(i)
							key, ok := fieldType.Type.Field(i).Tag.Lookup(APIFuncArgTagHeader)
							if !ok {
								key = fieldType.Name
							}
							req.Header.Set(key, header.String())
						}
					} else if headers.Kind() == reflect.Map {
						if fieldType.Type.Key().Kind() != reflect.String || fieldType.Type.Elem().Kind() != reflect.String {
							return fmt.Errorf("%w with tag headers only support map[string]string", ErrInvalidRequestFunction)
						}
						iter := headers.MapRange()
						for iter.Next() {
							req.Header.Set(iter.Key().String(), iter.Value().String())
						}
					}
				case APIFuncArgTagBody:
					kind := fieldType.Type.Kind()
					if kind == reflect.Struct ||
						(kind == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct) ||
						kind == reflect.Map ||
						kind == reflect.Slice ||
						kind == reflect.Array {
						body, err := json.Marshal(field.Interface())
						if err != nil {
							return err
						}
						req.Body = io.NopCloser(bytes.NewBuffer(body))
					}
				case APIFuncArgTagQuery:
					key := tagValue
					kind := fieldType.Type.Kind()
					switch kind {
					case reflect.String:
						querys = append(querys, key+"="+field.String())
					case reflect.Bool:
						querys = append(querys, key+"="+strconv.FormatBool(field.Bool()))
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						querys = append(querys, key+"="+strconv.FormatInt(field.Int(), 10))
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						querys = append(querys, key+"="+strconv.FormatUint(field.Uint(), 10))
					case reflect.Float32, reflect.Float64:
						querys = append(querys, key+"="+strconv.FormatFloat(field.Float(), 'f', -1, 64))
					}
				}
			}
		}
		if len(urlParams) > 0 {
			req.URL.Path = req.URL.Path + strings.Join(layers, "/")
		} else {
			req.URL.Path = req.URL.Path + url
		}
		if len(querys) > 0 {
			req.URL.RawQuery = strings.Join(querys, "&")
		}

		return nil
	}
}

func newBaseRequest(opts *buildOptions, method string) (*http.Request, error) {
	req, err := http.NewRequest(method, opts.baseUrl, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}
