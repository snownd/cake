package cake

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type responseBuilder = func(io.Reader, error) reflect.Value

var responseBuilderCache map[reflect.Type][]responseBuilder = make(map[reflect.Type][]responseBuilder)
var resMutex *sync.RWMutex = &sync.RWMutex{}

func makeResponse(funcType reflect.Type, contentType string, results *[]reflect.Value, data io.Reader, err error) {
	resMutex.RLock()
	cache, ok := responseBuilderCache[funcType]
	resMutex.RUnlock()
	// data, err := io.ReadAll(body)
	if ok {
		for _, builder := range cache {
			*results = append(*results, builder(data, err))
		}
		return
	}
	cache = make([]responseBuilder, 0)
	resMutex.Lock()
	defer resMutex.Unlock()
	numOut := funcType.NumOut()
	if numOut == 0 {
		responseBuilderCache[funcType] = cache
		return
	}
	if numOut > 2 {
		panic(fmt.Errorf("%w: only support 0 to 2 results", ErrInvalidRequestFunction))
	}
	for i := 0; i < numOut; i++ {
		t := funcType.Out(i)
		switch t.Kind() {
		case reflect.Ptr:
			if strings.HasPrefix(contentType, ContentTypeJson) {
				builder := func(r io.Reader, e error) reflect.Value {
					if e != nil {
						return reflect.Zero(t)
					}
					value := reflect.New(t.Elem())
					if e = json.NewDecoder(r).Decode(value.Interface()); e != nil {
						return reflect.Zero(t)
					}
					return value
				}
				cache = append(cache, builder)
			} else {
				cache = append(cache, func(r io.Reader, e error) reflect.Value {
					if e != nil {
						return reflect.Zero(t)
					}
					err = fmt.Errorf("%w with type= %s", ErrUnexpectedResponseContentType, contentType)
					return reflect.Zero(t)
				})
			}
		case reflect.Slice:
			if contentType == ContentTypeJson {
				builder := func(r io.Reader, e error) reflect.Value {
					if e != nil {
						return reflect.Zero(t)
					}
					value := reflect.New(t)
					decoder := json.NewDecoder(r)
					if e = decoder.Decode(value.Interface()); e != nil {
						return reflect.Zero(t)
					}
					return value.Elem()
				}
				cache = append(cache, builder)
			} else {
				cache = append(cache, func(r io.Reader, e error) reflect.Value {
					if e != nil {
						return reflect.Zero(t)
					}
					err = fmt.Errorf("%w with type= %s", ErrUnexpectedResponseContentType, contentType)
					return reflect.Zero(t)
				})
			}
		case reflect.String:
			cache = append(cache, func(r io.Reader, e error) reflect.Value {
				if e != nil {
					return reflect.Zero(t)
				}
				b, e := io.ReadAll(r)
				if e != nil {
					return reflect.Zero(t)
				}
				return reflect.ValueOf(string(b))
			})
		case reflect.Interface:
			if IsError(t) {
				if i != numOut-1 {
					panic(fmt.Errorf("%w error should be last function result", ErrInvalidRequestFunction))
				}
				cache = append(cache, func(r io.Reader, e error) reflect.Value {
					if e != nil {
						return reflect.ValueOf(e)
					}
					return _nilError
				})
			} else {
				panic(fmt.Errorf("%w only accept error interface as response", ErrInvalidRequestFunction))
			}
		default:
			cache = append(cache, func(r io.Reader, e error) reflect.Value {
				return reflect.Zero(t)
			})
		}
	}
	responseBuilderCache[funcType] = cache
	for _, builder := range cache {
		*results = append(*results, builder(data, err))
	}
}
