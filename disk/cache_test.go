package disk

import (
	"bytes"
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

func BenchmarkCache(b *testing.B) {
	dir, err := ioutil.TempDir("", "diskcache")
	if err != nil {
		b.Fatalf("unexpected error: %q", err)
	}
	defer os.RemoveAll(dir)
	c := New(WithDir(dir))

	key := "key"
	content := []byte("hello")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			c.Set(key, content)
			got, ok := c.Get(key)
			if ok && !bytes.Equal(got, content) {
				b.Fatalf("unexpected content: got %s, want %s", got, content)
			}
			c.Delete(key)
		}
	})
}
