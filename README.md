## Cake

Cake is a lightweight HTTP client library for GO, inspired by Java [Open-Feign](https://github.com/OpenFeign/feign).


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
	apiIntf, err := factory.Build(&TestApi{}, cake.WithBaseURL("https://61567ea3e039a0001725aa18.mockapi.io/api/v1"))
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
  Data *User `body:""`
}

type TestApi struct {
	Users func(ctx context.Context, config *UserListRequestConfig) ([]*User, error) `method:"GET" url:"/users"`
  CreateUser func(ctx context.Context, config *UserCreateRequestConfig) ([]*User, error) `method:"POST" url:"/users"`
}
```

For more use cases, see [example](./example/main.go)