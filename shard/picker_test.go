package shard
import (
	"strings"
	"testing"

	"github.com/mikegleasonjr/getcached/mocks"
	"github.com/stretchr/testify/mock"
)

func TestPick(t *testing.T) {
	testCases := []struct {
		desc      string
		proxies   []string
		hashMocks map[string]uint32
		want      map[string]string
	}{
		{
			desc:    "case1",
			proxies: []string{"http://p1.com", "http://p2.com/handler"},
			hashMocks: map[string]uint32{
				"http://p1.com":                    0,
				"http://p2.com/handler":            1,
				"http://some.url/res-proxy1.js":    0,
				"http://some.url/res-proxy2.js":    1,
				"http://another.url/res-proxy1.js": 0,
				"http://another.url/res-proxy2.js": 1,
			},
			want: map[string]string{
				"http://some.url/res-proxy1.js":    "http://p1.com",
				"http://some.url/res-proxy2.js":    "http://p2.com/handler",
				"http://another.url/res-proxy1.js": "http://p1.com",
				"http://another.url/res-proxy2.js": "http://p2.com/handler",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			hashMock := new(mocks.HashFn)
			defer hashMock.AssertExpectations(t)

			for origin, hash := range tC.hashMocks {
				origin := origin
				hashMock.On("Fn", mock.MatchedBy(func(data []byte) bool {
					return strings.HasSuffix(string(data), origin)
				})).Return(hash)
			}

			p := New(WithHashFn(hashMock.Fn))
			p.Set(tC.proxies...)

			for origin, want := range tC.want {
				got := p.Pick(origin)
				if got != want {
					t.Errorf("unexpected proxy picked: got %q, want %q", got, want)
				}
			}
		})
	}
}
