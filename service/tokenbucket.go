package service

import (
	"log"
	"sync"
	"time"
)

type TokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mutex      sync.Mutex
	logging    bool // optional logging
}

func NewTokenBucket(maxTokens, refillRate int, logging bool) *TokenBucket {
	return &TokenBucket{
		tokens:     float64(maxTokens),
		maxTokens:  float64(maxTokens),
		refillRate: float64(refillRate),
		lastRefill: time.Now(),
		logging:    logging,
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refill()

	if tb.tokens >= 1 {
		tb.tokens--
		if tb.logging {
			log.Println("Request allowed, remaining tokens:", tb.tokens)
		}
		return true
	}

	if tb.logging {
		log.Println("Request denied due to rate limiting")
	}
	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.tokens+(elapsed*tb.refillRate), tb.maxTokens)
	tb.lastRefill = now
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
