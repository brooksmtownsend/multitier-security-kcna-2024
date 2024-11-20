//go:build !wasip2

package main

import (
	"fmt"
	"log"
	"net/http"
)

var httpClient = http.Client{}

func main() {
	router := Router()
	port := ":8080"
	fmt.Printf("Starting server on %s\n", port)
	// Start the server
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
