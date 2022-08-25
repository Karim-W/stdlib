package stdlib

import (
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
	l, _ := zap.NewProduction()
	c := ClientProvider(l)
	res := map[string]interface{}{}
	ctx := NewContext()
	stat, err := c.Get(ctx, "http://localhost:8081/health-check", "", nil, &res)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	fmt.Println(stat)
}

func TestCreateClient(t *testing.T) {
	l, _ := zap.NewProduction()
	c := ClientProvider(l)
	res := UserRoleRequest{}
	body := map[string]interface{}{
		"name": "test",
	}
	ctx := NewContext()
	stat, err := c.Post(ctx, "https://httpbin.org/post", "", nil, body, &res)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	fmt.Println(stat)
}
