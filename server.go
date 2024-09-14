package main

import (
	"go-cache/cache"
	"log"
	"net/http"
	"time"
)

func main() {

	cacheManager := cache.NewCacheManager(3, 60)

	http.HandleFunc("/cache/get", cacheManager.Get)
	http.HandleFunc("/cache/set", cacheManager.Set)
	port := ":8080"

	log.Printf("Starting server on port %s at %s", port, time.Now().Format(time.RFC3339))

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}
