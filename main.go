package main

import (
	"fmt"
	"log"
	"net/http"

	ratelimiter "github.com/hasimy-as/tokenbucket-rate-limiter/service"
)

func main() {
	limiter := ratelimiter.NewTokenBucket(10, 1, true) // Max 10 tokens, 1 token per second, enable logging

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
