package stdlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

type Client interface {
	Get(ctx Context, baseUrl string, query string, opt *ClientOptions, dest interface{}) (int, error)
	Put(ctx Context, baseUrl string, query string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Del(ctx Context, baseUrl string, query string, opt *ClientOptions, dest interface{}) (int, error)
	Post(ctx Context, baseUrl string, query string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	Patch(ctx Context, baseUrl string, query string, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
	doRequest(ctx Context, opt *ClientOptions, body interface{}, dest interface{}) (int, error)
}

type ClientOptions struct {
	Authorization string             `json:"authorization"`
	ContentType   string             `json:"content_type"`
	Query         string             `json:"query"`
	Headers       *map[string]string `json:"headers"`
	Timeout       *time.Time         `json:"timeout"`
	RequestType   string             `json:"request_type"`
	url           string
	method        string
}

type httpCLientImpl struct {
	l *logger
	c http.Client
}

func ClientProvider() Client {
	return &httpCLientImpl{
		l: getLoggerInstance(),
		c: http.Client{},
	}
}

func (h *httpCLientImpl) Get(ctx Context, baseUrl string, query string, opt *ClientOptions, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "GET"
	opt.url = baseUrl + url.QueryEscape(query)
	return h.doRequest(ctx, opt, nil, dest)
}

func (h *httpCLientImpl) Put(ctx Context, baseUrl string, query string, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "PUT"
	opt.url = baseUrl + url.QueryEscape(query)
	return h.doRequest(ctx, opt, body, dest)
}

func (h *httpCLientImpl) Patch(ctx Context, baseUrl string, query string, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "PATCH"
	opt.url = baseUrl + url.QueryEscape(query)
	return h.doRequest(ctx, opt, body, dest)
}

func (h *httpCLientImpl) Post(ctx Context, baseUrl string, query string, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "POST"
	opt.url = baseUrl + url.QueryEscape(query)
	return h.doRequest(ctx, opt, body, dest)
}

func (h *httpCLientImpl) Del(ctx Context, baseUrl string, query string, opt *ClientOptions, dest interface{}) (int, error) {
	if opt == nil {
		opt = &ClientOptions{}
	}
	opt.method = "DELETE"
	opt.url = baseUrl + url.QueryEscape(query)
	return h.doRequest(ctx, opt, nil, dest)
}

func (h *httpCLientImpl) doRequest(ctx Context, opt *ClientOptions, body interface{}, dest interface{}) (int, error) {
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
				h.l.Infof("[HTTP Client]\t%s\t%s\tStatus:%d\tLatency:%0.2f ms", opt.method, opt.url, resp.StatusCode, float64(latency)/1000.0)
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

func (h *httpCLientImpl) formulatePayload(body interface{}, rType string) (string, io.ReadCloser, error) {
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
