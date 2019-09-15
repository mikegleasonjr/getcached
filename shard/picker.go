package shard

import (
	"hash/fnv"

	"github.com/mikegleasonjr/getcached/shard/consistenthash"
)

const defaultReplicas = 100

var defaultHashFn = fnv32a

// Shard picks servers from a Consistent Hash Ring.
type Shard struct {
	replicas int
	hashFn   consistenthash.Hash
	hashMap  *consistenthash.Map
}

// New creates a Shard.
func New(options ...func(*Shard)) *Shard {
	p := &Shard{
		replicas: defaultReplicas,
		hashFn:   defaultHashFn,
	}

	for _, option := range options {
		option(p)
	}

	return p
}

// Pick implements getcached.Picker.
func (p *Shard) Pick(origin string) string {
	if p.hashMap == nil {
		return ""
	}
	return p.hashMap.Get(origin)
}

// Set implements getcached.Picker.
func (p *Shard) Set(proxies ...string) {
	p.hashMap = consistenthash.New(p.replicas, p.hashFn)
	p.hashMap.Add(proxies...)
}

// WithReplicas set the number of replicas of the consistent hash ring.
func WithReplicas(replicas int) func(*Shard) {
	return func(p *Shard) {
		p.replicas = replicas
	}
}

// WithHashFn set the hash funtion of the consistent hash ring.
func WithHashFn(hashFn consistenthash.Hash) func(*Shard) {
	return func(p *Shard) {
		p.hashFn = hashFn
	}
}

func fnv32a(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}
