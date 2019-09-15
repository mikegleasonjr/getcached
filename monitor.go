package getcached

import (
	"sync/atomic"

	"github.com/gregjones/httpcache"
)

// Stats are cache statistics returned from a Monitor.
type Stats struct {
	Gets      int64 // total gets
	Hits      int64 // cache hits
	HitsBytes int64 // cache hits (in bytes)
	Misses    int64 // cache misses
	Sets      int64 // total sets
	SetsBytes int64 // total sets (in bytes)
	Deletes   int64 // total deletes
}

// Monitor is a cache decorator which keeps tracks
// of various statistics about a Cache. Monitor itself
// implements httpcache.Cache so it can take the place
// of the monitored cache.
type Monitor struct {
	c         httpcache.Cache
	gets      AtomicInt
	hits      AtomicInt
	hitsBytes AtomicInt
	misses    AtomicInt
	sets      AtomicInt
	setsBytes AtomicInt
	deletes   AtomicInt
}

// NewMonitor creates a Monitor.
func NewMonitor(c httpcache.Cache) *Monitor {
	return &Monitor{c: c}
}

// Stats returns the current cache stats.
func (m *Monitor) Stats() *Stats {
	return &Stats{
		Gets:      m.gets.Get(),
		Hits:      m.hits.Get(),
		HitsBytes: m.hitsBytes.Get(),
		Misses:    m.misses.Get(),
		Sets:      m.sets.Get(),
		SetsBytes: m.setsBytes.Get(),
		Deletes:   m.deletes.Get(),
	}
}

// Get implements httpcache.Cache.
func (m *Monitor) Get(key string) ([]byte, bool) {
	m.gets.Add(1)

	b, hit := m.c.Get(key)
	if hit {
		m.hits.Add(1)
		m.hitsBytes.Add(int64(len(b)))
	} else {
		m.misses.Add(1)
	}

	return b, hit
}

// Set implements httpcache.Cache.
func (m *Monitor) Set(key string, resp []byte) {
	m.sets.Add(1)
	m.setsBytes.Add(int64(len(resp)))

	m.c.Set(key, resp)
}

// Delete implements httpcache.Cache.
func (m *Monitor) Delete(key string) {
	m.deletes.Add(1)

	m.c.Delete(key)
}

// An AtomicInt is an int64 to be accessed atomically.
type AtomicInt int64

// Add atomically adds n to i.
func (i *AtomicInt) Add(n int64) {
	atomic.AddInt64((*int64)(i), n)
}

// Get atomically gets the value of i.
func (i *AtomicInt) Get() int64 {
	return atomic.LoadInt64((*int64)(i))
}
