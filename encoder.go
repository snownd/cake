package cake

import (
	"bytes"
	"fmt"
	"io"
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
	return len(data), bytes.NewBuffer(data), nil
}

type DefaultTextEncoder struct {
}

func (e *DefaultTextEncoder) ContentType() string {
	return ContentTypeJson
}

func (e *DefaultTextEncoder) EncodeBody(body interface{}) (int, io.Reader, error) {
	v, ok := body.(string)
	if ok {
		return len(v), bytes.NewBufferString(v), nil
	}
	s, ok := body.(fmt.Stringer)
	if !ok {
		return 0, nil, fmt.Errorf("%w expected to be a fmt.Stringer", ErrInvalidBody)
	}
	v = s.String()
	return len(v), bytes.NewBufferString(v), nil
}
