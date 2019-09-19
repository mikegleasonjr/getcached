package getcached

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gregjones/httpcache"
)

type key struct{}

var originKey = key{}

// Proxy is a caching proxy server.
type Proxy struct {
	rp *httputil.ReverseProxy
	tr *httpcache.Transport
}

// New creates a Proxy using options.
func New(options ...func(*Proxy)) *Proxy {
	tr := httpcache.NewTransport(httpcache.NewMemoryCache())

	p := &Proxy{
		tr: tr,
		rp: &httputil.ReverseProxy{
			Transport: tr,
			Director: func(req *http.Request) {
				origin := req.Context().Value(originKey).(*url.URL)
				req.URL = origin
				req.Host = origin.Host
			},
		},
	}

	for _, option := range options {
		option(p)
	}

	return p
}

// ServeHTTP enables Proxy to be used as an http.Handler.
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	q := req.URL.Query().Get("q")
	if q == "" {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	origin, err := url.Parse(q)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	ctx := context.WithValue(req.Context(), originKey, origin)
	p.rp.ServeHTTP(rw, req.WithContext(ctx))
}

// WithProxyTransport configures a Proxy to use
// a specific http.RoundTripper.
func WithProxyTransport(tr http.RoundTripper) func(*Proxy) {
	return func(p *Proxy) {
		p.tr.Transport = tr
	}
}

// WithCache configures a Proxy to use a specific
// httpcache.Cache.
func WithCache(c httpcache.Cache) func(*Proxy) {
	return func(p *Proxy) {
		p.tr.Cache = c
	}
}

// WithErrorLogger configures a Proxy to use an error logger.
func WithErrorLogger(l *log.Logger) func(*Proxy) {
	return func(p *Proxy) {
		p.rp.ErrorLog = l
	}
}

// WithBufferPool configures a Proxy to use a BufferPool.
func WithBufferPool(pool httputil.BufferPool) func(*Proxy) {
	return func(p *Proxy) {
		p.rp.BufferPool = pool
	}
}
