package syncstate

import "time"

type State struct {
	Tokens   float64
	LastFill time.Time
	Version  uint64
}

func Clone(s State) State { return s }

func Bump(s State) State {
	s.Version++
	return s
}
