//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world hello --out gen ./wit
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/":
		rootHandler(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/hello":
		helloHandler(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/hello-file":
		helloFileHandler(w, r)
	// OAuth Implementation
	case r.Method == http.MethodGet && r.URL.Path == "/login":
		loginHandler(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/oauth/callback":
		callbackHandler(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// Simple root handler to return an HTML page
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Get client's IP and User-Agent
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	userAgent := r.Header.Get("User-Agent")

	// Create a more interesting response with HTML formatting
	response := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px;">
    <h1>Welcome to the Go Server!</h1>
    <h3>Request Details:</h3>
    <ul>
        <li>Time: %s</li>
        <li>Client IP: %s</li>
        <li>User Agent: %s</li>
        <li>Protocol: %s</li>
    </ul>
    <p>Try visiting <a href="/hello">/hello</a> for a different endpoint!</p>
    <p>Try visiting <a href="/login">/login</a> for a secure different endpoint!</p>
</body>
</html>`,
		time.Now().Format(time.RFC1123),
		clientIP,
		userAgent,
		r.Proto)

	// Set content type to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, response)
}

// Handler to return a hello world message, and then implement counting
func helloHandler(w http.ResponseWriter, _ *http.Request) {
	// bucket := store.Open("default")
	// if err := bucket.Err(); err != nil {
	// 	http.Error(w, err.String(), http.StatusInternalServerError)
	// 	return
	// }

	// count := atomics.Increment(*bucket.OK(), "hello", 1)
	// if err := count.Err(); err != nil {
	// 	http.Error(w, err.String(), http.StatusInternalServerError)
	// 	return
	// }
	// fmt.Fprintf(w, "Hello from Go! We've said hello %d times.\n", *count.OK())
	fmt.Fprintf(w, "Hello from Go!\n")
}

// Implement a file-based counter
func helloFileHandler(w http.ResponseWriter, _ *http.Request) {
	// Read the current count from file
	count := 0
	data, err := os.ReadFile("counter.txt")
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Error reading counter file: %s", err), http.StatusInternalServerError)
		return
	}
	if len(data) > 0 {
		count, err = strconv.Atoi(string(data))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing counter value: %s", err), http.StatusInternalServerError)
			return
		}
	}

	// Increment the counter
	count++

	// Write new count back to file
	err = os.WriteFile("counter.txt", []byte(strconv.Itoa(count)), 0644)
	if err != nil {
		http.Error(w, "Error writing counter file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Hello from Go! We've said hello %d times.\n", count)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	oauthConfig, err := oauthConfig()
	if err != nil {
		http.Error(w, "Failed to get OAuth config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// callbackHandler handles the GitHub callback and exchanges the authorization code for an access token.
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	oauthConfig, err := oauthConfig()
	if err != nil {
		http.Error(w, "Failed to get OAuth config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the authorization code from the URL query parameters
	code := r.FormValue("code")

	// Use custom HTTP client for token exchange
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Use token to make authenticated requests with custom HTTP client
	client := oauthConfig.Client(ctx, token)
	userInfo, err := getUserInfo(client)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Print user information
	fmt.Fprintf(w, "User Info: %s\n", userInfo)
}

// // getUserInfo fetches user information from GitHub's API using the authenticated client.
func getUserInfo(client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	// Add required headers
	req.Header.Set("User-Agent", "Go-HTTP-Client/Multitier-Security-Example")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	userInfoJSON, err := json.MarshalIndent(userInfo, "", "  ")
	if err != nil {
		return "", fmt.Errorf("formatting JSON: %w", err)
	}
	return string(userInfoJSON), nil
}

// Fetch the OAuth2 config including the client ID and secret
// as secrets.
func oauthConfig() (oauth2.Config, error) {
	// clientId := secretstore.Get("client_id")
	// if err := clientId.Err(); err != nil {
	// 	return oauth2.Config{}, fmt.Errorf("getting client ID: %s", err.String())
	// }
	// clientSecret := secretstore.Get("client_secret")
	// if err := clientSecret.Err(); err != nil {
	// 	return oauth2.Config{}, fmt.Errorf("getting client secret: %s", err.String())
	// }

	// fmt.Fprintf(os.Stderr, "Client ID: %d\n", clientId.OK())
	// fmt.Fprintf(os.Stderr, "Client Secret: %d\n", clientSecret.OK())

	// clientIdReal := reveal.Reveal(*clientId.OK())
	// clientSecretReal := reveal.Reveal(*clientSecret.OK())
	// return oauth2.Config{
	// 	ClientID:     *clientIdReal.String_(),
	// 	ClientSecret: *clientSecretReal.String_(),
	// 	RedirectURL:  "http://127.0.0.1:8000/oauth/callback",
	// 	Scopes:       []string{},
	// 	Endpoint:     github.Endpoint,
	// }, nil
	return oauth2.Config{}, nil
}
