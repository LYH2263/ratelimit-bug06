package bucket

import (
	"time"

	"github.com/LYH2263/go-ratelimit/internal/syncstate"
)

func Peek(st syncstate.State, rate float64, burst int, now time.Time) syncstate.State {
	return refill(st, rate, burst, now)
}

func SecondsUntil(st syncstate.State, rate float64, burst int, now time.Time, target float64) float64 {
	st = refill(st, rate, burst, now)
	if st.Tokens >= target {
		return 0
	}
	need := target - st.Tokens
	if rate <= 0 {
		return -1
	}
	return need / rate
}
