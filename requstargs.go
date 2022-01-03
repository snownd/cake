package cake

import (
	"fmt"
	urlUtils "net/url"
	"reflect"
	"strconv"
	"strings"
)

type requestConfigFieldBuilder func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error

func makeArgBuilderForRequestConfigCached(t reflect.Type, index int, url string, opts *buildOptions) argBuilder {
	urlLayers := strings.Split(url, "/")
	urlParams := make(map[string]int)
	for i, l := range urlLayers {
		if strings.HasPrefix(l, ":") {
			urlParams[l[1:]] = i
		}
	}
	isPtr := t.Kind() == reflect.Ptr
	if isPtr {
		t = t.Elem()
	}
	numField := t.NumField()
	builders := make([]requestConfigFieldBuilder, numField)
	for i := 0; i < numField; i++ {
		fieldType := t.Field(i)
		tagmap := NewTagMap(fieldType.Tag)
		for tagName, tagValue := range tagmap {
			switch tagName {
			case APIFuncArgTagForm:
				kind := fieldType.Type.Kind()
				ct := ContentTypeForm
				if tagValue != "-" && tagValue != "" {
					ct = tagValue
				}
				builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
					form := field
					if kind == reflect.Ptr {
						form = field.Elem()
					}
					data := make(urlUtils.Values)
					req.header[HeaderContentType] = []string{ct}
					if form.Kind() == reflect.Struct {
						for h := 0; h < form.NumField(); h++ {
							row := form.Field(h)
							key, ok := form.Type().Field(h).Tag.Lookup(APIFuncArgTagForm)
							if !ok {
								key = row.Type().Name()
							}
							switch form.Type().Field(h).Type.Kind() {
							case reflect.String:
								data.Set(key, row.String())
							case reflect.Bool:
								data.Set(key, strconv.FormatBool(row.Bool()))
							case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
								data.Set(key, strconv.FormatInt(row.Int(), 10))
							case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
								data.Set(key, strconv.FormatUint(row.Uint(), 10))
							case reflect.Float32, reflect.Float64:
								data.Set(key, strconv.FormatFloat(row.Float(), 'f', -1, 64))
							}
						}
					} else if form.Kind() == reflect.Map {
						if fieldType.Type.Key().Kind() != reflect.String || fieldType.Type.Elem().Kind() != reflect.String {
							return fmt.Errorf("%w with tag form only support map[string]string", ErrInvalidRequestFunction)
						}
						iter := form.MapRange()
						for iter.Next() {
							data.Set(iter.Key().String(), iter.Value().String())
						}
					}
					req.body = strings.NewReader(data.Encode())
					return nil
				}
			case APIFuncArgTagParam:
				l, ok := urlParams[tagValue]
				if ok {
					builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
						layers[l] = field.String()
						return nil
					}
				}
			case APIFuncArgTagHeader:
				key := tagValue
				if tagValue == "-" {
					key = fieldType.Name
				}
				builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
					req.header.Set(key, field.String())
					return nil
				}
			case APIFuncArgTagHeaders:
				kind := fieldType.Type.Kind()
				builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
					headers := field
					if kind == reflect.Ptr {
						headers = field.Elem()
					}
					if headers.Kind() == reflect.Struct {
						for h := 0; h < headers.NumField(); h++ {
							header := headers.Field(h)
							key, ok := headers.Type().Field(h).Tag.Lookup(APIFuncArgTagHeader)
							if !ok {
								key = header.Type().Name()
							}
							req.header.Set(key, header.String())
						}
					} else if headers.Kind() == reflect.Map {
						if fieldType.Type.Key().Kind() != reflect.String || fieldType.Type.Elem().Kind() != reflect.String {
							return fmt.Errorf("%w with tag headers only support map[string]string", ErrInvalidRequestFunction)
						}
						iter := headers.MapRange()
						for iter.Next() {
							req.header.Set(iter.Key().String(), iter.Value().String())
						}
					}
					return nil
				}
			case APIFuncArgTagBody:
				kind := fieldType.Type.Kind()
				ct := tagValue
				if ct == "" {
					ct = opts.contentType
				}
				if kind == reflect.Struct ||
					(kind == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct) ||
					kind == reflect.Map ||
					kind == reflect.Slice ||
					kind == reflect.Array ||
					kind == reflect.String {
					builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
						encoder, ok := opts.encoders[ct]
						if !ok {
							encoder = textEncoder
						}
						l, body, err := encoder.EncodeBody(field.Interface())
						if err != nil {
							return err
						}
						req.body = body
						req.header.Set(HeaderContentType, ct)
						req.header.Set(HeaderContentLength, strconv.Itoa(l))
						return nil
					}
				}
			case APIFuncArgTagQuery:
				key := tagValue
				kind := fieldType.Type.Kind()
				switch kind {
				case reflect.String:
					builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
						*querys = append(*querys, key+"="+urlUtils.QueryEscape(field.String()))
						return nil
					}
				case reflect.Bool:
					builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
						*querys = append(*querys, key+"="+strconv.FormatBool(field.Bool()))
						return nil
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
						*querys = append(*querys, key+"="+strconv.FormatInt(field.Int(), 10))
						return nil
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
						*querys = append(*querys, key+"="+strconv.FormatUint(field.Uint(), 10))
						return nil
					}
				case reflect.Float32, reflect.Float64:
					builders[i] = func(field reflect.Value, req *requestTemplate, layers []string, querys *[]string) error {
						*querys = append(*querys, key+"="+strconv.FormatFloat(field.Float(), 'f', -1, 64))
						return nil
					}
				}
			}
		}
	}

	return func(args []reflect.Value, req *requestTemplate) error {
		layers := make([]string, len(urlLayers))
		copy(layers, urlLayers)
		config := args[index]
		if isPtr {
			config = config.Elem()
		}
		querys := make([]string, 0)
		for i := 0; i < numField; i++ {
			field := config.Field(i)
			builder := builders[i]
			if builder == nil {
				continue
			}
			if err := builder(field, req, layers, &querys); err != nil {
				return err
			}
		}
		if len(urlParams) > 0 {
			req.url = req.url + strings.Join(layers, "/")
		} else {
			req.url = req.url + url
		}
		if len(querys) > 0 {
			req.url = req.url + "?" + strings.Join(querys, "&")
		}

		return nil
	}
}
