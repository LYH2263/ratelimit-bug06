package ratelimit

import (
	"fmt"

	"github.com/LYH2263/go-ratelimit/internal/bucket"
)

func (l *Limiter) Reset(key string) error {
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return ErrClosed
	}
	q := l.quotaFor(key)
	l.mu.RUnlock()
	st := bucket.NewState(q.Burst, l.clk.Now())
	return l.store.Save(key, st)
}

func (l *Limiter) ResetAll() error {
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return ErrClosed
	}
	l.mu.RUnlock()
	keys := append([]string(nil), l.store.Keys()...)
	for i, k := range keys {
		if err := l.store.Delete(k); err != nil {
			return fmt.Errorf("ratelimit: reset-all aborted after %d keys: %w", i, err)
		}
	}
	return nil
}
