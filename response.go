package cake

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

type responseBuilder = func([]byte, error) reflect.Value

var responseBuilderCache map[reflect.Type][]responseBuilder = make(map[reflect.Type][]func([]byte, error) reflect.Value)
var resMutex *sync.RWMutex = &sync.RWMutex{}

func makeJSONResponse(funcType reflect.Type, results *[]reflect.Value, data []byte, err error) {
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
	cache = make([]func([]byte, error) reflect.Value, 0)
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
	for i := 0; i < funcType.NumOut(); i++ {
		t := funcType.Out(i)
		switch t.Kind() {
		case reflect.Ptr:
			builder := func(data []byte, e error) reflect.Value {
				if e != nil {
					return reflect.Zero(t.Elem())
				}
				value := reflect.New(t.Elem())
				if e := json.Unmarshal(data, value.Interface()); e != nil {
					return reflect.Zero(t.Elem())
				}
				return value
			}
			cache = append(cache, builder)
		case reflect.Slice:
			builder := func(data []byte, e error) reflect.Value {
				if e != nil {
					return reflect.Zero(t)
				}
				value := reflect.New(t)
				if e := json.Unmarshal(data, value.Interface()); e != nil {
					return reflect.Zero(t.Elem())
				}
				return value.Elem()
			}
			cache = append(cache, builder)
		case reflect.Interface:
			if IsError(t) {
				cache = append(cache, func(data []byte, e error) reflect.Value {
					// TODO status != 20x
					if e != nil {
						return reflect.ValueOf(e)
					}
					return _nilError
				})
			} else {
				panic(fmt.Errorf("%w only accept error interface as response", ErrInvalidRequestFunction))
			}
		}
	}
	responseBuilderCache[funcType] = cache
	for _, builder := range cache {
		*results = append(*results, builder(data, err))
	}
}
