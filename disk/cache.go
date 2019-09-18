package disk

import (
	"github.com/prologic/bitcask"
)

const (
	defaultDir  = "/tmp"
	defaultSize = 100 << 20 // 100mb
	defaultSync = true
)

// Cache caches requests to disk.
type Cache struct {
	db *bitcask.Bitcask
}

type opts struct {
	dir  string
	size uint64
	sync bool
}

// New creates a Cache backed by diskv.
func New(options ...func(*opts)) (*Cache, error) {
	o := &opts{
		dir:  defaultDir,
		size: defaultSize,
		sync: defaultSync,
	}

	for _, option := range options {
		option(o)
	}

	db, err := bitcask.Open(
		o.dir,
		bitcask.WithMaxDatafileSize(int(o.size)),
		bitcask.WithSync(o.sync))
	if err != nil {
		return nil, err
	}

	return &Cache{db: db}, nil
}

// Get gets an item from the cache.
func (c *Cache) Get(key string) ([]byte, bool) {
	b, err := c.db.Get([]byte(key))
	if err != nil {
		return nil, false
	}
	return b, true
}

// Set saves a response to the cache as key.
func (c *Cache) Set(key string, resp []byte) {
	c.db.Put([]byte(key), resp)
}

// Delete deletes an item from the cache.
func (c *Cache) Delete(key string) {
	c.db.Delete([]byte(key))
}

// WithSize sets a cache maximum size.
func WithSize(size uint64) func(*opts) {
	return func(o *opts) {
		o.size = size
	}
}

// WithDir sets a cache directory.
func WithDir(dir string) func(*opts) {
	return func(o *opts) {
		o.dir = dir
	}
}

// WithSync makes the cache sync every writes.
func WithSync(sync bool) func(*opts) {
	return func(o *opts) {
		o.sync = sync
	}
}
