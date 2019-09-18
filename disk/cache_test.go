package disk

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gregjones/httpcache/test"
)

func TestCache(t *testing.T) {
	dir, err := ioutil.TempDir("", "diskcache")
	if err != nil {
		t.Fatalf("unexpected error: %q", err)
	}
	defer os.RemoveAll(dir)
	c, err := New(WithDir(dir))
	if err != nil {
		t.Fatalf("unexpected error: %q", err)
	}
	test.Cache(t, c)
}
