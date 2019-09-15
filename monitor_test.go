package getcached

import (
	"math/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mikegleasonjr/getcached/mocks"
)

func TestStats(t *testing.T) {
	cache := new(mocks.Cache)
	defer cache.AssertExpectations(t)

	mon := NewMonitor(cache)
	want := &Stats{}

	cache.On("Get", "hit10").Once().Return(randBytes(10), true)
	want.Gets++
	want.Hits++
	want.HitsBytes += 10
	mon.Get("hit10")

	cache.On("Get", "miss").Once().Return(nil, false)
	want.Gets++
	want.Misses++
	mon.Get("miss")

	b := randBytes(20)
	cache.On("Set", "set20", b).Once()
	want.Sets++
	want.SetsBytes += 20
	mon.Set("set20", b)

	cache.On("Delete", "del").Once()
	want.Deletes++
	mon.Delete("del")

	got := mon.Stats()
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("stats mismatch (-want +got):\n%s", diff)
	}
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
