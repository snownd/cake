package cake_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/snownd/cake"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type TestData struct {
	Foo string   `json:"foo"`
	Bar []string `json:"bar"`
}

type TestGetRequest struct {
	cake.RequestConfig
	ID       string `param:"id"`
	QueryFoo int    `query:"foo"`
}

type TestPostRequest struct {
	cake.RequestConfig
	Data *TestData `body:"application/json"`
}

type TestCakeFoo struct {
	GetData  func(*TestGetRequest) (*TestData, error)  `url:"/data/:id"`
	PostData func(*TestPostRequest) (*TestData, error) `method:"POST" url:"/data"`
}

var msg = []byte(`{"foo":"bar", "bar":["foo1", "foo2"]}`)

func BenchmarkHTTPClientGet(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if queryFoo := r.URL.Query().Get("foo"); queryFoo == "" {
			b.Fatal("queryMissed", r.URL)
		}
		rw.Header().Add("Content-Type", "application/json")
		rw.Write(msg)
	}))
	defer ts.Close()

	tr := &http.Transport{}
	defer tr.CloseIdleConnections()
	cl := &http.Client{
		Transport: tr,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, err := cl.Get(ts.URL + "/data/" + strconv.Itoa(i) + "?foo=" + strconv.Itoa(i))
		if err != nil {
			b.Fatal("Get:", err)
		}
		body, e := io.ReadAll(res.Body)
		if e != nil {
			b.Fatal("ReadAll:", e)
		}
		data := &TestData{}
		if e = json.Unmarshal(body, data); e != nil {
			b.Fatal("JsonUnmarshal", e)
		}
		if len(data.Bar) == 0 {
			b.Fatal("Json:", string(body))
		}
		res.Body.Close()
	}
}

func BenchmarkCakeGet(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if queryFoo := r.URL.Query().Get("foo"); queryFoo == "" {
			b.Fatal("queryMissed", r.URL)
		}
		rw.Header().Add("Content-Type", "application/json")
		rw.Write(msg)

	}))
	defer ts.Close()

	tr := &http.Transport{}
	defer tr.CloseIdleConnections()
	hc := &http.Client{
		Transport: tr,
	}
	t, err := cake.NewFactoryWithClient(hc).Build(&TestCakeFoo{}, cake.WithBaseURL(ts.URL))
	if err != nil {
		b.Fatal("CakeBuild:", err)
	}
	c := t.(*TestCakeFoo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, e := c.GetData(&TestGetRequest{ID: strconv.Itoa(i), QueryFoo: i})
		if e != nil {
			b.Fatal("CakeGet", e)
		}
		if len(r.Bar) == 0 {
			b.Fatal("CakeJson:", r)
		}
	}
}

func BenchmarkHTTPClientPost(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			b.Fatal("ReadBody:", err)
		}
		rw.Header().Add("Content-Type", "application/json")
		rw.Write(data)
	}))
	defer ts.Close()

	tr := &http.Transport{}
	defer tr.CloseIdleConnections()
	cl := &http.Client{
		Transport: tr,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reqData := &TestData{
			Foo: strconv.Itoa(i),
			Bar: make([]string, 2),
		}
		reqBody, _ := json.Marshal(reqData)
		res, err := cl.Post(ts.URL+"/data", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			b.Fatal("Post:", err)
		}
		body, e := io.ReadAll(res.Body)
		if e != nil {
			b.Fatal("ReadAll:", e)
		}
		data := &TestData{}
		if e = json.Unmarshal(body, data); e != nil {
			b.Fatal("JsonUnmarshal", e)
		}
		if data.Foo != reqData.Foo {
			b.Fatal("ResponseJson:", string(body))
		}
		res.Body.Close()
	}
}

func BenchmarkCakePost(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			b.Fatal("ReadBody:", err)
		}
		rw.Header().Add("Content-Type", "application/json")
		rw.Write(data)
	}))
	defer ts.Close()

	tr := &http.Transport{}
	defer tr.CloseIdleConnections()
	hc := &http.Client{
		Transport: tr,
	}
	t, err := cake.NewFactoryWithClient(hc).Build(&TestCakeFoo{}, cake.WithBaseURL(ts.URL))
	if err != nil {
		b.Fatal("CakeBuild:", err)
	}
	c := t.(*TestCakeFoo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := &TestData{
			Foo: strconv.Itoa(i),
			Bar: make([]string, 2),
		}
		r, e := c.PostData(&TestPostRequest{
			Data: data,
		})
		if e != nil {
			b.Fatal("CakeGet", e)
		}
		if data.Foo != r.Foo {
			b.Fatal("ResponseJson:", r)
		}
	}
}
