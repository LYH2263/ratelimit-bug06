package clock

import "time"

type Clock interface {
	Now() time.Time
}

type System struct{}

func (System) Now() time.Time { return time.Now() }

type Fixed struct{ T time.Time }

func (f Fixed) Now() time.Time { return f.T }

type Step struct {
	T time.Time
	D time.Duration
}

func (s *Step) Now() time.Time {
	t := s.T
	s.T = s.T.Add(s.D)
	return t
}
