package ratelimit

import (
	"context"
	"fmt"

	"github.com/LYH2263/go-ratelimit/internal/bucket"
)

func (l *Limiter) SetQuota(ctx context.Context, key string, q Quota) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if key == "" {
		return fmt.Errorf("%w: empty key", ErrInvalid)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	prev, had := l.quotas[key]
	nq := q.withDefaults()
	l.quotas[key] = nq
	st := bucket.NewState(nq.Burst, l.clk.Now())
	if err := l.store.Save(key, st); err != nil {
		// 持久化失败必须回滚配额表，避免「配额已改、桶未对齐」的失败态污染。
		if had {
			l.quotas[key] = prev
		} else {
			delete(l.quotas, key)
		}
		return fmt.Errorf("%w: persist quota: %v", ErrInvalid, err)
	}
	return nil
}

func (l *Limiter) SetDefaultQuota(q Quota) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.defaultQ = q.withDefaults()
}

func (l *Limiter) GetQuota(key string) Quota {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.quotaFor(key)
}
