package getcached

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mikegleasonjr/getcached/mocks"
	"github.com/stretchr/testify/mock"
)

func TestClient(t *testing.T) {
	picker := new(mocks.Picker)
	defer picker.AssertExpectations(t)

	transport := new(mocks.RoundTripper)
	defer transport.AssertExpectations(t)

	response := new(http.Response)

	picker.
		On("Pick", "http://origin.net/resource").
		Once().
		Return("http://proxy.local/handler")

	transport.
		On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://proxy.local/handler?q="+url.QueryEscape("http://origin.net/resource") &&
				req.Host == "proxy.local" &&
				req.Method == "POST"
		})).
		Once().
		Return(response, nil)

	c := NewClient(WithPicker(picker), WithClientTransport(transport))

	req := httptest.NewRequest("POST", "http://origin.net/resource", nil)
	res, err := c.RoundTrip(req)

	if err != nil {
		t.Errorf("unexpected error: %q", err)
	}

	if got, want := res, response; got != want {
		t.Errorf("unexpected response: got %#v, want %#v", got, want)
	}
}
