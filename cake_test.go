package cake_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/snownd/cake"
	"github.com/stretchr/testify/assert"
)

func TestGetWithPathParam(t *testing.T) {
	name := "cake"
	path := "/foo/" + name
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte("OK"))
	}))
	defer ts.Close()

	type config struct {
		cake.RequestConfig
		Name string `param:"name"`
	}

	type client struct {
		SimpleGet func(*config) (string, error) `url:"/foo/:name"`
	}
	f := cake.New()
	defer f.Close()
	ci, err := f.Build(&client{}, cake.WithBaseURL(ts.URL))
	if !assert.NoError(t, err) {
		return
	}
	if c, ok := ci.(*client); assert.True(t, ok) {
		r, err := c.SimpleGet(&config{Name: name})
		if assert.NoError(t, err) {
			assert.Equal(t, r, "OK")
		}
	}
}

func TestGetWithQuery(t *testing.T) {
	queryStr := "queryStr"
	queryInt := 100
	queryBool := true
	queryFloat := 100.1
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, queryStr, query.Get("queryStrKey"))
		assert.Equal(t, strconv.Itoa(queryInt), query.Get("queryIntKey"))
		assert.Equal(t, "true", query.Get("queryBoolKey"))
		assert.Equal(t, strconv.FormatFloat(queryFloat, 'f', -1, 64), query.Get("queryFloatKey"))
		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte("OK"))
	}))
	defer ts.Close()

	type config struct {
		cake.RequestConfig
		QueryStr   string  `query:"queryStrKey"`
		QueryInt   int     `query:"queryIntKey"`
		QueryBool  bool    `query:"queryBoolKey"`
		QueryFloat float64 `query:"queryFloatKey"`
	}

	type client struct {
		Query func(*config) (string, error) `url:"/foo"`
	}
	f := cake.New()
	defer f.Close()
	ci, err := f.Build(&client{}, cake.WithBaseURL(ts.URL))
	if !assert.NoError(t, err) {
		return
	}
	if c, ok := ci.(*client); assert.True(t, ok) {
		r, err := c.Query(&config{
			QueryStr:   queryStr,
			QueryInt:   queryInt,
			QueryBool:  queryBool,
			QueryFloat: queryFloat,
		})
		if assert.NoError(t, err) {
			assert.Equal(t, r, "OK")
		}
	}
}

func TestConfigHeader(t *testing.T) {
	name := "cake"
	path := "/foo"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, name, r.Header.Get("x-app-name"))
		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte("OK"))
	}))
	defer ts.Close()

	type config struct {
		cake.RequestConfig
		Name string `header:"X-App-Name"`
	}

	type client struct {
		SimpleGet func(*config) (string, error) `url:"/foo"`
	}
	f := cake.New()
	defer f.Close()
	ci, err := f.Build(&client{}, cake.WithBaseURL(ts.URL))
	if !assert.NoError(t, err) {
		return
	}
	if c, ok := ci.(*client); assert.True(t, ok) {
		r, err := c.SimpleGet(&config{Name: name})
		if assert.NoError(t, err) {
			assert.Equal(t, r, "OK")
		}
	}
}
