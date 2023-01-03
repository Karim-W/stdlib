package stdlib

import (
	"context"
	"net/http"
	"regexp"

	appinsightstrace "github.com/BetaLixT/appInsightsTrace"
	"go.uber.org/zap"
)

type Client interface {
	Get(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error)
	Put(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Del(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error)
	Post(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Patch(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Invoke(ctx context.Context, method string, url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	doRequest(ctx context.Context, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	SetAuthHandler(provider AuthProvider)
	WithLogger(l *zap.Logger) Client
	WithTracer(t *appinsightstrace.AppInsightsCore) Client
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
		c: &http.Client{},
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

// WithTracer returns a new instance of Client with a new tracer
// params:
//   - t *appinsightstrace.AppInsightsCore
//
// returns:
//   - Client
func (t *tracedhttpCLientImpl) WithTracer(tracer *appinsightstrace.AppInsightsCore) Client {
	return &tracedhttpCLientImpl{
		l: t.l,
		c: t.c,
		t: tracer,
	}
}

type AuthProvider interface {
	GetAuthHeader() string
}

func EmbedNamedPositionArgs(stringObject string, args ...string) string {
	// example stringObject: "https://example.com/{server}/test/{test_name}"
	// example args: "server1", "test1"
	// result: "https://example.com/server1/test/test1"
	expression := regexp.MustCompile(`\{[a-zA-Z0-9_]+\}`)
	for _, arg := range args {
		index := expression.FindStringIndex(stringObject)
		if index == nil {
			break
		}
		stringObject = stringObject[:index[0]] + arg + stringObject[index[1]:]
	}
	return stringObject
}
