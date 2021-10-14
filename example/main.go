package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/snownd/cake"
)

type User struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserRequestConfig struct {
	cake.RequestConfig
	ID string `param:"id"`
}

type UserListRequestConfig struct {
	cake.RequestConfig
	Limit    int    `query:"limit"`
	Page     int    `query:"page"`
	XTraceID string `header:"X-Trace-Id"`
}

// https://61567ea3e039a0001725aa18.mockapi.io/api/v1
// {
//   "createdAt": "2021-09-30T22:22:18.253Z",
//   "name": "Andres Weissnat",
//   "avatar": "https://cdn.fakercloud.com/avatars/smalonso_128.jpg",
//   "phone": "505-588-2692",
//   "id": "1"
//  },
type TestApi struct {
	// default method = GET
	User  func(ctx context.Context, config *UserRequestConfig) (*User, error)       `url:"/users/:id" headers:"x-request-name=users;x-request-app=cake-example"`
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
		XTraceID: "caketest1",
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
