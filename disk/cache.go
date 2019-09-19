package disk

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

const (
	defaultDir = "/tmp"
)

// Cache caches requests to disk.
type Cache struct {
	dir   string
	mu    sync.RWMutex // guards mus
	locks map[string]*lock
}

type lock struct {
	sync.RWMutex
	key  string
	refs int
}

// New creates a Cache backed by a directory.
// Panics is directory does not exists.
func New(options ...func(*Cache)) *Cache {
	c := &Cache{dir: defaultDir, locks: map[string]*lock{}}

	for _, option := range options {
		option(c)
	}

	if s, err := os.Stat(c.dir); os.IsNotExist(err) || !s.IsDir() {
		panic(fmt.Sprintf("%q does not exists", c.dir))
	}

	return c
}

// Get gets an item from the cache.
func (c *Cache) Get(key string) ([]byte, bool) {
	fullpath := c.fullPath(key)

	l := c.getLock(key)
	defer c.releaseLock(l)

	l.RLock()
	defer l.RUnlock()

	b, err := ioutil.ReadFile(fullpath)
	return b, err == nil
}

// Set saves a response to the cache as key.
func (c *Cache) Set(key string, resp []byte) {
	fullpath := c.fullPath(key)

	l := c.getLock(key)
	defer c.releaseLock(l)

	l.Lock()
	defer l.Unlock()

	err := ioutil.WriteFile(fullpath, resp, 0644)
	if err != nil {
		os.Remove(fullpath)
	}
}

// Delete deletes an item from the cache.
func (c *Cache) Delete(key string) {
	fullpath := c.fullPath(key)

	l := c.getLock(key)
	defer c.releaseLock(l)

	l.Lock()
	defer l.Unlock()

	os.Remove(fullpath)
}

func (c *Cache) fullPath(key string) string {
	h := md5.New()
	h.Write([]byte(key))
	filename := hex.EncodeToString(h.Sum(nil)) + ".cache"
	return path.Join(c.dir, filename)
}

func (c *Cache) getLock(key string) *lock {
	c.mu.Lock()
	defer c.mu.Unlock()

	l, ok := c.locks[key]
	if !ok {
		l = &lock{key: key}
		c.locks[key] = l
	}
	l.refs++

	return l
}

func (c *Cache) releaseLock(l *lock) {
	c.mu.Lock()
	defer c.mu.Unlock()

	l.refs--
	if l.refs == 0 {
		delete(c.locks, l.key)
	}
}

// WithDir sets a cache directory.
func WithDir(dir string) func(*Cache) {
	return func(c *Cache) {
		c.dir = dir
	}
}
