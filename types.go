package cake

import (
	"io"
	"net/http"
	"reflect"
	"strconv"
)

type cakeConfigSentinel interface {
	cakeConfigSentinel()
}

type RequestConfig struct{ cakeConfigSentinel }

type BodyEncoder interface {
	ContentType() string
	EncodeBody(body interface{}) (int, io.Reader, error)
}

type TagMap = map[string]string

func NewTagMap(tag reflect.StructTag) TagMap {
	tm := make(map[string]string)
	for tag != "" {
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		value, err := strconv.Unquote(qvalue)
		if err != nil {
			// ignore invalid tag
			continue
		}
		tm[name] = value
		tag = tag[i+1:]
	}
	return tm
}

type RequestError interface {
	error
	Unwrap() error
	StatusCode() int
	Request() *http.Request
	Response() *http.Response
}

type requestError struct {
	req *http.Request
	res *http.Response
}

func (re requestError) Error() string {
	return ErrRequestFailed.Error() + " with response status:" + re.res.Status
}

func (re requestError) Unwrap() error {
	return ErrRequestFailed
}

func (re requestError) StatusCode() int {
	return re.res.StatusCode
}

func (re requestError) Request() *http.Request {
	return re.req
}

func (re requestError) Response() *http.Response {
	return re.res
}

func NewRequestError(req *http.Request, res *http.Response) RequestError {
	return &requestError{req: req, res: res}
}
