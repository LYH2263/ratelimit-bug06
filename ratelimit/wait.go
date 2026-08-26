package ratelimit

import (
	"context"
	"errors"
	"time"
)

func (l *Limiter) Wait(ctx context.Context, key string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		l.mu.RLock()
		closed := l.closed
		l.mu.RUnlock()
		if closed {
			return ErrClosed
		}
		ok, err := l.AllowCtx(ctx, key)
		if ok {
			return nil
		}
		if err != nil {
			if errors.Is(err, ErrClosed) {
				return err
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
}
