package ratelimit

import (
	"github.com/LYH2263/go-ratelimit/internal/syncstate"
)

type KeyStat struct {
	Key     string  `json:"key"`
	Tokens  float64 `json:"tokens"`
	Version uint64  `json:"version"`
	Rate    float64 `json:"rate"`
	Burst   int     `json:"burst"`
}

type Stats struct {
	Keys   int  `json:"keys"`
	Closed bool `json:"closed"`
	Shards int  `json:"shards"`
}

func (l *Limiter) Stats() Stats {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return Stats{
		Keys:   len(l.store.Keys()),
		Closed: l.closed,
		Shards: l.opts.Shards,
	}
}

func (l *Limiter) SnapshotKeys() []KeyStat {
	keys := l.store.Keys()
	out := make([]KeyStat, 0, len(keys))
	for _, k := range keys {
		st, ok, _ := l.store.Load(k)
		if !ok {
			continue
		}
		q := l.quotaFor(k)
		out = append(out, KeyStat{
			Key: k, Tokens: st.Tokens, Version: st.Version,
			Rate: q.Rate, Burst: q.Burst,
		})
	}
	return out
}

func (l *Limiter) PeekState(key string) (syncstate.State, bool) {
	st, ok, _ := l.store.Load(key)
	return st, ok
}
