package cake

import (
	"errors"
	"net/http"
)

type RequestMethod = string

var UserAgent = []string{"cake/" + Version}
var Accept = []string{"application/json", "text/plain"}
var AcceptEncoding = []string{"gzip", "deflate"}

const (
	HeaderAccept          = "Accept"
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentType     = "Content-Type"
	HeaderContentEncoding = "Content-Encoding"
	HeaderUserAgent       = "User-Agent"
)

const (
	MethodGet     RequestMethod = http.MethodGet
	MethodPost    RequestMethod = http.MethodPost
	MethodPut     RequestMethod = http.MethodPut
	MethodDelete  RequestMethod = http.MethodDelete
	MethodHead    RequestMethod = http.MethodHead
	MethodOptions RequestMethod = http.MethodOptions
	MethodPatch   RequestMethod = http.MethodPatch
	MethodTrace   RequestMethod = http.MethodTrace
)

const (
	APIFuncArgTagParam   = "param"
	APIFuncArgTagHeader  = "header"
	APIFuncArgTagHeaders = "headers"
	APIFuncArgTagBody    = "body"
	APIFuncArgTagQuery   = "query"
	// todo
	// APIFuncArgTagQueryString = "queryString"
)

const (
	TagMethod = "method"
	TagURL    = "url"
	TagHeader = "header"
)

var ErrInvalidBuildTarget = errors.New("cake: invalid build target")

var ErrInvalidRequestFunction = errors.New("cake: invalid request function")
