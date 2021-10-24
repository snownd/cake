package cake

import (
	"errors"
	"net/http"
)

const (
	ContentTypeJson = "application/json"
	contentTypeText = "text/plain"
)

var UserAgent = []string{"cake/" + Version}
var Accept = []string{ContentTypeJson, contentTypeText}
var AcceptEncoding = []string{"gzip", "deflate"}

const (
	HeaderAccept          = "Accept"
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentType     = "Content-Type"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentLength   = "Content-Length"
	HeaderUserAgent       = "User-Agent"
)

type RequestMethod = string

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
	TagMethod  = "method"
	TagURL     = "url"
	TagHeaders = "headers"
)

var ErrInvalidBuildTarget = errors.New("cake: invalid build target")

var ErrInvalidRequestFunction = errors.New("cake: invalid request function")

var ErrUnexpectedResponseContentType = errors.New("cake: unexpected resposne Content-Type")

var ErrInvalidBody = errors.New("cake: invalid body")

var ErrRequestFailed = errors.New("cake: request failed")
