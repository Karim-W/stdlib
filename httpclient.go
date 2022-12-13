package stdlib

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type Client interface {
	Get(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error)
	Put(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Del(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error)
	Post(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Patch(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	doRequest(ctx context.Context, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	SetAuthHandler(provider AuthProvider)
	WithLogger(l *zap.Logger) Client
}

// ClientProvider returns a new instance of Client
// This is a constructor function
// params:
//   - none
//
// returns:
//   - Client
//   - error
func ClientProvider() (Client, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &tracedhttpCLientImpl{
		l: l,
		c: http.Client{},
	}, nil
}

// WithLogger returns a new instance of Client with a new logger
// params:
//   - l *zap.Logger
//
// returns:
//   - Client
func (t *tracedhttpCLientImpl) WithLogger(l *zap.Logger) Client {
	return &tracedhttpCLientImpl{
		l: l,
		c: t.c,
	}
}

type AuthProvider interface {
	GetAuthHeader() string
}
