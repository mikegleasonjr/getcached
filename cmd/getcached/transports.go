package main

import (
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

var (
	// ErrResponseSize tells the response body exceeds a certain size.
	ErrResponseSize = errors.New("response size exceeded")
)

// DefaultTransport is a default http.Roundtripper
// with sensible defaults.
func DefaultTransport() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// BodySizeCheckerTransport replaces the response body with a
// ReadCloser that returns an error when the body exceeds a certain
// size.
func BodySizeCheckerTransport(max int64, rt http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		res, err := rt.RoundTrip(req)
		if res != nil && res.Body != nil {
			res.Body = &bodySizeChecker{RC: res.Body, N: max}
		}
		return res, err
	})
}

type bodySizeChecker struct {
	RC io.ReadCloser // underlying reader
	N  int64         // max bytes remaining
}

func (f *bodySizeChecker) Read(p []byte) (n int, err error) {
	if f.N <= 0 {
		return 0, ErrResponseSize
	}
	if int64(len(p)) > f.N {
		p = p[0:f.N]
	}
	n, err = f.RC.Read(p)
	f.N -= int64(n)
	return
}

func (f *bodySizeChecker) Close() error {
	return f.RC.Close()
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
