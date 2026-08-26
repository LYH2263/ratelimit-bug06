package shard

import (
	"sync"

	"github.com/LYH2263/go-ratelimit/internal/syncstate"
)

type bucket struct {
	mu    sync.RWMutex
	items map[string]any
}

type Map struct {
	shards []*bucket
}

func NewMap(n int) *Map {
	if n < 1 {
		n = 1
	}
	m := &Map{shards: make([]*bucket, n)}
	for i := range m.shards {
		m.shards[i] = &bucket{items: make(map[string]any)}
	}
	return m
}

func (m *Map) shardFor(key string) *bucket {
	return m.shards[hashKey(key, len(m.shards))]
}

func (m *Map) Get(key string) (any, bool) {
	b := m.shardFor(key)
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.items[key]
	return v, ok
}

func (m *Map) Set(key string, v any) {
	b := m.shardFor(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items[key] = v
}

func (m *Map) Delete(key string) {
	b := m.shardFor(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.items, key)
}

func (m *Map) CAS(key string, expect, next syncstate.State) (bool, error) {
	b := m.shardFor(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	cur, ok := b.items[key]
	if !ok {
		if expect.Version != 0 {
			return false, nil
		}
		b.items[key] = next
		return true, nil
	}
	st := cur.(syncstate.State)
	if st.Version != expect.Version {
		return false, nil
	}
	b.items[key] = next
	return true, nil
}

func (m *Map) Keys() []string {
	out := make([]string, 0)
	for _, b := range m.shards {
		b.mu.RLock()
		for k := range b.items {
			out = append(out, k)
		}
		b.mu.RUnlock()
	}
	return out
}

func (m *Map) ShardCount() int { return len(m.shards) }

func (m *Map) Clear() {
	for _, b := range m.shards {
		b.mu.Lock()
		b.items = make(map[string]any)
		b.mu.Unlock()
	}
}
