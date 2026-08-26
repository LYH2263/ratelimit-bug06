package ratelimit

import (
	"context"
	"testing"
)

func TestBug06_ReserveCancelReturnsTokens(t *testing.T) {
	l, _ := New(Options{})
	defer l.Close()
	_ = l.SetQuota(context.Background(), "k", Quota{Rate: 1, Burst: 1})
	r := l.Reserve("k", 1)
	if !r.OK() {
		t.Fatal("reserve")
	}
	r.Cancel()
	if !l.Allow("k") {
		t.Fatal("tokens not returned after Cancel")
	}
}
