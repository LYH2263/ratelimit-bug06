package ratelimit

import "time"

// Quota defines sustained rate and burst capacity for one key.
type Quota struct {
	Rate  float64
	Burst int
	Every time.Duration
}

func (q Quota) withDefaults() Quota {
	if q.Rate <= 0 {
		q.Rate = 10
	}
	if q.Burst <= 0 {
		q.Burst = int(q.Rate)
		if q.Burst < 1 {
			q.Burst = 1
		}
	}
	if q.Every <= 0 {
		q.Every = time.Second
	}
	return q
}

func (q Quota) TokensPerInterval() float64 {
	q = q.withDefaults()
	if q.Every == time.Second {
		return q.Rate
	}
	return q.Rate * q.Every.Seconds()
}
