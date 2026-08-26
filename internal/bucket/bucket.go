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
