## Cake

Cake is a lightweight HTTP client library for GO, inspired by java [feign](https://github.com/OpenFeign/feign).


### Installation

```bash
# With Go Modules, recommanded with go version > 1.16
require github.com/snownd/cake
```

### Usage

#### Simple Get

```go
type User struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserListRequestConfig struct {
	cake.RequestConfig
	Limit    int    `query:"limit"`
	Page     int    `query:"page"`
}

type TestApi struct {
	Users func(ctx context.Context, config *UserListRequestConfig) ([]*User, error) `method:"GET" url:"/users"`
}

func main() {
	factory := cake.NewFactoryWithClient(http.DefaultClient)
	apiIntf, err := factory.Build(&TestApi{}, cake.WithBaseURL("https://{id}.mockapi.io/api/v1"))
	if err != nil {
		panic(err)
	}
	api := apiIntf.(*TestApi)
	u, err := api.Users(context.Background(), &UserListRequestConfig{
		Limit:    10,
		Page:     1,
	})
	if err != nil {
		panic(err)
	}
	r, _ := json.Marshal(u)
	fmt.Println(string(r))
	if len(u) != 10 {
		panic(errors.New("invalid result set"))
	}
}

```

#### Post with body

```go
type UserCreateRequestConfig struct {
	cake.RequestConfig
  Data *User `body:"application-json"`
}

type TestApi struct {
	Users func(ctx context.Context, config *UserListRequestConfig) ([]*User, error) `method:"GET" url:"/users"`
  CreateUser func(ctx context.Context, config *UserCreateRequestConfig) ([]*User, error) `method:"POST" url:"/users"`
}
```

For more, see [example](./example/main.go)

### Performance

Ran `GOMAXPROCS=1 go test -bench=. -benchtime=5s -benchmem` on a Macbook Pro 2017 i5 with go1.16.7:

```
goos: darwin
goarch: amd64
pkg: github.com/snownd/cake
cpu: Intel(R) Core(TM) i5-7360U CPU @ 2.30GHz
BenchmarkHTTPClientGet             79760             74347 ns/op            6334 B/op         76 allocs/op
BenchmarkCakeGet                   72801             81319 ns/op            7269 B/op         93 allocs/op
BenchmarkHTTPClientPost            72429             82820 ns/op            7796 B/op         91 allocs/op
BenchmarkCakePost                  69199             86652 ns/op            8282 B/op        103 allocs/op
PASS
ok      github.com/snownd/cake  27.511s

For Get request, there are 5 out of 17 allocs simply caused by extra headers like Accept, Accept-Encoding and User-Agent.
```

There is a bit of performance impacts because of uses of reflect. Still, it should be fast enough for most cases.
