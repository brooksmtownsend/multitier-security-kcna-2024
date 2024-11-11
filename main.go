//go:build wasip2

package main

import (
	"net/http"
	"time"

	"go.wasmcloud.dev/component/net/wasihttp"
)

var (
	wasiTransport = &wasihttp.Transport{ConnectTimeout: 30 * time.Second}
	httpClient    = &http.Client{Transport: wasiTransport}
)

func init() {
	// Register the handleRequest function as the handler for all incoming requests.
	wasihttp.HandleFunc(handleRequest)
}

// Since we don't run this program like a CLI, the `main` function is empty. Instead,
// we call the `handleRequest` function when an HTTP request is received.
func main() {}
