package bucket

import (
	"math"
	"time"

	"github.com/LYH2263/go-ratelimit/internal/syncstate"
)

func NewState(burst int, now time.Time) syncstate.State {
	if burst < 1 {
		burst = 1
	}
	return syncstate.State{
		Tokens:   float64(burst),
		LastFill: now,
		Version:  1,
	}
}

func refill(st syncstate.State, rate float64, burst int, now time.Time) syncstate.State {
	if burst < 1 {
		burst = 1
	}
	if rate <= 0 {
		rate = 10
	}
	if now.Before(st.LastFill) {
		return st
	}
	elapsed := now.Sub(st.LastFill).Seconds()
	add := elapsed * rate
	st.Tokens = math.Min(float64(burst), st.Tokens+add)
	st.LastFill = now
	return st
}

func Take(st syncstate.State, rate float64, burst int, now time.Time, n int) (syncstate.State, bool, int) {
	st = refill(st, rate, burst, now)
	if st.Tokens < float64(n) {
		return st, false, int(st.Tokens)
	}
	st.Tokens -= float64(n)
	st = syncstate.Bump(st)
	rem := int(st.Tokens)
	return st, true, rem
}

// Put 归还 n 个令牌到桶（Take 的逆操作），封顶 burst，供 Cancel/credit 归还半占用令牌。
func Put(st syncstate.State, rate float64, burst int, now time.Time, n int) (syncstate.State, int) {
	st = refill(st, rate, burst, now)
	if n < 0 {
		n = 0
	}
	st.Tokens = math.Min(float64(burst), st.Tokens+float64(n))
	st = syncstate.Bump(st)
	return st, int(st.Tokens)
}

func DelayFor(st syncstate.State, rate float64, burst int, now time.Time, n int) time.Duration {
	st = refill(st, rate, burst, now)
	if st.Tokens >= float64(n) {
		return 0
	}
	need := float64(n) - st.Tokens
	if rate <= 0 {
		return -1
	}
	sec := need / rate
	return time.Duration(sec * float64(time.Second))
}
