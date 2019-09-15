package getcached

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/mikegleasonjr/getcached/mocks"
	"github.com/stretchr/testify/mock"
)

func TestProxyCacheMiss(t *testing.T) {
	transport := new(mocks.RoundTripper)
	defer transport.AssertExpectations(t)

	response := new(http.Response)
	response.Body = ioutil.NopCloser(strings.NewReader("content"))

	transport.
		On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://origin.net/resource" &&
				req.Host == "origin.net" &&
				req.Method == "GET"
		})).
		Once().
		Return(response, nil)

	p := New(WithProxyTransport(transport))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?q="+url.QueryEscape("http://origin.net/resource"), nil)
	p.ServeHTTP(rr, req)

	if got, want := rr.Code, response.StatusCode; got != want {
		t.Errorf("unexpected status code: got %d, want %d", got, want)
	}

	if got, want := rr.HeaderMap.Get(httpcache.XFromCache), ""; got != want {
		t.Errorf("unexpected %q header: got %q, want %q", httpcache.XFromCache, got, want)
	}
}

func TestProxyCacheHit(t *testing.T) {
	cache := new(mocks.Cache)
	defer cache.AssertExpectations(t)

	response := new(http.Response)
	response.Body = ioutil.NopCloser(strings.NewReader("content"))
	response.Header = http.Header{
		"date":    []string{time.Now().Format(time.RFC1123)},
		"Expires": []string{time.Now().Add(time.Hour).Format(time.RFC1123)},
	}
	b, _ := httputil.DumpResponse(response, true)

	cache.
		On("Get", "http://origin.net/resource").
		Once().
		Return(b, true)

	p := New(WithCache(cache))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?q="+url.QueryEscape("http://origin.net/resource"), nil)

	p.ServeHTTP(rr, req)

	if got, want := rr.Code, response.StatusCode; got != want {
		t.Errorf("unexpected status code: got %d, want %d", got, want)
	}

	if got, want := rr.HeaderMap.Get(httpcache.XFromCache), "1"; got != want {
		t.Errorf("unexpected %q header: got %q, want %q", httpcache.XFromCache, got, want)
	}
}
