package cake_test

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/snownd/cake"
	"github.com/stretchr/testify/assert"
)

func TestRequestMiddleware(t *testing.T) {
	path := "/foo/bar"
	mw1Path := ""
	index := 1
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		if assert.Equal(t, 3, index) {
			index++
		} else {
			t.FailNow()
		}
		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte("OK"))
	}))
	defer ts.Close()

	type client struct {
		SimpleGet func(context.Context) (string, error) `url:"/foo/bar"`
	}
	f := cake.New()
	defer f.Close()

	mw1 := func() cake.RequestHandler {
		return func(c *cake.RequestContext) error {
			mw1Path = c.Request.URL.Path
			if assert.Equal(t, 1, index) {
				index++
			}
			if err := c.Next(); err != nil {
				return err
			}
			if assert.Equal(t, 5, index) {
				index++
			}
			return nil
		}
	}
	mw2 := func() cake.RequestHandler {
		return func(c *cake.RequestContext) error {
			if assert.Equal(t, 2, index) {
				index++
			}
			err := c.Next()
			if assert.NoError(t, err) && assert.Equal(t, 4, index) {
				index++
			}
			return nil
		}
	}

	ci, err := f.Build(&client{},
		cake.WithBaseURL(ts.URL),
		cake.WithRequestMiddleware(mw1),
		cake.WithRequestMiddleware(mw2),
	)
	if !assert.NoError(t, err) {
		return
	}
	if c, ok := ci.(*client); assert.True(t, ok) {
		r, err := c.SimpleGet(context.Background())
		if assert.NoError(t, err) {
			assert.Equal(t, "OK", r)
			assert.Equal(t, path, mw1Path)
		}
	}
}

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

func TestFuncHeader(t *testing.T) {
	path := "/foo"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "cake", r.Header.Get("x-app-name"))
		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte("OK"))
	}))
	defer ts.Close()

	type client struct {
		SimpleGet func() (string, error) `url:"/foo" headers:"X-App-Name=cake"`
	}
	f := cake.New()
	defer f.Close()
	ci, err := f.Build(&client{}, cake.WithBaseURL(ts.URL))
	if !assert.NoError(t, err) {
		return
	}
	if c, ok := ci.(*client); assert.True(t, ok) {
		r, err := c.SimpleGet()
		if assert.NoError(t, err) {
			assert.Equal(t, r, "OK")
		}
	}
}

func TestRequestStructHeaders(t *testing.T) {
	name := "cake"
	path := "/foo"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, name, r.Header.Get("x-app-name"))
		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte("OK"))
	}))
	defer ts.Close()
	type headers struct {
		Name string `header:"X-App-Name"`
	}
	type config struct {
		cake.RequestConfig
		Headers *headers `headers:""`
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
		r, err := c.SimpleGet(&config{Headers: &headers{Name: name}})
		if assert.NoError(t, err) {
			assert.Equal(t, r, "OK")
		}
	}
}

func TestRequestMapHeaders(t *testing.T) {
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
		Headers map[string]string `headers:""`
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
		r, err := c.SimpleGet(&config{Headers: map[string]string{"X-App-Name": name}})
		if assert.NoError(t, err) {
			assert.Equal(t, r, "OK")
		}
	}
}

func TestPostRequestWithBody(t *testing.T) {
	path := "/foo"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		rw.Header().Set("Content-Type", "application/json;charset=utf-8")
		data, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		rw.Write(data)
	}))
	defer ts.Close()

	type testBody struct {
		Foo string
		Bar int
	}

	type config struct {
		cake.RequestConfig
		Data testBody `body:"application/json"`
	}

	type client struct {
		PostWithBody func(*config) (*testBody, error) `method:"POST" url:"/foo"`
	}
	f := cake.New()
	defer f.Close()
	ci, err := f.Build(&client{}, cake.WithBaseURL(ts.URL))
	if !assert.NoError(t, err) {
		return
	}
	if c, ok := ci.(*client); assert.True(t, ok) {
		r, err := c.PostWithBody(&config{Data: testBody{Foo: "bar", Bar: 1}})
		if assert.NoError(t, err) {
			assert.Equal(t, r.Foo, "bar")
		}
	}
}

func TestPostRequestWithShortBodytag(t *testing.T) {
	path := "/foo"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		rw.Header().Set("Content-Type", "application/json;charset=utf-8")
		data, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		rw.Write(data)
	}))
	defer ts.Close()

	type testBody struct {
		Foo string
		Bar int
	}

	type config struct {
		cake.RequestConfig
		Data testBody `body:"json"`
	}

	type client struct {
		PostWithBody func(*config) (*testBody, error) `method:"POST" url:"/foo"`
	}
	f := cake.New()
	defer f.Close()
	ci, err := f.Build(&client{}, cake.WithBaseURL(ts.URL))
	if !assert.NoError(t, err) {
		return
	}
	if c, ok := ci.(*client); assert.True(t, ok) {
		r, err := c.PostWithBody(&config{Data: testBody{Foo: "bar", Bar: 1}})
		if assert.NoError(t, err) {
			assert.Equal(t, r.Foo, "bar")
		}
	}
}

func TestPostRequestWithURLEncodedForm(t *testing.T) {
	path := "/foo"
	type testForm struct {
		Foo string `form:"foo"`
		Bar int    `form:"bar"`
	}
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		rw.Header().Set("Content-Type", "application/json;charset=utf-8")
		assert.NoError(t, r.ParseForm())
		f := r.Form.Get("foo")
		b, err := strconv.Atoi(r.Form.Get("bar"))
		assert.NoError(t, err)
		res := testForm{
			Foo: f,
			Bar: b,
		}
		data, err := json.Marshal(res)
		assert.NoError(t, err)
		rw.Write(data)
	}))
	defer ts.Close()

	type config struct {
		cake.RequestConfig
		Data testForm `form:""`
	}

	type client struct {
		PostWithBody func(*config) (*testForm, error) `method:"POST" url:"/foo"`
	}
	f := cake.New()
	defer f.Close()
	ci, err := f.Build(&client{}, cake.WithBaseURL(ts.URL))
	if !assert.NoError(t, err) {
		return
	}
	if c, ok := ci.(*client); assert.True(t, ok) {
		r, err := c.PostWithBody(&config{Data: testForm{Foo: "bar", Bar: 1}})
		if assert.NoError(t, err) {
			assert.Equal(t, r.Foo, "bar")
		}
	}
}

func TestResponseErr(t *testing.T) {
	path := "/foo"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		rw.Header().Set("Content-Type", "application/json")
		rw.Write([]byte(`{"code: 0}`))
	}))
	defer ts.Close()
	type res struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
	}

	type client struct {
		SimpleGet func() (*res, error) `url:"/foo"`
	}
	f := cake.New()
	defer f.Close()
	ci, err := f.Build(&client{}, cake.WithBaseURL(ts.URL))
	if !assert.NoError(t, err) {
		return
	}
	if c, ok := ci.(*client); assert.True(t, ok) {
		_, err := c.SimpleGet()
		if err == nil {
			log.Fatal("should have error")
		}
	}
}
