package stdlib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
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
	doRequest(ctx context.Context, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
}

type tracedhttpCLientImpl struct {
	l *zap.Logger
	c http.Client
	t *tracer.AppInsightsCore
}

func TracedClientProvider(
	t *tracer.AppInsightsCore,
	l *zap.Logger,
) TracedClient {
	return &tracedhttpCLientImpl{
		l: l,
		c: http.Client{},
		t: t,
	}
}

func (h *tracedhttpCLientImpl) Get(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "GET"
	opt.url = Url + url.QueryEscape(opt.Query)
	now := time.Now()
	code, err := h.doRequest(ctx, opt, nil, dest)
	if err != nil {
		h.t.TraceException(ctx, err, 0, nil)
		h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, false, now, time.Now(), map[string]string{
			"code":         fmt.Sprintf("%d", code),
			"errorMessage": err.Error(),
		})
		return code, err
	}
	h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, true, now, time.Now(), map[string]string{
		"code": fmt.Sprintf("%d", code),
	})
	return code, err
}

func (h *tracedhttpCLientImpl) Put(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "PUT"
	opt.url = Url + url.QueryEscape(opt.Query)
	now := time.Now()
	code, err := h.doRequest(ctx, opt, body, dest)
	if err != nil {
		h.t.TraceException(ctx, err, 0, nil)
		h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, false, now, time.Now(), map[string]string{
			"code":         fmt.Sprintf("%d", code),
			"errorMessage": err.Error(),
		})
		return code, err
	}
	h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, true, now, time.Now(), map[string]string{
		"code": fmt.Sprintf("%d", code),
	})
	return code, err
}

func (h *tracedhttpCLientImpl) Patch(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "PATCH"
	opt.url = Url + url.QueryEscape(opt.Query)
	now := time.Now()
	code, err := h.doRequest(ctx, opt, body, dest)
	if err != nil {
		h.t.TraceException(ctx, err, 0, nil)
		h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, false, now, time.Now(), map[string]string{
			"code":         fmt.Sprintf("%d", code),
			"errorMessage": err.Error(),
		})
		return code, err
	}
	h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, true, now, time.Now(), map[string]string{
		"code": fmt.Sprintf("%d", code),
	})
	return code, err
}

func (h *tracedhttpCLientImpl) Post(ctx context.Context, Url string, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "POST"
	opt.url = Url + url.QueryEscape(opt.Query)
	now := time.Now()
	code, err := h.doRequest(ctx, opt, body, dest)
	if err != nil {
		h.t.TraceException(ctx, err, 0, nil)
		h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, false, now, time.Now(), map[string]string{
			"code":         fmt.Sprintf("%d", code),
			"errorMessage": err.Error(),
		})
		return code, err
	}
	h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, true, now, time.Now(), map[string]string{
		"code": fmt.Sprintf("%d", code),
	})
	return code, err
}

func (h *tracedhttpCLientImpl) Del(ctx context.Context, Url string, opt *ClientOptions, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "DELETE"
	opt.url = Url + url.QueryEscape(opt.Query)
	now := time.Now()
	code, err := h.doRequest(ctx, opt, nil, dest)
	if err != nil {
		h.t.TraceException(ctx, err, 0, nil)
		h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, false, now, time.Now(), map[string]string{
			"code":         fmt.Sprintf("%d", code),
			"errorMessage": err.Error(),
		})
		return code, err
	}
	h.t.TraceDependency(ctx, "0000", "http", Url, opt.method+" "+Url, true, now, time.Now(), map[string]string{
		"code": fmt.Sprintf("%d", code),
	})
	return code, err
}

func (h *tracedhttpCLientImpl) doRequest(ctx context.Context, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if req, err := http.NewRequest(opt.method, opt.url, nil); err != nil {
		return 0, err
	} else {
		if contentType, reqBody, err := h.formulatePayload(body, opt.RequestType); err != nil {
			return 0, err
		} else {
			hd := opt.Headers
			if hd != nil {
				for k, v := range *hd {
					req.Header.Set(k, v)
				}
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", contentType)
			if body != nil {
				req.Body = reqBody
			}
			now := time.Now()
			if resp, err := h.c.Do(req); err != nil {
				return 0, err
			} else {
				latency := time.Since(now).Microseconds()
				h.l.Info("[HTTP Client]",
					zap.String("method", opt.method),
					zap.String("url", opt.url),
					zap.Int("status", resp.StatusCode),
					zap.Float64("latency", float64(latency)/1000.0),
				)
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					if resp.Body != nil && resp.ContentLength != 0 {
						if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
							return resp.StatusCode, fmt.Errorf("error decoding response: %v", err)
						}
					}
					return resp.StatusCode, nil
				} else {
					return resp.StatusCode, fmt.Errorf("got non 200 code (%d) calling %s", resp.StatusCode, opt.url)
				}
			}
		}
	}
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
