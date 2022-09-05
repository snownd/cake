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

type responseValueBuilder = func(io.Reader, error) (reflect.Value, error)

type responseBuilder struct {
	contentType string
	builders    []responseValueBuilder
}

var responseBuilderCache map[reflect.Type]responseBuilder = make(map[reflect.Type]responseBuilder)
var resMutex *sync.RWMutex = &sync.RWMutex{}

func makeResponse(funcType reflect.Type, contentType string, results *[]reflect.Value, data io.Reader, err error) {
	resMutex.RLock()
	cache, ok := responseBuilderCache[funcType]
	resMutex.RUnlock()
	// data, err := io.ReadAll(body)
	if ok && cache.contentType == contentType {
		e := err
		for _, builder := range cache.builders {
			var v reflect.Value
			v, e = builder(data, e)
			*results = append(*results, v)
		}
		return
	}
	cache = responseBuilder{
		contentType: contentType,
	}
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
				builder := func(r io.Reader, e error) (reflect.Value, error) {
					if e != nil {
						return reflect.Zero(t), e
					}
					value := reflect.New(t.Elem())
					if e = json.NewDecoder(r).Decode(value.Interface()); e != nil {
						return reflect.Zero(t), e
					}
					return value, nil
				}
				cache.builders = append(cache.builders, builder)
			} else {
				cache.builders = append(cache.builders, func(r io.Reader, e error) (reflect.Value, error) {
					if e != nil {
						return reflect.Zero(t), e
					}
					rawRes, _ := io.ReadAll(r)
					err = fmt.Errorf("%w with type= %s responses data: %s", ErrUnexpectedResponseContentType, contentType, rawRes)
					return reflect.Zero(t), err
				})
			}
		case reflect.Slice:
			if contentType == ContentTypeJson {
				builder := func(r io.Reader, e error) (reflect.Value, error) {
					if e != nil {
						return reflect.Zero(t), e
					}
					value := reflect.New(t)
					decoder := json.NewDecoder(r)
					if e = decoder.Decode(value.Interface()); e != nil {
						return reflect.Zero(t), e
					}
					return value.Elem(), nil
				}
				cache.builders = append(cache.builders, builder)
			} else {
				cache.builders = append(cache.builders, func(r io.Reader, e error) (reflect.Value, error) {
					if e != nil {
						return reflect.Zero(t), e
					}
					return reflect.Zero(t), fmt.Errorf("%w with type= %s", ErrUnexpectedResponseContentType, contentType)
				})
			}
		case reflect.String:
			cache.builders = append(cache.builders, func(r io.Reader, e error) (reflect.Value, error) {
				if e != nil {
					return reflect.Zero(t), e
				}
				b, e := io.ReadAll(r)
				if e != nil {
					return reflect.Zero(t), e
				}
				return reflect.ValueOf(string(b)), nil
			})
		case reflect.Interface:
			if IsError(t) {
				if i != numOut-1 {
					panic(fmt.Errorf("%w error should be last function result", ErrInvalidRequestFunction))
				}
				cache.builders = append(cache.builders, func(r io.Reader, e error) (reflect.Value, error) {
					if e != nil {
						return reflect.ValueOf(e), nil
					}
					return _nilError, nil
				})
			} else {
				panic(fmt.Errorf("%w only accept error interface as response", ErrInvalidRequestFunction))
			}
		default:
			cache.builders = append(cache.builders, func(r io.Reader, e error) (reflect.Value, error) {
				return reflect.Zero(t), fmt.Errorf("%w with type= %s", ErrUnexpectedResponseContentType, contentType)
			})
		}
	}
	responseBuilderCache[funcType] = cache
	e := err
	for _, builder := range cache.builders {
		var v reflect.Value
		v, e = builder(data, e)
		*results = append(*results, v)
	}
}
