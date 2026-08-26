package ratelimit

func (l *Limiter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	return nil
}

func (l *Limiter) CloseAndFlushCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := len(l.store.Keys())
	l.closed = true
	return n
}
