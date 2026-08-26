package ratelimit

import (
	"github.com/LYH2263/go-ratelimit/internal/clock"
)

type Options struct {
	Store  Store
	Clock  clock.Clock
	Shards int
}

func (o Options) withDefaults() Options {
	if o.Store == nil {
		o.Store = NewMemoryStore()
	}
	if o.Clock == nil {
		o.Clock = clock.System{}
	}
	if o.Shards <= 0 {
		o.Shards = 32
	}
	return o
}
