package lru

import (
	"container/list"
	"math"
	"sync"

	"github.com/gregjones/httpcache"
)

const defaultSize = 25 << 20 // 25MB

// Cache is an LRU cache. It is safe for concurrent access.
// It itself uses a cache for its underlying storage.
type Cache struct {
	c     httpcache.Cache
	mu    sync.Mutex
	cap   int64
	items map[string]*item
	list  *list.List
}

type item struct {
	key     string
	size    uint64
	element *list.Element
}

// New creates a new Cache with c as its
// underlying storage and a capacity of cap bytes.
func New(options ...func(*Cache)) httpcache.Cache {
	c := &Cache{
		c:     defaultCache(),
		cap:   defaultSize,
		items: make(map[string]*item),
		list:  list.New(),
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// Get looks up a key's value from the cache and refreshes it.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	c.mu.Lock()
	item, ok := c.items[key]
	if !ok {
		c.mu.Unlock()
		return
	}
	c.list.MoveToFront(item.element)
	c.mu.Unlock()
	return c.c.Get(key)
}

// Set adds or refreshes a value in the cache.
func (c *Cache) Set(key string, resp []byte) {
	victims := []string{} // to prevent lock contention of slow storage
	var added uint64      // bytes added to cache (can be negative)

	c.mu.Lock()
	if itm, exists := c.items[key]; exists {
		c.list.MoveToFront(itm.element)
		added = uint64(len(resp)) - itm.size
		itm.size = uint64(len(resp))
	} else {
		itm := &item{key: key, size: uint64(len(resp))}
		itm.element = c.list.PushFront(itm)
		c.items[key] = itm
		added = uint64(itm.size)
	}
	c.cap -= int64(added)
	for c.cap < 0 && c.list.Len() > 1 {
		itm := c.list.Back().Value.(*item)
		victims = append(victims, itm.key)
		c.purge(itm)
	}
	c.mu.Unlock()

	for _, key := range victims {
		c.c.Delete(key)
	}
	c.c.Set(key, resp)
}

// Delete removes the provided key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	if item, exists := c.items[key]; exists {
		c.purge(item)
	}
	c.mu.Unlock()

	c.c.Delete(key)
}

func (c *Cache) purge(item *item) {
	delete(c.items, item.key)
	c.list.Remove(item.element)
	c.cap += int64(item.size)
}

// WithCache configures a Cache to use a specific
// httpcache.Cache.
func WithCache(hc httpcache.Cache) func(*Cache) {
	return func(c *Cache) {
		c.c = hc
	}
}

// WithSize configures a Cache to use a specific
// capacity (in bytes).
func WithSize(size uint64) func(*Cache) {
	if size >= math.MaxInt64 {
		panic("size must fit an int64")
	}
	return func(c *Cache) {
		c.cap = int64(size)
	}
}

func defaultCache() httpcache.Cache {
	return httpcache.NewMemoryCache()
}
