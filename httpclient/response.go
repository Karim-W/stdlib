package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type HTTPResponse interface {
	GetStatusCode() int
	SetResult(responseBody any) error
	CatchError() error
	Catch(errorObject any) error
	IsSuccess() bool
	GetTraceInfo() HttpTraceInfo
	GetUrl() string
	GetMethod() string
	GetHeaders() http.Header
	GetBody() []byte
	GetCookies() []*http.Cookie
	GetElapsedTime() time.Duration
}

func (r *_HttpRequest) GetStatusCode() int { return r.statusCode }

func (r *_HttpRequest) SetResult(responseBody any) error {
	if r.err != nil {
		return r.err
	}
	return json.Unmarshal(*r.resBody, responseBody)
}

func (r *_HttpRequest) CatchError() error {
	if r.err != nil {
		return r.err
	}
	if r.statusCode > 299 || r.statusCode < 200 {
		return fmt.Errorf("request failed with status code %d", r.statusCode)
	}
	return nil
}

func (r *_HttpRequest) Catch(errorObject any) error { return json.Unmarshal(*r.resBody, errorObject) }

func (r *_HttpRequest) IsSuccess() bool {
	return r.statusCode >= 200 && r.statusCode < 300 && r.err == nil
}

// RUN REQUEST
func (r *_HttpRequest) afterRequest() {
	r.withLock = false
	r.lock.Unlock()
	r.traces.endTime = time.Now()
}

func (r *_HttpRequest) doRequest() HTTPResponse {
	if r.withLock {
		defer r.afterRequest()
	}
	if r.err != nil {
		return r
	}
	var req *http.Request
	if r.body != nil {
		req, r.err = http.NewRequest(r.method, r.readOnlyUrl, bytes.NewBuffer(*r.body))
	} else {
		req, r.err = http.NewRequest(r.method, r.readOnlyUrl, nil)
	}
	if r.err != nil {
		return r
	}
	for i := range r.httpHooks.Before {
		r.httpHooks.Before[i](req)
	}
	req = req.WithContext(r.traces.CreateContext(r.ctx))
	req.Header = r.headers
	for _, cookie := range r.Cookies {
		req.AddCookie(cookie)
	}
	r.startTime = time.Now()
	r.response, r.err = r.client.Do(req)
	for i := range r.httpHooks.After {
		r.httpHooks.After[i](req, r.response, r.err)
	}
	if r.err != nil {
		r.statusCode = -1
		return r
	}
	r.statusCode = r.response.StatusCode
	defer r.response.Body.Close()
	var byts []byte
	byts, r.err = ioutil.ReadAll(r.response.Body)
	if r.err != nil {
		return r
	}
	r.resBody = &byts
	return r
}

func (r *_HttpRequest) GetTraceInfo() HttpTraceInfo {

	endTime := r.traces.endTime
	if endTime.IsZero() {
		endTime = time.Now()
	}
	ti := HttpTraceInfo{
		IsConnReused:  r.traces.gotConnInfo.Reused,
		IsConnWasIdle: r.traces.gotConnInfo.WasIdle,
		ConnIdleTime:  r.traces.gotConnInfo.IdleTime,
	}
	if !r.traces.tlsHandshakeStart.IsZero() && !r.traces.tlsHandshakeDone.IsZero() {
		ti.TLSHandshakeTime = r.traces.tlsHandshakeDone.Sub(r.traces.tlsHandshakeStart)
	} else {
		ti.TLSHandshakeTime = endTime.Sub(r.traces.tlsHandshakeStart)
	}

	ti.TotalTime = endTime.Sub(r.traces.getConn)

	dnsDone := r.traces.dnsDone
	if dnsDone.IsZero() {
		dnsDone = endTime
	}

	if !r.traces.dnsStart.IsZero() {
		ti.DNSLookupTime = dnsDone.Sub(r.traces.dnsStart)
	}

	// Only calculate on successful conner.traces.ons
	if !r.traces.connectDone.IsZero() {
		ti.TCPConnectTime = r.traces.connectDone.Sub(dnsDone)
	}

	// Only calculate on successful conner.traces.ons
	if !r.traces.gotConn.IsZero() {
		ti.ConnectTime = r.traces.gotConn.Sub(r.traces.getConn)
	}

	// Only calculate on successful conner.traces.ons
	if !r.traces.gotFirstResponseByte.IsZero() {
		ti.FirstResponseTime = r.traces.gotFirstResponseByte.Sub(r.traces.gotConn)
		ti.ResponseTime = endTime.Sub(r.traces.gotFirstResponseByte)
	}

	// Capture remote address info when conner.traces.on is non-nil
	if r.traces.gotConnInfo.Conn != nil {
		ti.RemoteAddr = r.traces.gotConnInfo.Conn.RemoteAddr()
	}

	return ti
}

func (r *_HttpRequest) GetUrl() string { return r.readOnlyUrl }

func (r *_HttpRequest) GetMethod() string { return r.method }

func (r *_HttpRequest) GetHeaders() http.Header { return r.headers }

func (r *_HttpRequest) GetBody() []byte { return *r.body }

func (r *_HttpRequest) GetCookies() []*http.Cookie { return r.Cookies }

func (r *_HttpRequest) GetElapsedTime() time.Duration { return r.traces.endTime.Sub(r.startTime) }
