package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	ratelimiter "github.com/hasimy-as/tokenbucket-rate-limiter/service"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Replace localhost with your redis server address
	})

	limiter := ratelimiter.NewRedisTokenBucket(rdb, "client_1", 10, 1) // Max 10 tokens, 1 token per second

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "too many requests sent!", http.StatusTooManyRequests)
			return
		}
		fmt.Fprintf(w, "request has been passed")
	})

	log.Println("server served on 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
