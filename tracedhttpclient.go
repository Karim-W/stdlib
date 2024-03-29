package stdlib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"time"

	tracer "github.com/BetaLixT/appInsightsTrace"
	"go.uber.org/zap"
)

type TracedClient interface {
	Get(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error)
	Put(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Del(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error)
	Post(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Patch(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Invoke(ctx context.Context, method string, url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	// doRequest(ctx context.Context, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	SetAuthHandler(provider AuthProvider)
	WithTransport(transport http.Transport) TracedClient
	WithStandardTransport() TracedClient
	WithClientName(clientName string) TracedClient
	Close()
}

// ClientOptions is a struct that contains the options for the http client
// Parameters:
//
//		Authorization: The authorization header
//		ContentType: The content type of the request
//		Query: The query string
//		Headers: The headers to be sent with the request
//		Timeout: The timeout for the request
//		RequestType: The type of request to be made
//	  PositionalArgs: The positional arguments to be sent with the request
type ClientOptions struct {
	Authorization  string             `json:"authorization"`
	ContentType    string             `json:"content_type"`
	Query          string             `json:"query"`
	Headers        *map[string]string `json:"headers"`
	Timeout        *time.Time         `json:"timeout"`
	RequestType    string             `json:"request_type"`
	url            string
	method         string
	PositionalArgs []string
}
type tracedhttpCLientImpl struct {
	l          *zap.Logger
	c          *http.Client
	t          *tracer.AppInsightsCore
	auth       AuthProvider
	clientName string
	transport  *http.Transport
}

// TracedClientProvider returns a new instance of the TracedClient
// Params:
//   - t: The tracer to be used check the tracer package for more details at https://github.com/BetaLixT/appInsightsTrace
//   - l: The zap logger to be used
//
// Returns:
//   - TracedClient: The TracedClient instance
func TracedClientProvider(
	t *tracer.AppInsightsCore,
	l *zap.Logger,
) TracedClient {
	return &tracedhttpCLientImpl{
		l: l,
		c: &http.Client{
			Timeout: 30 * time.Second,
		},
		t: t,
	}
}

func (h *tracedhttpCLientImpl) WithTransport(transport http.Transport) TracedClient {
	h.transport = &transport
	return h
}

func (h *tracedhttpCLientImpl) WithStandardTransport() TracedClient {
	h.transport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	return h
}

func (h *tracedhttpCLientImpl) Close() {
	h.c.CloseIdleConnections()
	h.c = nil
}

// TracedClientProviderWithName returns a new instance of the TracedClient
// Params:
//   - t: The tracer to be used check the tracer package for more details at
//   - l: The zap logger to be used
//   - clientName: The name of the client
//
// Returns:
//   - TracedClient: The TracedClient instance
func TracedClientProviderWithName(
	t *tracer.AppInsightsCore,
	l *zap.Logger,
	clientName string,
) TracedClient {
	return &tracedhttpCLientImpl{
		l:          l,
		c:          &http.Client{},
		t:          t,
		clientName: clientName,
	}
}

// SetAuthHandler sets the auth provider for the client
// Params:
//   - provider: The auth provider to be used
//
// Returns:
//   - nil
func (h *tracedhttpCLientImpl) SetAuthHandler(provider AuthProvider) {
	h.auth = provider
}

// WithClientName sets the client name for the client
// Params:
//   - clientName: The name of the client
//
// Returns:
//   - TracedClient: The TracedClient instance
func (h *tracedhttpCLientImpl) WithClientName(clientName string) TracedClient {
	h.clientName = clientName
	return h
}

// Get() makes a GET HTTP request
// Params:
//   - ctx: context.Context => The context to be used
//   - Url: string => The url to be called
//   - opt: *ClientOptions => The options for the request
//   - dest: any => The destination to be used for the response
//
// Returns:
//   - int: The status code of the response
//   - error: The error if any
func (h *tracedhttpCLientImpl) Get(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "GET"
	opt.url = Url + url.QueryEscape(opt.Query)
	return h.doRequest(ctx, opt, nil, dest)
}

// Put() makes a PUT HTTP request
// Params:
//   - ctx: context.Context => The context to be used
//   - Url: string => The url to be called
//   - opt: *ClientOptions => The options for the request
//   - body: any => The body to be sent with the request
//   - dest: any => The destination to be used for the response
//
// Returns:
//   - int: The status code of the response
//   - error: The error if any
func (h *tracedhttpCLientImpl) Put(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "PUT"
	opt.url = Url + url.QueryEscape(opt.Query)
	return h.doRequest(ctx, opt, body, dest)
}

// Patch() makes a PATCH HTTP request
// Params:
//   - ctx: context.Context => The context to be used
//   - Url: string => The url to be called
//   - opt: *ClientOptions => The options for the request
//   - body: any => The body to be sent with the request
//   - dest: any => The destination to be used for the response
//
// Returns:
//   - int: The status code of the response
//   - error: The error if any
func (h *tracedhttpCLientImpl) Patch(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "PATCH"
	opt.url = Url + url.QueryEscape(opt.Query)
	return h.doRequest(ctx, opt, body, dest)
}

// Post() makes a POST HTTP request
// Params:
//   - ctx: context.Context => The context to be used
//   - Url: string => The url to be called
//   - opt: *ClientOptions => The options for the request
//   - body: any => The body to be sent with the request
//   - dest: any => The destination to be used for the response
//
// Returns:
//   - int: The status code of the response
//   - error: The error if any
func (h *tracedhttpCLientImpl) Post(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "POST"
	opt.url = Url + url.QueryEscape(opt.Query)
	return h.doRequest(ctx, opt, body, dest)
}

// Del() makes a DELETE HTTP request
// Params:
//   - ctx: context.Context => The context to be used
//   - Url: string => The url to be called
//   - opt: *ClientOptions => The options for the request
//   - dest: any => The destination to be used for the response
//
// Returns:
//   - int: The status code of the response
//   - error: The error if any
func (h *tracedhttpCLientImpl) Del(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "DELETE"
	opt.url = Url + url.QueryEscape(opt.Query)
	return h.doRequest(ctx, opt, nil, dest)
}

// Invoke()
// Invokes an HTTP request
// Params:
//   - ctx: context.Context => The context to be used
//   - method: string => The method to be used
//   - opt: *ClientOptions => The options for the request
//   - body: any => The body to be sent with the request
//   - dest: any => The destination to be used for the response
//
// Returns:
//   - int: The status code of the response
//   - error: The error if any
func (h *tracedhttpCLientImpl) Invoke(
	ctx context.Context, method string, Url string,
	opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = method
	opt.url = EmbedNamedPositionArgs(Url, opt.PositionalArgs...) + url.QueryEscape(opt.Query)
	return h.doRequest(ctx, opt, body, dest)
}

func (h *tracedhttpCLientImpl) doRequest(ctx context.Context, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	req, err := http.NewRequest(opt.method, opt.url, nil)
	if err != nil {
		return 0, err
	}
	if h.transport != nil {
		h.c.Transport = h.transport
	}
	contentType, reqBody, err := h.formulatePayload(body, opt.RequestType)
	if err != nil {
		return 0, err
	}
	hd := opt.Headers
	if hd != nil {
		for k, v := range *hd {
			req.Header.Set(k, v)
		}
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", contentType)
	if h.auth != nil {
		req.Header.Add("Authorization", h.auth.GetAuthHeader())
	}
	ver, tid, rid, sid, flg := "", "", "", "", ""
	if h.t != nil {
		sid, err := GenerateParentId()
		ver, tid, _, rid, flg = h.t.ExtractTraceInfo(ctx)
		if err == nil {
			req.Header.Add(
				"traceparent",
				fmt.Sprintf("%s-%s-%s-%s", ver, tid, sid, flg),
			)
		} else {
			req.Header.Add(
				"traceparent",
				fmt.Sprintf(
					"%s-%s-%s-%s",
					ver,
					tid,
					rid,
					flg,
				),
			)
		}
	}
	if body != nil {
		req.Body = reqBody
	}
	// remote name
	var remoteName string
	if h.clientName != "" {
		remoteName = h.clientName
	} else {
		remoteName = req.URL.Hostname()
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	now := time.Now()
	resp, err := h.c.Do(req)
	if err != nil {
		code := 502
		if resp != nil {
			code = resp.StatusCode
			defer resp.Body.Close()
		}
		if h.t != nil {
			h.t.TraceException(ctx, err, 0, nil)
			h.t.TraceDependency(ctx, sid, "http", remoteName,
				fmt.Sprintf("%s %s", req.Method, req.URL.RequestURI()), false, now, time.Now(), map[string]string{
					"code":         fmt.Sprintf("%d", code),
					"errorMessage": err.Error(),
				})
		}
		return code, err
	}
	defer resp.Body.Close()
	if h.t != nil {
		h.t.TraceDependency(ctx, sid, "http", remoteName,
			fmt.Sprintf("%s %s", req.Method, req.URL.RequestURI()), resp.StatusCode > 199 && resp.StatusCode < 300, now, time.Now(), map[string]string{
				"code": fmt.Sprintf("%d", resp.StatusCode),
			})
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if resp.Body != nil && (resp.ContentLength > 4 || resp.ContentLength == -1) && dest != nil {
			if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
				h.l.Error("Error decoding response body", zap.Error(err))
				return resp.StatusCode, fmt.Errorf("error decoding response: %v", err)
			}
		}
		return resp.StatusCode, nil
	}
	if resp.Body != nil && resp.ContentLength > 4 {
		body, _ := ioutil.ReadAll(resp.Body)
		h.l.Error("error response from server", zap.Int("code", resp.StatusCode),
			zap.String("response", string(body)))
		return resp.StatusCode, fmt.Errorf("got non 200 code (%d) calling %s", resp.StatusCode, opt.url)
	}
	var respo []byte
	if err := json.NewDecoder(resp.Body).Decode(&respo); err != nil {
		h.l.Error("Error decoding response body", zap.Error(err))
		return resp.StatusCode, fmt.Errorf("error decoding response: %v", err)
	}
	h.l.Error("error response from server", zap.Int("code", resp.StatusCode), zap.String("Response", string(respo)))
	return resp.StatusCode, fmt.Errorf("got non 200 code (%d) calling %s", resp.StatusCode, opt.url)
}

func (h *tracedhttpCLientImpl) formulatePayload(body interface{}, rType string) (string, io.ReadCloser, error) {
	switch rType {
	case "json":
		if strBody, err := json.Marshal(body); err != nil {
			return "", nil, err
		} else {
			return "application/json", ioutil.NopCloser(bytes.NewBuffer(strBody)), nil
		}
	case "www-form-urlencoded":
		if kv, ok := body.(map[string]interface{}); ok {
			strBody := ""
			for k, v := range kv {
				strBody += k + "=" + fmt.Sprintf("%v", v) + "&"
			}
			//remove last &
			strBody = strBody[:len(strBody)-1]
			return "application/x-www-form-urlencoded", ioutil.NopCloser(bytes.NewBuffer([]byte(strBody))), nil
		} else {
			return "", nil, fmt.Errorf("invalid body type for www-form-urlencoded")
		}
	case "form-data":
		if kv, ok := body.(map[string]interface{}); !ok {
			return "", nil, fmt.Errorf("body must be a map[string]interface{} ideally map[string]string but i compromised")
		} else {
			payload := &bytes.Buffer{}
			writer := multipart.NewWriter(payload)
			_ = writer.WriteField("r", "r")
			_ = writer.WriteField("s", "s")
			for k, v := range kv {
				if str, ok := v.(string); ok {
					_ = writer.WriteField(string(k), str)
				} else {
					_ = writer.WriteField(string(k), fmt.Sprintf("%v", v))
				}
			}
			if err := writer.Close(); err != nil {
				return "", nil, err
			} else {
				return "multipart/form-data", ioutil.NopCloser(payload), nil
			}
		}
	case "graphql":
		return "", nil, fmt.Errorf("graphql is not supported yet")
	default:
		if body != nil {
			if strBody, err := json.Marshal(body); err != nil {
				return "", nil, err
			} else {
				return "application/json", ioutil.NopCloser(bytes.NewBuffer(strBody)), nil
			}
		} else {
			return "application/json", nil, nil
		}
	}
}
