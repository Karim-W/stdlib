package stdlib

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap"
)

type UserRoleRequest struct {
	Args    Args        `json:"args"`
	Data    string      `json:"data"`
	Files   Args        `json:"files"`
	Form    Args        `json:"form"`
	Headers Headers     `json:"headers"`
	JSON    interface{} `json:"json"`
	Origin  string      `json:"origin"`
	URL     string      `json:"url"`
}

type Args struct {
}

type Headers struct {
	Accept         string `json:"Accept"`
	AcceptEncoding string `json:"Accept-Encoding"`
	AcceptLanguage string `json:"Accept-Language"`
	ContentLength  string `json:"Content-Length"`
	ContentType    string `json:"Content-Type"`
	Host           string `json:"Host"`
	Origin         string `json:"Origin"`
	Referer        string `json:"Referer"`
	UserAgent      string `json:"User-Agent"`
	XAmznTraceID   string `json:"X-Amzn-Trace-Id"`
}

func TestGetRequst(t *testing.T) {
	c, _ := ClientProvider()
	res := map[string]interface{}{}
	ctx := context.TODO()
	stat, err := c.Get(ctx, "http://localhost:8081/health-check", nil, &res)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	fmt.Println(stat)
}

func TestCreateClient(t *testing.T) {
	c, _ := ClientProvider()
	res := UserRoleRequest{}
	body := map[string]interface{}{
		"name": "test",
	}
	ctx := context.TODO()
	stat, err := c.Post(ctx, "https://httpbin.org/post", nil, body, &res)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	fmt.Println(stat)
}
func TestInvalidCreateTracedClient(t *testing.T) {
	l, err := zap.NewProduction()
	if err != nil {
		t.Error(err)
	}
	c := TracedClientProvider(nil, l)
	ctx := context.TODO()
	body := map[string]interface{}{
		"name": "test",
	}
	var res interface{}
	stat, err := c.Post(ctx, "https://AAAAhttpbin.org/post", nil, body, &res)
	if err == nil {
		t.Error("Expected error")
	}
	fmt.Println(res)
	fmt.Println(stat)
}
func TestValidCreateTracedClient(t *testing.T) {
	l, err := zap.NewProduction()
	if err != nil {
		t.Error(err)
	}
	c := TracedClientProvider(nil, l)
	ctx := context.TODO()
	body := map[string]interface{}{
		"name": "test",
	}
	var res interface{}
	stat, err := c.Post(ctx, "https://httpbin.org/post", nil, body, &res)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	fmt.Println(stat)
}
