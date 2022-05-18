## Cake

Cake is a lightweight HTTP client library for GO, inspired by java [feign](https://github.com/OpenFeign/feign).


### Installation

```bash
# With Go Modules, recommanded with go version > 1.16
go get github.com/snownd/cake
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
  Data *User `body:"application/json"`
}

type TestApi struct {
	Users func(ctx context.Context, config *UserListRequestConfig) ([]*User, error) `method:"GET" url:"/users"`
  CreateUser func(ctx context.Context, config *UserCreateRequestConfig) ([]*User, error) `method:"POST" url:"/users"`
}
```

For more, see [example](./example/main.go)

### Performance

Ran `GOMAXPROCS=1 go test -bench=. -benchtime=5s -benchmem` on a Macbook Pro 14(M1 Pro) with go1.18:

```
goos: darwin
goarch: arm64
pkg: github.com/snownd/cake
BenchmarkHTTPClientGet            240624             23627 ns/op            6368 B/op         76 allocs/op
BenchmarkCakeGet                  235514             24969 ns/op            7479 B/op         96 allocs/op
BenchmarkHTTPClientPost           239326             24438 ns/op            7841 B/op         91 allocs/op
BenchmarkCakePost                 222906             26909 ns/op            8866 B/op        106 allocs/op
PASS
ok      github.com/snownd/cake  24.612s

For Get request, there are 5 out of 20 allocs simply caused by extra headers like Accept, Accept-Encoding and User-Agent.
```

There is a bit of performance impacts because of uses of reflect(nearly 8%). Still, it should be fast enough for most cases.
