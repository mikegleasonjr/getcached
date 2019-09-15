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

	test.Cache(t, New(WithDir(dir)))
}
