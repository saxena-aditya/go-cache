package main

import (
	"go-cache/cache"
	"log"
	"net/http"
	"time"
)

func main() {

	cacheManager := cache.NewCacheManager(3)

	http.HandleFunc("/cache/get", cacheManager.Get)
	http.HandleFunc("/cache/set", cacheManager.Set)
	// Define the port
	port := ":8080"

	// Log server start message
	log.Printf("Starting server on port %s at %s", port, time.Now().Format(time.RFC3339))

	// Start the server and log any errors
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed: %s", err)
	}

}
