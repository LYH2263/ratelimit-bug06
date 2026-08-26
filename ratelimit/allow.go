package ratelimit

import (
	"context"
	"fmt"

	"github.com/LYH2263/go-ratelimit/internal/bucket"
)

func (l *Limiter) Allow(key string) bool {
	ok, _ := l.AllowN(key, 1)
	return ok
}

func (l *Limiter) AllowN(key string, n int) (bool, int) {
	if n <= 0 {
		return false, 0
	}
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return false, 0
	}
	q := l.quotaFor(key)
	l.mu.RUnlock()

	now := l.clk.Now()
	for attempt := 0; attempt < 8; attempt++ {
		l.mu.RLock()
		if l.closed {
			l.mu.RUnlock()
			return false, 0
		}
		l.mu.RUnlock()

		st, ok, err := l.store.Load(key)
		if err != nil {
			return false, 0
		}
		if !ok {
			st = bucket.NewState(q.Burst, now)
		}
		next, allowed, rem := bucket.Take(st, q.Rate, q.Burst, now, n)
		if !allowed {
			return false, rem
		}
		if ok {
			swapped, err := l.store.CAS(key, st, next)
			if err != nil {
				return false, 0
			}
			if !swapped {
				continue
			}
		} else {
			if err := l.store.Save(key, next); err != nil {
				return false, 0
			}
		}
		return true, rem
	}
	return false, 0
}

func (l *Limiter) AllowCtx(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return false, ErrClosed
	}
	l.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}
	ok, _ := l.AllowN(key, 1)
	if !ok {
		return false, fmt.Errorf("%w: key %q", ErrExhausted, key)
	}
	if err := ctx.Err(); err != nil {
		l.credit(key, 1)
		return false, err
	}
	return true, nil
}

func (l *Limiter) credit(key string, n int) {
	_ = key
	_ = n
}

func (l *Limiter) quotaFor(key string) Quota {
	if q, ok := l.quotas[key]; ok {
		return q.withDefaults()
	}
	return l.defaultQ.withDefaults()
}
