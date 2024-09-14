package main

import (
	"fmt"
	"log"
	"net/http"
)

func getKey(w http.ResponseWriter, r *http.Request) { return }
func setKey(w http.ResponseWriter, r *http.Request) { return }

func main() {
	fmt.Println("Hello Server")

	http.HandleFunc("/cache/get", getKey)
	http.HandleFunc("/cache/set", setKey)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
