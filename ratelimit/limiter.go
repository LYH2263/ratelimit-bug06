package ratelimit

import (
	"sync"

	"github.com/LYH2263/go-ratelimit/internal/clock"
)

type Limiter struct {
	mu       sync.RWMutex
	closed   bool
	opts     Options
	store    Store
	clk      clock.Clock
	quotas   map[string]Quota
	defaultQ Quota
	snapBuf  []KeyStat
}

func New(opts Options) (*Limiter, error) {
	opts = opts.withDefaults()
	return &Limiter{
		opts:     opts,
		store:    opts.Store,
		clk:      opts.Clock,
		quotas:   make(map[string]Quota),
		defaultQ: Quota{Rate: 10, Burst: 10},
	}, nil
}

func (l *Limiter) Store() Store { return l.store }

func (l *Limiter) Clock() clock.Clock { return l.clk }
