package service

import (
	"context"
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
		return false
	}

	if res < 0 {
		rtb.client.Incr(ctx, rtb.key)
		return false
	}

	return true
}

func (rtb *RedisTokenBucket) refill() {
	now := time.Now().Unix()
	lastRefillKey := rtb.key + ":lastRefill"

	lastRefill, err := rtb.client.Get(ctx, lastRefillKey).Int64()
	if err == redis.Nil {
		lastRefill = now
		rtb.client.Set(ctx, lastRefillKey, now, 0)
	}

	elapsed := now - lastRefill
	newTokens := int(elapsed) * rtb.refillRate
	if newTokens > 0 {
		rtb.client.Set(ctx, lastRefillKey, now, 0)
		rtb.client.IncrBy(ctx, rtb.key, int64(newTokens))
		rtb.client.SetNX(ctx, rtb.key, int64(rtb.maxTokens), 0)
	}
}
