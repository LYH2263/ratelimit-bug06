package ratelimit

import "errors"

var (
	ErrClosed      = errors.New("ratelimit: limiter closed")
	ErrNotFound    = errors.New("ratelimit: key not found")
	ErrInvalid     = errors.New("ratelimit: invalid argument")
	ErrExhausted   = errors.New("ratelimit: quota exhausted")
	ErrCASConflict = errors.New("ratelimit: store cas conflict")
)
