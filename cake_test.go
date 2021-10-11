package cake_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/snownd/cake"
)

type TestRes struct {
	Foo string   `json:"foo"`
	Bar []string `json:"bar"`
}

type TestGetRequest struct {
	cake.RequestConfig
	ID       string `param:"id"`
	QueryFoo int    `query:"foo"`
}

type TestCakeFoo struct {
	GetData func(*TestGetRequest) (*TestRes, error) `url:"/data/:id"`
}

var msg = []byte(`{"foo":"bar", "bar":["foo1", "foo2"]}`)

func BenchmarkHTTPClientGet(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
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
		res, err := cl.Get(ts.URL + "/data/" + strconv.Itoa(i) + "?foo" + strconv.Itoa(i))
		if err != nil {
			b.Fatal("Get:", err)
		}
		body, e := io.ReadAll(res.Body)
		if e != nil {
			b.Fatal("ReadAll:", e)
		}
		data := &TestRes{}
		json.Unmarshal(body, data)
		if len(data.Bar) == 0 {
			b.Fatal("json:", string(body))
		}
	}
}

func BenchmarkCakeGet(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
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
