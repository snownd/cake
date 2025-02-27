package cake

import (
	"io"
	"net/http"
	"reflect"
	"strconv"
)

type RequestMiddlewareDesprate func(*http.Request) error

type RequestContext struct {
	Request  *http.Request
	Response *http.Response
	index    int
	handlers []RequestHandler
}

func (c *RequestContext) Next() error {
	c.index++
	for c.index < len(c.handlers) {
		if err := c.handlers[c.index](c); err != nil {
			return err
		}
		c.index++
	}
	return nil
}

type RequestHandler func(c *RequestContext) error

type RequestMiddleware = RequestHandler

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
	Body() []byte
}

type requestError struct {
	err  error
	req  *http.Request
	res  *http.Response
	body []byte
}

func (re requestError) Error() string {
	if re.err != nil {
		return re.err.Error()
	}
	if re.res == nil {
		return ErrRequestFailed.Error()
	}
	return ErrRequestFailed.Error() + " with response status:" + re.res.Status
}

func (re requestError) Unwrap() error {
	if re.err != nil {
		return re.err
	}
	return ErrRequestFailed
}

func (re requestError) StatusCode() int {
	if re.res == nil {
		// The HTTP 499 status code is not a standard HTTP status code, meaning it is not defined in the HTTP/1.1 specification.
		// This status code is used by the nginx web server to indicate that the client closed the connection before the server could send a response
		return 499
	}
	return re.res.StatusCode
}

func (re requestError) Request() *http.Request {
	return re.req
}

func (re requestError) Response() *http.Response {
	return re.res
}

func (re requestError) Body() []byte {
	return re.body
}

func NewRequestError(req *http.Request, res *http.Response) RequestError {
	return newRequestError(ErrRequestFailed, req, res)
}

func newRequestError(err error, req *http.Request, res *http.Response) RequestError {
	var body []byte
	if res != nil {
		body, _ = io.ReadAll(res.Body)
	}
	return &requestError{err: err, req: req, res: res, body: body}
}
