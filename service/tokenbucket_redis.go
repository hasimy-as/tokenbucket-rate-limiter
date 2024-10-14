package service

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisTokenBucket struct {
	client     *redis.Client
	key        string
	maxTokens  int
	refillRate int // tokens per second
}

func NewRedisTokenBucket(client *redis.Client, key string, maxTokens, refillRate int) *RedisTokenBucket {
	log.Printf("Initializing token bucket: key=%s, maxTokens=%d, refillRate=%d tokens/sec", key, maxTokens, refillRate)
	return &RedisTokenBucket{
		client:     client,
		key:        key,
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
}

func (rtb *RedisTokenBucket) Allow() bool {
	rtb.refill()

	res, err := rtb.client.Decr(ctx, rtb.key).Result()
	if err != nil {
		log.Printf("Error decrementing tokens for key=%s: %v", rtb.key, err)
		return false
	}

	if res < 0 {
		rtb.client.Incr(ctx, rtb.key)
		log.Printf("Rate limit hit for key=%s. Tokens exhausted.", rtb.key)
		return false
	}

	log.Printf("Request allowed for key=%s. Remaining tokens=%d", rtb.key, res)
	return true
}

func (rtb *RedisTokenBucket) refill() {
	now := time.Now().Unix()
	lastRefillKey := rtb.key + ":lastRefill"

	lastRefill, err := rtb.client.Get(ctx, lastRefillKey).Int64()
	if err == redis.Nil {
		lastRefill = now
		rtb.client.Set(ctx, lastRefillKey, now, 0)
		log.Printf("Setting initial refill time for key=%s", lastRefillKey)
	} else if err != nil {
		log.Printf("Error retrieving last refill time for key=%s: %v", lastRefillKey, err)
		return
	}

	elapsed := now - lastRefill
	newTokens := int(elapsed) * rtb.refillRate
	if newTokens > 0 {
		rtb.client.Set(ctx, lastRefillKey, now, 0)
		rtb.client.IncrBy(ctx, rtb.key, int64(newTokens))
		rtb.client.SetNX(ctx, rtb.key, int64(rtb.maxTokens), 0)
		log.Printf("Refilled %d tokens for key=%s. Elapsed time=%d seconds", newTokens, rtb.key, elapsed)
	}
}
