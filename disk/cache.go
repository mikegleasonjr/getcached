package disk

import (
	"crypto/md5"
	"encoding/hex"
	"io"

	"github.com/peterbourgon/diskv"
)

// TODO, use fixed-key length (MD5?)
// https://github.com/peterbourgon/diskv/issues/40

// Cache caches requests to disk.
type Cache struct {
	d *diskv.Diskv
}

// New creates a Cache backed by diskv.
func New(options ...func(*diskv.Options)) *Cache {
	opts := &diskv.Options{}

	for _, option := range options {
		option(opts)
	}

	return &Cache{
		d: diskv.New(*opts),
	}
}

// Get gets an item from the cache.
func (c *Cache) Get(key string) ([]byte, bool) {
	key = fixedLengthKey(key)

	b, err := c.d.Read(key)
	if err != nil {
		return nil, false
	}

	return b, true
}

// Set saves a response to the cache as key.
func (c *Cache) Set(key string, resp []byte) {
	key = fixedLengthKey(key)
	c.d.Write(key, resp)
}

// Delete deletes an item from the cache.
func (c *Cache) Delete(key string) {
	key = fixedLengthKey(key)
	c.d.Erase(key)
}

// WithSize sets a cache maximum size.
func WithSize(size uint64) func(*diskv.Options) {
	return func(o *diskv.Options) {
		o.CacheSizeMax = size
	}
}

// WithDir sets a cache directory.
func WithDir(dir string) func(*diskv.Options) {
	return func(o *diskv.Options) {
		o.BasePath = dir
	}
}

func fixedLengthKey(key string) string {
	h := md5.New()
	io.WriteString(h, key)
	return hex.EncodeToString(h.Sum(nil))
}
