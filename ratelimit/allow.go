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

// credit 归还 n 个令牌（Take 的逆操作），供 Cancel 归还半占用令牌、
// 及 AllowCtx 在 ctx 取消后回滚已占用令牌。与 AllowN 同样走 Load/CAS 重试，
// 归还量封顶 burst，避免并发下归还超过容量。
func (l *Limiter) credit(key string, n int) {
	if n <= 0 {
		return
	}
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return
	}
	q := l.quotaFor(key)
	l.mu.RUnlock()

	now := l.clk.Now()
	for attempt := 0; attempt < 8; attempt++ {
		l.mu.RLock()
		if l.closed {
			l.mu.RUnlock()
			return
		}
		l.mu.RUnlock()

		st, ok, err := l.store.Load(key)
		if err != nil {
			return
		}
		if !ok {
			// 桶已不存在：归还的令牌无处安放，直接新建一个满桶即可，
			// 等价于不归还（反正容量已重置），这里保留 Save 以维持状态存在性。
			st = bucket.NewState(q.Burst, now)
		}
		next, rem := bucket.Put(st, q.Rate, q.Burst, now, n)
		_ = rem
		if ok {
			swapped, err := l.store.CAS(key, st, next)
			if err != nil {
				return
			}
			if !swapped {
				continue
			}
		} else {
			if err := l.store.Save(key, next); err != nil {
				return
			}
		}
		return
	}
}

func (l *Limiter) quotaFor(key string) Quota {
	if q, ok := l.quotas[key]; ok {
		return q.withDefaults()
	}
	return l.defaultQ.withDefaults()
}
