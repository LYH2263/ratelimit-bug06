package ratelimit

import (
	"sync"

	"github.com/LYH2263/go-ratelimit/internal/shard"
	"github.com/LYH2263/go-ratelimit/internal/syncstate"
)

type Store interface {
	Load(key string) (syncstate.State, bool, error)
	Save(key string, st syncstate.State) error
	CAS(key string, expect, next syncstate.State) (bool, error)
	Delete(key string) error
	Keys() []string
}

type MemoryStore struct {
	mu       sync.RWMutex
	items    map[string]syncstate.State
	keyCache []string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]syncstate.State)}
}

func (m *MemoryStore) Load(key string) (syncstate.State, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, ok := m.items[key]
	return st, ok, nil
}

func (m *MemoryStore) Save(key string, st syncstate.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = st
	return nil
}

func (m *MemoryStore) CAS(key string, expect, next syncstate.State) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.items[key]
	if !ok && expect.Version != 0 {
		return false, nil
	}
	if ok && !syncstate.Match(cur, expect) {
		return false, nil
	}
	m.items[key] = next
	return true, nil
}

func (m *MemoryStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
	return nil
}

func (m *MemoryStore) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.items))
	for k := range m.items {
		out = append(out, k)
	}
	return out
}

type ShardedStore struct {
	shards *shard.Map
}

func NewShardedStore(n int) *ShardedStore {
	return &ShardedStore{shards: shard.NewMap(n)}
}

func (s *ShardedStore) Load(key string) (syncstate.State, bool, error) {
	v, ok := s.shards.Get(key)
	if !ok {
		return syncstate.State{}, false, nil
	}
	return v.(syncstate.State), true, nil
}

func (s *ShardedStore) Save(key string, st syncstate.State) error {
	s.shards.Set(key, st)
	return nil
}

func (s *ShardedStore) CAS(key string, expect, next syncstate.State) (bool, error) {
	return s.shards.CAS(key, expect, next)
}

func (s *ShardedStore) Delete(key string) error {
	s.shards.Delete(key)
	return nil
}

func (s *ShardedStore) Keys() []string {
	return s.shards.Keys()
}
