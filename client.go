package getcached

import (
	"errors"
	"net/http"
	"net/url"
	"sync"

	"github.com/mikegleasonjr/getcached/shard"
)

var (
	// ErrNoProxies is returned when no
	// proxies are defined.
	ErrNoProxies = errors.New("no proxies")
)

// Client is a client of a list of proxies.
type Client struct {
	transport http.RoundTripper
	mu        sync.RWMutex // guards picker ops
	picker    Picker
}

// NewClient creates a Client.
func NewClient(options ...func(*Client)) *Client {
	c := &Client{
		picker:    shard.New(),
		transport: http.DefaultTransport,
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// Set sets the list of proxies the Client
// can reach.
func (c *Client) Set(proxies ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.picker.Set(proxies...)
}

// RoundTrip makes Client a RoundTripper so
// it can be used as a Transport. This is where
// the proxy is chosen and handed the request
// according to the origin requested.
func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	origin := req.URL.String()

	c.mu.RLock()
	chosen := c.picker.Pick(origin)
	c.mu.RUnlock()

	if chosen == "" {
		return nil, ErrNoProxies
	}

	proxy, err := url.Parse(chosen)
	if err != nil {
		return nil, err
	}

	proxy.RawQuery = "q=" + url.QueryEscape(origin)

	cpy := clone(req) // per RoundTripper contract
	cpy.URL = proxy
	cpy.Host = proxy.Host

	return c.transport.RoundTrip(cpy)
}

// WithPicker configures a Client to use
// a specific Picker.
func WithPicker(p Picker) func(*Client) {
	return func(c *Client) {
		c.picker = p
	}
}

// WithClientTransport configures a Client
// to use a specific http.RoundTripper.
func WithClientTransport(tr http.RoundTripper) func(*Client) {
	return func(c *Client) {
		c.transport = tr
	}
}

// clones a request, credits goes to:
// https://github.com/golang/oauth2/blob/master/transport.go#L36
func clone(r *http.Request) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
