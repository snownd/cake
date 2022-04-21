package cake

import (
	"errors"
	"net/http"
)

const (
	ContentTypeJson = "application/json"
	ContentTypeText = "text/plain"
	ContentTypeForm = "application/x-www-form-urlencoded"
)

type ContentTypeAlias string

func (c ContentTypeAlias) GetContentType() string {
	switch c {
	case ContentTypeAliasJson:
		return ContentTypeJson
	case ContentTypeAliasText:
		return ContentTypeText
	case ContentTypeAliasForm:
		return ContentTypeForm
	default:
		return string(c)
	}
}

const (
	// alias for ContentTypeJson
	ContentTypeAliasJson ContentTypeAlias = "json"
	// alias for ContentTypeText
	ContentTypeAliasText = "text"
	// alias for ContentTypeForm
	ContentTypeAliasForm = "form"
)

var UserAgent = []string{"cake/" + Version}
var Accept = []string{ContentTypeJson, ContentTypeText}
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
	// application/x-www-form-urlencoded
	APIFuncArgTagForm = "form"
)

const (
	TagMethod  = "method"
	TagURL     = "url"
	TagHeaders = "headers"
)

var ErrInvalidBuildTarget = errors.New("cake: invalid build target")

var ErrInvalidRequestFunction = errors.New("cake: invalid request function")

var ErrUnexpectedResponseContentType = errors.New("cake: unexpected response Content-Type")

var ErrInvalidBody = errors.New("cake: invalid body")

var ErrRequestFailed = errors.New("cake: request failed")
