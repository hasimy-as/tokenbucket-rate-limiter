package service

import (
	"time"
)

type TokenBucket struct {
	tokens     int
	maxTokens  int
	refillRate int // tokens per second
	lastRefill time.Time
}

func NewTokenBucket(maxTokens, refillRate int) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	newTokens := int(elapsed) * tb.refillRate

	if newTokens > 0 {
		tb.tokens = min(tb.tokens+newTokens, tb.maxTokens)
		tb.lastRefill = now
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
