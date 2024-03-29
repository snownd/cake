package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/snownd/cake"
)

const mockAPIID = ""

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
	Limit int    `query:"limit"`
	Page  int    `query:"page"`
	Token string `header:"Authorization"`
}

type UserCreateRequestConfig struct {
	cake.RequestConfig
	// default content-type = text/plain when use body tag like `body:""`
	// use cake.WithDefaultContenType("application/json") can change it
	Data *User `body:"application/json"`
}

type TestApi struct {
	// default method = GET
	User       func(ctx context.Context, config *UserRequestConfig) (*User, error)       `url:"/users/:id" headers:"x-request-name=users;x-request-app=cake-example"`
	Users      func(ctx context.Context, config *UserListRequestConfig) ([]*User, error) `method:"GET" url:"/users"`
	CreateUser func(ctx context.Context, config *UserCreateRequestConfig) (*User, error) `method:"POST" url:"/users"`
	DeleteUser func(ctx context.Context, config *UserRequestConfig) (*User, error)       `method:"DELETE" url:"/users/:id"`
}

func main() {

	factory := cake.New()
	defer factory.Close()
	// click https://mockapi.io/clone/61567ea3e039a0001725aa19 to create a mockapi project
	apiIntf, err := factory.Build(&TestApi{}, cake.WithBaseURL("https://"+mockAPIID+".mockapi.io/api/v1"), cake.WithRequestMiddleware(func(c *cake.RequestContext) error {
		url := c.Request.URL.String()
		method := c.Request.Method
		fmt.Println("requestTo", url, "method", method)
		start := time.Now()
		c.Next()
		fmt.Println("requestTo", url, "cost", time.Since(start))
		return nil
	}))
	if err != nil {
		panic(err)
	}
	api := apiIntf.(*TestApi)
	newUser, err := api.CreateUser(context.Background(), &UserCreateRequestConfig{
		Data: &User{
			Phone:     "710-839-4565 x109",
			Name:      "New User",
			Avatar:    "https://cdn.fakercloud.com/avatars/carlosjgsousa_128.jpg",
			CreatedAt: time.Now(),
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("create user", newUser)
	user, err := api.User(context.Background(), &UserRequestConfig{
		ID: newUser.ID,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("get user", user)
	if newUser.Name != user.Name {
		panic(errors.New("get wrong user"))
	}
	dUser, err := api.DeleteUser(context.Background(), &UserRequestConfig{
		ID: newUser.ID,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("delete user", dUser)
	_, err = api.User(context.Background(), &UserRequestConfig{
		ID: newUser.ID,
	})
	if err != nil {
		var rErr cake.RequestError
		if errors.As(err, &rErr) {
			fmt.Printf("expect 404, got: %d ,res %s \n", rErr.StatusCode(), rErr.Body())
		} else {
			panic(err)
		}
	}

	users, err := api.Users(context.Background(), &UserListRequestConfig{
		Limit: 10,
		Page:  1,
		Token: "Bearer Y2FrZS1leGFtcGxlLXRva2Vu",
	})
	if err != nil {
		panic(err)
	}
	if len(users) != 10 {
		panic(errors.New("invalid result set"))
	}
}
