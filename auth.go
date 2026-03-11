package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// getToken reads the GITHUB_TOKEN environment variable and returns it along with a boolean
// indicating whether the token exists
func getToken() (string, bool) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", false
	}
	return token, true
}

// createAuthenticatedRequest creates an HTTP request with Authorization header if token is provided
func createAuthenticatedRequest(method, urlStr, token string) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}

	// Only add Authorization header if token is not empty
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

// authenticatedHTTPClient is an interface that supports the Do method
type authenticatedHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// fetchEventsWithPerPage fetches GitHub events with pagination support
// This function now supports optional authentication via GITHUB_TOKEN env var
func fetchEventsWithPerPage(username string, perPage int) ([]GitHubEvent, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	// Validate username format to prevent URL injection
	if !isValidUsername(username) {
		return nil, fmt.Errorf("invalid username format: %s", username)
	}

	// Validate per_page parameter
	if err := validatePerPage(perPage); err != nil {
		return nil, err
	}

	// Use default if perPage is 0
	if perPage == 0 {
		perPage = 30
	}

	urlStr := fmt.Sprintf("https://api.github.com/users/%s/events?per_page=%d", url.PathEscape(username), perPage)

	// Get token if available
	token, hasToken := getToken()

	// If we have a token, use authenticated request
	if hasToken && token != "" {
		return fetchAuthenticatedEvents(urlStr, token)
	}

	// Fall back to unauthenticated request (backward compatible)
	return fetchUnauthenticatedEvents(urlStr)
}

// fetchAuthenticatedEvents makes an authenticated request to GitHub API
func fetchAuthenticatedEvents(urlStr, token string) ([]GitHubEvent, error) {
	req, err := createAuthenticatedRequest("GET", urlStr, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Use default client to execute the request
	client, ok := defaultClient.(interface{ Do(req *http.Request) (*http.Response, error) })
	if !ok {
		// Fall back to simple GET if client doesn't support Do
		return fetchUnauthenticatedEvents(urlStr)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %v", err)
	}
	defer resp.Body.Close()

	// Handle specific error codes
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode == http.StatusForbidden {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read rate limit response: %v", err)
		}
		bodyStr := string(body)
		if isRateLimitError(bodyStr) {
			return nil, fmt.Errorf("GitHub API rate limit exceeded (403). Use a GITHUB_TOKEN for higher rate limits (5000/hour)")
		}
		return nil, fmt.Errorf("GitHub API returned status 403: %s", bodyStr)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed (401). Check that your GITHUB_TOKEN is valid")
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %v", err)
		}
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var events []GitHubEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return events, nil
}

// fetchUnauthenticatedEvents makes a simple GET request (backward compatible)
func fetchUnauthenticatedEvents(urlStr string) ([]GitHubEvent, error) {
	resp, err := defaultClient.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var events []GitHubEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return events, nil
}

// isRateLimitError checks if the response body indicates a rate limit error
func isRateLimitError(body string) bool {
	return strings.Contains(body, "rate limit") ||
		strings.Contains(body, "Rate limit") ||
		strings.Contains(body, "API rate limit") ||
		strings.Contains(body, "rate_limit")
}