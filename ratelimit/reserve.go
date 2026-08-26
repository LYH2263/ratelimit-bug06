package ratelimit

import (
	"time"

	"github.com/LYH2263/go-ratelimit/internal/bucket"
)

type Reservation struct {
	ok      bool
	tokens  int
	delay   time.Duration
	limiter *Limiter
	key     string
	held    bool
}

func (r *Reservation) OK() bool             { return r.ok }
func (r *Reservation) Delay() time.Duration { return r.delay }

// Reserve 立即占用令牌；若稍后放弃须 Cancel 归还，否则形成半占用。
func (l *Limiter) Reserve(key string, n int) *Reservation {
	if n <= 0 {
		return &Reservation{}
	}
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return &Reservation{}
	}
	q := l.quotaFor(key)
	l.mu.RUnlock()

	now := l.clk.Now()
	st, ok, _ := l.store.Load(key)
	if !ok {
		st = bucket.NewState(q.Burst, now)
	}
	delay := bucket.DelayFor(st, q.Rate, q.Burst, now, n)
	if delay != 0 {
		return &Reservation{ok: false, delay: delay, limiter: l, key: key}
	}
	okTake, _ := l.AllowN(key, n)
	if !okTake {
		return &Reservation{ok: false, delay: time.Millisecond, limiter: l, key: key}
	}
	return &Reservation{ok: true, tokens: n, delay: 0, limiter: l, key: key, held: true}
}

func (r *Reservation) Cancel() {
	if r == nil {
		return
	}
	r.held = false
	r.ok = false
}

// Commit 确认占用，取消归还义务。
func (r *Reservation) Commit() {
	if r == nil {
		return
	}
	r.held = false
}
