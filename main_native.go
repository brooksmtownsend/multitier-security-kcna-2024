//go:build !wasip2

package main

import (
	"fmt"
	"log"
	"net/http"
)

var (
	httpClient = http.Client{}
)

func main() {
	// Register our HTTP handlers
	http.HandleFunc("/", handleRequest)

	// Start the server
	port := ":8080"
	fmt.Printf("Starting server on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
