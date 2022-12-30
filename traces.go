package stdlib

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"
)

type clientTrace struct {
	getConn              time.Time
	dnsStart             time.Time
	dnsDone              time.Time
	connectDone          time.Time
	tlsHandshakeStart    time.Time
	tlsHandshakeDone     time.Time
	gotConn              time.Time
	gotFirstResponseByte time.Time
	endTime              time.Time
	gotConnInfo          httptrace.GotConnInfo
}

type HTTPHook struct {
	Before []func(req *http.Request)
	After  []func(req *http.Request, res *http.Response, err error)
}

type HttpTraceInfo struct {
	// DNSLookupTime is a duration that transport took to perform
	// DNS lookup.
	DNSLookupTime time.Duration

	// ConnectTime is a duration that took to obtain a successful connection.
	ConnectTime time.Duration

	// TCPConnectTime is a duration that took to obtain the TCP connection.
	TCPConnectTime time.Duration

	// TLSHandshakeTime is a duration that TLS handshake took place.
	TLSHandshakeTime time.Duration

	// FirstResponseTime is a duration that server took to respond first byte since
	// connection ready (after tls handshake if it's tls and not a reused connection).
	FirstResponseTime time.Duration

	// ResponseTime is a duration since first response byte from server to
	// request completion.
	ResponseTime time.Duration

	// TotalTime is a duration that total request took end-to-end.
	TotalTime time.Duration

	// IsConnReused is whether this connection has been previously
	// used for another HTTP request.
	IsConnReused bool

	// IsConnWasIdle is whether this connection was obtained from an
	// idle pool.
	IsConnWasIdle bool

	// ConnIdleTime is a duration how long the connection was previously
	// idle, if IsConnWasIdle is true.
	ConnIdleTime time.Duration

	// RemoteAddr returns the remote network address.
	RemoteAddr net.Addr
}

func (t *clientTrace) CreateContext(c context.Context) context.Context {
	var ctx = c
	if c == nil {
		ctx = context.TODO()
	}
	return httptrace.WithClientTrace(
		ctx,
		&httptrace.ClientTrace{
			DNSStart: func(_ httptrace.DNSStartInfo) {
				t.dnsStart = time.Now()
			},
			DNSDone: func(_ httptrace.DNSDoneInfo) {
				t.dnsDone = time.Now()
			},
			ConnectStart: func(_, _ string) {
				if t.dnsDone.IsZero() {
					t.dnsDone = time.Now()
				}
				if t.dnsStart.IsZero() {
					t.dnsStart = t.dnsDone
				}
			},
			ConnectDone: func(net, addr string, err error) {
				t.connectDone = time.Now()
			},
			GetConn: func(_ string) {
				t.getConn = time.Now()
			},
			GotConn: func(ci httptrace.GotConnInfo) {
				t.gotConn = time.Now()
				t.gotConnInfo = ci
			},
			GotFirstResponseByte: func() {
				t.gotFirstResponseByte = time.Now()
			},
			TLSHandshakeStart: func() {
				t.tlsHandshakeStart = time.Now()
			},
			TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
				t.tlsHandshakeDone = time.Now()
			},
		},
	)
}
