package httpclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Soreing/retrier"
	"github.com/karim-w/stdlib"
)

type HTTPRequest interface {
	AddHeader(key string, value string) HTTPRequest
	AddHeaders(headers map[string]string) HTTPRequest
	AddQuery(key string, value string) HTTPRequest
	AddQueryArray(key string, value []string) HTTPRequest
	AddBody(body interface{}) HTTPRequest
	AddBasicAuth(username string, password string) HTTPRequest
	AddBearerAuth(token string) HTTPRequest
	SetNamedPathParams(regexp string, values []string) HTTPRequest
	Dev() HTTPRequest
	DevFromEnv() HTTPRequest
	WithCookie(cookie *http.Cookie) HTTPRequest
	WithRetries(
		policy RetryPolicy,
		retries int,
		amount time.Duration,
	) HTTPRequest
	WithContext(ctx context.Context) HTTPRequest
	// AddBeforeHook(handler func(req *http.Request) error) HTTPRequest
	AddAfterHook(handler func(
		req *http.Request,
		resp *http.Response,
		meta HTTPMetadata,
		err error)) HTTPRequest
	Begin() HTTPRequest
	Get() HTTPResponse
	Put() HTTPResponse
	Del() HTTPResponse
	Post() HTTPResponse
	Patch() HTTPResponse
	Invoke(
		ctx context.Context,
		method string,
		url string,
		opt *stdlib.ClientOptions,
		body interface{},
	) HTTPResponse
}

type _HttpRequest struct {
	httpHooks   *HTTPHook
	statusCode  int
	startTime   time.Time
	endTime     time.Time
	lock        sync.RWMutex
	readOnlyUrl string
	baseUrl     string
	headers     http.Header
	querried    bool
	body        []byte
	err         error
	DevMode     bool
	Cookies     []*http.Cookie
	ctx         context.Context
	withLock    bool
	response    *http.Response
	resBody     []byte
	traces      *clientTrace
	method      string
	client      *http.Client
	retries     struct {
		retryPolicy RetryPolicy
		retryCount  int
		initialWait time.Duration
	}
}

type RetryOptions struct {
	RetryType   string
	RetryPolicy string
	RetryCount  int
}

func (r *_HttpRequest) AddHeader(key string, value string) HTTPRequest {
	r.headers.Add(key, value)
	return r
}

func (r *_HttpRequest) AddHeaders(headers map[string]string) HTTPRequest {
	for k, v := range headers {
		r.headers.Add(k, v)
	}
	return r
}

func (r *_HttpRequest) AddQuery(key string, value string) HTTPRequest {
	if !r.querried {
		r.readOnlyUrl += "?"
		r.querried = true
	} else {
		r.readOnlyUrl += "&"
	}
	r.readOnlyUrl = key + "=" + value
	return r
}

func (r *_HttpRequest) AddQueryArray(key string, value []string) HTTPRequest {
	if !r.querried {
		r.readOnlyUrl += "?"
		r.querried = true
	} else {
		r.readOnlyUrl += "&"
	}
	for _, v := range value {
		r.readOnlyUrl += key + "=" + v + "&"
	}
	r.readOnlyUrl = r.readOnlyUrl[:len(r.readOnlyUrl)-1]
	return r
}

func (r *_HttpRequest) AddBody(body interface{}) HTTPRequest {
	byts, err := json.Marshal(body)
	if err != nil {
		r.err = err
		return r
	}
	copy(r.body, byts)
	return r
}

func (r *_HttpRequest) AddBasicAuth(username string, password string) HTTPRequest {
	unencoded := username + ":" + password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(unencoded))
	r.headers.Add("Authorization", basicAuth)
	return r
}

func (r *_HttpRequest) AddBearerAuth(token string) HTTPRequest {
	bearerAuth := "Bearer " + token
	r.headers.Add("Authorization", bearerAuth)
	return r
}

func (r *_HttpRequest) SetNamedPathParams(regexp string, values []string) HTTPRequest {
	r.readOnlyUrl = stdlib.EmbedNamedPositionArgs(r.readOnlyUrl, values...)
	return r
}

func (r *_HttpRequest) Dev() HTTPRequest {
	r.DevMode = true
	return r
}

func (r *_HttpRequest) DevFromEnv() HTTPRequest {
	envVar := os.Getenv("DEV_MODE")
	if envVar == "" {
		r.DevMode = false
		return r
	}
	r.DevMode = true
	return r
}

func (r *_HttpRequest) WithCookie(cookie *http.Cookie) HTTPRequest {
	r.Cookies = append(r.Cookies, cookie)
	return r
}

func (r *_HttpRequest) WithRetries(
	policy RetryPolicy,
	retries int,
	amount time.Duration,
) HTTPRequest {
	r.retries.retryPolicy = policy
	r.retries.retryCount = retries
	r.retries.initialWait = amount
	return r
}

func (r *_HttpRequest) WithContext(ctx context.Context) HTTPRequest {
	r.ctx = ctx
	return r
}

// func (r *_HttpRequest) AddBeforeHook(handler func(req *http.Request) error) HTTPRequest {
// 	r.httpHooks.Before = append(r.httpHooks.Before, handler)
// 	return r
// }

func (r *_HttpRequest) AddAfterHook(handler func(
	req *http.Request,
	resp *http.Response,
	meta HTTPMetadata,
	err error),
) HTTPRequest {
	r.httpHooks.After = append(r.httpHooks.After, handler)
	return r
}

func (r *_HttpRequest) Begin() HTTPRequest {
	r.lock.Lock()
	r.withLock = true
	return r
}

func (r *_HttpRequest) getRetrier() *retrier.Retrier {
	switch r.retries.retryPolicy {
	case CONSTANT_BACKOFF:
		return retrier.NewRetrier(
			r.retries.retryCount,
			retrier.ConstantDelay(r.retries.initialWait),
		)
	case EXPONENTIAL_BACKOFF:
		return retrier.NewRetrier(
			r.retries.retryCount,
			retrier.ExponentialDelay(2, 2),
		)
	default:
		return nil
	}
}

func (r *_HttpRequest) Get() HTTPResponse {
	r.method = "GET"
	retrier := r.getRetrier()
	if retrier == nil {
		return r.doRequest()
	}
	var resp HTTPResponse
	retrier.Run(func() error {
		resp = r.doRequest()
		if resp.CatchError() != nil {
			return resp.CatchError()
		}
		return nil
	})
	return resp
}

func (r *_HttpRequest) Put() HTTPResponse {
	r.method = "PUT"
	retrier := r.getRetrier()
	if retrier == nil {
		return r.doRequest()
	}
	var resp HTTPResponse
	retrier.Run(func() error {
		resp = r.doRequest()
		if resp.CatchError() != nil {
			return resp.CatchError()
		}
		return nil
	})
	return resp
}

func (r *_HttpRequest) Post() HTTPResponse {
	r.method = "POST"
	retrier := r.getRetrier()
	if retrier == nil {
		return r.doRequest()
	}
	var resp HTTPResponse
	retrier.Run(func() error {
		resp = r.doRequest()
		if resp.CatchError() != nil {
			return resp.CatchError()
		}
		return nil
	})
	return resp
}

func (r *_HttpRequest) Patch() HTTPResponse {
	r.method = "PATCH"
	retrier := r.getRetrier()
	if retrier == nil {
		return r.doRequest()
	}
	var resp HTTPResponse
	retrier.Run(func() error {
		resp = r.doRequest()
		if resp.CatchError() != nil {
			return resp.CatchError()
		}
		return nil
	})
	return resp
}

func (r *_HttpRequest) Del() HTTPResponse {
	r.method = "DELETE"
	retrier := r.getRetrier()
	if retrier == nil {
		return r.doRequest()
	}
	var resp HTTPResponse
	retrier.Run(func() error {
		resp = r.doRequest()
		if resp.CatchError() != nil {
			return resp.CatchError()
		}
		return nil
	})
	return resp
}

func (r *_HttpRequest) Invoke(
	ctx context.Context,
	method string,
	url string,
	opt *stdlib.ClientOptions,
	body interface{},
) HTTPResponse {
	r.WithContext(ctx).
		AddBody(body).
		AddHeaders(*opt.Headers)
	if len(opt.PositionalArgs) > 0 {
		r.SetNamedPathParams(r.readOnlyUrl, opt.PositionalArgs)
	}
	switch method {
	case "GET":
		return r.Get()
	case "POST":
		return r.Post()
	case "PUT":
		return r.Put()
	case "PATCH":
		return r.Patch()
	case "DELETE":
		return r.Del()
	default:
		return r
	}
}

func Req(url string) HTTPRequest {
	return &_HttpRequest{
		readOnlyUrl: url,
		headers:     make(http.Header),
		traces:      &clientTrace{},
		client:      &http.Client{},
		httpHooks: &HTTPHook{
			Before: make([]func(*http.Request) error, 0, 2),
			After:  make([]func(*http.Request, *http.Response, HTTPMetadata, error), 0, 2),
		},
	}
}

func ReqCtx(ctx context.Context, url string) HTTPRequest {
	return &_HttpRequest{
		readOnlyUrl: url,
		headers:     make(http.Header),
		traces:      &clientTrace{},
		client:      &http.Client{},
		ctx:         ctx,
		httpHooks: &HTTPHook{
			Before: make([]func(*http.Request) error, 0, 2),
			After:  make([]func(*http.Request, *http.Response, HTTPMetadata, error), 0, 2),
		},
	}
}
