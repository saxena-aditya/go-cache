package main

import (
	"go-cache/cache"
	"log"
	"net/http"
	"os"
	"time"
)

var logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

func main() {

	cacheManager := cache.NewCacheManager(3, 60)

	http.HandleFunc("/cache/get", cacheManager.Get)
	http.HandleFunc("/cache/set", cacheManager.Set)
	port := ":8080"

	logger.Printf("Starting server on port %s at %s", port, time.Now().Format(time.RFC3339))

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}
