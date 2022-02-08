package cake

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

var jsonEncoder = &DefaultJSONEncoder{}
var textEncoder = &DefaultTextEncoder{}

type DefaultJSONEncoder struct {
}

func (e *DefaultJSONEncoder) ContentType() string {
	return ContentTypeJson
}

func (e *DefaultJSONEncoder) EncodeBody(body interface{}) (int, io.Reader, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}
	return len(data), bytes.NewReader(data), nil
}

type DefaultTextEncoder struct {
}

func (e *DefaultTextEncoder) ContentType() string {
	return ContentTypeJson
}

func (e *DefaultTextEncoder) EncodeBody(body interface{}) (int, io.Reader, error) {
	v, ok := body.(string)
	if ok {
		return len(v), strings.NewReader(v), nil
	}
	s, ok := body.(fmt.Stringer)
	if !ok {
		return 0, nil, fmt.Errorf("%w expected to be a fmt.Stringer", ErrInvalidBody)
	}
	v = s.String()
	return len(v), strings.NewReader(v), nil
}
