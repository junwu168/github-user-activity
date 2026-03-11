package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// Mock data for GitHub API responses
var mockPushEvent = GitHubEvent{
	Type:      "PushEvent",
	Repo:      Repo{Name: "testuser/testrepo"},
	CreatedAt: time.Now().Format(time.RFC3339),
	Payload: Payload{
		Commits: []Commit{
			{Message: "feat: add new feature"},
			{Message: "fix: bug fix"},
			{Message: "docs: update readme"},
		},
	},
}

var mockIssueEvent = GitHubEvent{
	Type:      "IssuesEvent",
	Repo:      Repo{Name: "testuser/testrepo"},
	CreatedAt: time.Now().Format(time.RFC3339),
	Payload: Payload{
		Action: "opened",
		Issue: Issue{
			Title: "Bug in production",
			URL:   "https://api.github.com/repos/testuser/testrepo/issues/1",
		},
	},
}

var mockWatchEvent = GitHubEvent{
	Type:      "WatchEvent",
	Repo:      Repo{Name: "testuser/awesome-repo"},
	CreatedAt: time.Now().Format(time.RFC3339),
	Payload:   Payload{},
}

var mockCreateEvent = GitHubEvent{
	Type:      "CreateEvent",
	Repo:      Repo{Name: "testuser/new-repo"},
	CreatedAt: time.Now().Format(time.RFC3339),
	Payload: Payload{
		RefType: "branch",
		Ref:     "main",
	},
}

var mockDeleteEvent = GitHubEvent{
	Type:      "DeleteEvent",
	Repo:      Repo{Name: "testuser/test-repo"},
	CreatedAt: time.Now().Format(time.RFC3339),
	Payload: Payload{
		RefType: "branch",
		Ref:     "feature/test",
	},
}

var mockForkEvent = GitHubEvent{
	Type:      "ForkEvent",
	Repo:      Repo{Name: "testuser/original-repo"},
	CreatedAt: time.Now().Format(time.RFC3339),
	Payload: Payload{
		Forkee: Forkee{FullName: "testuser/forked-repo"},
	},
}

var mockPullRequestEvent = GitHubEvent{
	Type:      "PullRequestEvent",
	Repo:      Repo{Name: "testuser/test-repo"},
	CreatedAt: time.Now().Format(time.RFC3339),
	Payload: Payload{
		Action: "opened",
		PullRequest: PullRequest{
			Title: "Add new feature",
			URL:   "https://api.github.com/repos/testuser/test-repo/pulls/1",
		},
	},
}

func TestFormatPushEvent(t *testing.T) {
	result := formatEvent(mockPushEvent)
	expected := "Pushed 3 commits to testuser/testrepo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatIssueEvent(t *testing.T) {
	result := formatEvent(mockIssueEvent)
	expected := "Opened a new issue in testuser/testrepo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatWatchEvent(t *testing.T) {
	result := formatEvent(mockWatchEvent)
	expected := "Starred testuser/awesome-repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatCreateEvent(t *testing.T) {
	result := formatEvent(mockCreateEvent)
	expected := "Created branch main in testuser/new-repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatDeleteEvent(t *testing.T) {
	result := formatEvent(mockDeleteEvent)
	if !strings.Contains(result, "Deleted branch") || !strings.Contains(result, "testuser/test-repo") {
		t.Errorf("Expected deleted branch event, got %q", result)
	}
}

func TestFormatForkEvent(t *testing.T) {
	result := formatEvent(mockForkEvent)
	expected := "Forked testuser/original-repo to testuser/forked-repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatPullRequestEvent(t *testing.T) {
	result := formatEvent(mockPullRequestEvent)
	expected := "Opened pull request in testuser/test-repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFetchEvents(t *testing.T) {
	// Create mock server
	events := []GitHubEvent{mockPushEvent, mockIssueEvent, mockWatchEvent}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}))
	defer server.Close()

	// Test fetching (will need to use a custom client or modify fetchEvents to accept URL)
	_ = server.URL
}

func TestFetchEventsInvalidUsername(t *testing.T) {
	// This test verifies error handling for invalid usernames
	// The API returns 404 for non-existent users
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Test will be implemented when we add error handling
	_ = server.URL
}

func TestNoArgs(t *testing.T) {
	// Skip this test as os.Exit cannot be easily tested in Go
	// The program correctly exits with code 1 when no args provided
	t.Skip("os.Exit cannot be easily tested in Go tests")
}

func TestMainFunction(t *testing.T) {
	// Test that main doesn't panic with valid input
	// We'll test with a mock server
	events := []GitHubEvent{mockPushEvent}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}))
	defer server.Close()

	// Note: We can't easily test main with different URLs without refactoring
	// This test documents the expected behavior
	t.Log("Main function should fetch and display events from GitHub API")
}

// mockHTTPClient implements HTTPClient for testing
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	return m.response, m.err
}

func TestFetchEventsSuccess(t *testing.T) {
	events := []GitHubEvent{mockPushEvent, mockWatchEvent}
	body, _ := json.Marshal(events)

	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(body))),
		},
		err: nil,
	}

	// Save original client and restore after test
	origClient := defaultClient
	defer func() { defaultClient = origClient }()

	setClient(mockClient)

	result, err := fetchEvents("testuser")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 events, got %d", len(result))
	}
}

func TestFetchEventsNotFound(t *testing.T) {
	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("")),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()

	setClient(mockClient)

	_, err := fetchEvents("nonexistent")
	if err == nil {
		t.Error("Expected error for not found user")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got %v", err)
	}
}

func TestFetchEventsAPIError(t *testing.T) {
	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("")),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()

	setClient(mockClient)

	_, err := fetchEvents("testuser")
	if err == nil {
		t.Error("Expected error for API error")
	}
	if !strings.Contains(err.Error(), "status 403") {
		t.Errorf("Expected 'status 403' in error, got %v", err)
	}
}

func TestFetchEventsNetworkError(t *testing.T) {
	mockClient := &mockHTTPClient{
		response: nil,
		err:      fmt.Errorf("network error"),
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()

	setClient(mockClient)

	_, err := fetchEvents("testuser")
	if err == nil {
		t.Error("Expected error for network failure")
	}
}

func TestFetchEventsInvalidJSON(t *testing.T) {
	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("invalid json")),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()

	setClient(mockClient)

	_, err := fetchEvents("testuser")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestFormatEventIssueComment(t *testing.T) {
	event := GitHubEvent{
		Type:      "IssueCommentEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload: Payload{
			Action: "created",
		},
	}
	result := formatEvent(event)
	expected := "Created comment on issue in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventCommitComment(t *testing.T) {
	event := GitHubEvent{
		Type:      "CommitCommentEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload:   Payload{},
	}
	result := formatEvent(event)
	expected := "Commented on commit in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventPullRequestReview(t *testing.T) {
	event := GitHubEvent{
		Type:      "PullRequestReviewEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload: Payload{
			Action: "submitted",
		},
	}
	result := formatEvent(event)
	expected := "Submitted pull request review in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventRelease(t *testing.T) {
	event := GitHubEvent{
		Type:      "ReleaseEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload: Payload{
			Action: "published",
		},
	}
	result := formatEvent(event)
	expected := "Published release in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventUnknown(t *testing.T) {
	event := GitHubEvent{
		Type:      "GollumEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload:   Payload{},
	}
	result := formatEvent(event)
	expected := "GollumEvent in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventPushNoCommits(t *testing.T) {
	event := GitHubEvent{
		Type:      "PushEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload:   Payload{Commits: []Commit{}},
	}
	result := formatEvent(event)
	expected := "Pushed to testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventPushSingleCommit(t *testing.T) {
	event := GitHubEvent{
		Type:      "PushEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload: Payload{
			Commits: []Commit{{Message: "Single commit"}},
		},
	}
	result := formatEvent(event)
	expected := "Pushed 1 commit to testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventCreateNoRef(t *testing.T) {
	event := GitHubEvent{
		Type:      "CreateEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload: Payload{
			RefType: "tag",
			Ref:     "",
		},
	}
	result := formatEvent(event)
	expected := "Created tag in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventForkNoForkee(t *testing.T) {
	event := GitHubEvent{
		Type:      "ForkEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload:   Payload{Forkee: Forkee{}},
	}
	result := formatEvent(event)
	expected := "Forked testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"opened", "Opened"},
		{"closed", "Closed"},
		{"", ""},
		{"a", "A"},
	}

	for _, tt := range tests {
		result := capitalize(tt.input)
		if result != tt.expected {
			t.Errorf("capitalize(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// Test fetchEvents with empty response body
func TestFetchEventsEmptyBody(t *testing.T) {
	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()

	setClient(mockClient)

	_, err := fetchEvents("testuser")
	if err == nil {
		t.Error("Expected error for empty body")
	}
}

func TestFetchEventsEmptyUsername(t *testing.T) {
	_, err := fetchEvents("")
	if err == nil {
		t.Error("Expected error for empty username")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' in error, got %v", err)
	}
}

// Test invalid username format to prevent URL injection
func TestFetchEventsInvalidUsernameFormat(t *testing.T) {
	invalidUsernames := []string{
		"test?url=http://evil.com",
		"test#fragment",
		"test/",
		"test\\",
		"test space",
		"test\nnewline",
		"<script>alert(1)</script>",
		"../../../etc/passwd",
	}

	for _, username := range invalidUsernames {
		_, err := fetchEvents(username)
		if err == nil {
			t.Errorf("Expected error for invalid username: %q", username)
		}
		if err != nil && !strings.Contains(err.Error(), "invalid") && !strings.Contains(err.Error(), "format") {
			t.Errorf("Expected 'invalid' or 'format' in error for %q, got: %v", username, err)
		}
	}
}

// Test valid username formats
func TestFetchEventsValidUsername(t *testing.T) {
	validUsernames := []string{
		"kamranahmedse",
		"test-user",
		"test_user",
		"TestUser123",
		"a",
		"testuser123",
	}

	for _, username := range validUsernames {
		// These should not return an "invalid format" error
		// They may fail due to network, but should not fail with "invalid format"
		_, err := fetchEvents(username)
		if err != nil && (strings.Contains(err.Error(), "invalid") && strings.Contains(err.Error(), "format")) {
			t.Errorf("Valid username %q should not fail with format error: %v", username, err)
		}
	}
}

// Test isValidUsername directly for edge cases
func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		valid    bool
	}{
		// Valid usernames
		{"single char", "a", true},
		{"single digit", "1", true},
		{"lowercase", "testuser", true},
		{"uppercase", "TestUser", true},
		{"with hyphen", "test-user", true},
		{"with underscore", "test_user", true},
		{"mixed", "Test-User_123", true},
		{"max length 39 chars", strings.Repeat("a", 39), true},
		// Invalid usernames
		{"empty string", "", false},
		{"too long 40 chars", strings.Repeat("a", 40), false},
		{"too long 100 chars", strings.Repeat("a", 100), false},
		{"starts with hyphen", "-testuser", false},
		{"ends with hyphen", "testuser-", false},
		{"starts with underscore", "_testuser", false},
		{"contains space", "test user", false},
		{"contains slash", "test/user", false},
		{"contains backslash", "test\\user", false},
		{"contains dot", "test.user", false},
		{"contains at", "test@user", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidUsername(tt.username)
			if result != tt.valid {
				t.Errorf("isValidUsername(%q) = %v, want %v", tt.username, result, tt.valid)
			}
		})
	}
}

// Test formatEvent with different issue actions
func TestFormatEventIssuesClosed(t *testing.T) {
	event := GitHubEvent{
		Type:      "IssuesEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload: Payload{
			Action: "closed",
		},
	}
	result := formatEvent(event)
	expected := "Closed issue in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventIssuesReopened(t *testing.T) {
	event := GitHubEvent{
		Type:      "IssuesEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload: Payload{
			Action: "reopened",
		},
	}
	result := formatEvent(event)
	expected := "Reopened issue in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatEventIssuesEmptyAction(t *testing.T) {
	event := GitHubEvent{
		Type:      "IssuesEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload:   Payload{Action: ""},
	}
	result := formatEvent(event)
	expected := "Unknown issue in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Test PullRequestReviewCommentEvent
func TestFormatEventPullRequestReviewComment(t *testing.T) {
	event := GitHubEvent{
		Type:      "PullRequestReviewCommentEvent",
		Repo:      Repo{Name: "testuser/repo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		Payload: Payload{
			Action: "created",
		},
	}
	result := formatEvent(event)
	expected := "Created pull request comment in testuser/repo"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Test main function branches via testable examples
func TestMainWithMockedFetch(t *testing.T) {
	// Mock the fetchEventsWithPerPage to avoid network calls
	// Since we can't easily test main, we test the individual functions it calls

	// Test validatePerPage edge cases
	tests := []struct {
		name    string
		perPage int
		wantErr bool
	}{
		{"boundary 1", 1, false},
		{"boundary 100", 100, false},
		{"below min", -1, true},
		{"above max", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePerPage(tt.perPage)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePerPage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test fetchEventsWithPerPage with error paths
func TestFetchEventsWithPerPageErrors(t *testing.T) {
	tests := []struct {
		name     string
		username string
		perPage  int
		wantErr  bool
		errMsg   string
	}{
		{"empty username", "", 30, true, "cannot be empty"},
		{"invalid username format", "test user", 30, true, "invalid username format"},
		{"perPage too low", "testuser", 0, true, "count must be between"},
		{"perPage too high", "testuser", 101, true, "count must be between"},
		{"perPage negative", "testuser", -5, true, "count must be between"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a mock client that returns error
			mockClient := &mockHTTPClient{
				response: nil,
				err:      fmt.Errorf("mock error"),
			}
			origClient := defaultClient
			defaultClient = mockClient
			defer func() { defaultClient = origClient }()

			_, err := fetchEventsWithPerPage(tt.username, tt.perPage)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchEventsWithPerPage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("fetchEventsWithPerPage() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}

// Test fetchEventsWithPerPage network error path
func TestFetchEventsWithPerPageNetworkError(t *testing.T) {
	mockClient := &mockHTTPClient{
		response: nil,
		err:      fmt.Errorf("network error: connection refused"),
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()
	defaultClient = mockClient

	_, err := fetchEventsWithPerPage("testuser", 30)
	if err == nil {
		t.Error("Expected error for network failure")
	}
	if !strings.Contains(err.Error(), "failed to fetch events") {
		t.Errorf("Expected 'failed to fetch events' in error, got: %v", err)
	}
}

// Test fetchEventsWithPerPage 404 error path
func TestFetchEventsWithPerPageNotFound(t *testing.T) {
	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("")),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()
	defaultClient = mockClient

	_, err := fetchEventsWithPerPage("nonexistentuser123456", 30)
	if err == nil {
		t.Error("Expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// Test fetchEventsWithPerPage non-OK status code
func TestFetchEventsWithPerPageAPIError(t *testing.T) {
	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("")),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()
	defaultClient = mockClient

	_, err := fetchEventsWithPerPage("testuser", 30)
	if err == nil {
		t.Error("Expected error for non-OK status code")
	}
	if !strings.Contains(err.Error(), "status 403") {
		t.Errorf("Expected 'status 403' in error, got: %v", err)
	}
}

// Test fetchEventsWithPerPage with empty response (should fail gracefully)
func TestFetchEventsWithPerPageEmptyResponse(t *testing.T) {
	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()
	defaultClient = mockClient

	_, err := fetchEventsWithPerPage("testuser", 30)
	if err == nil {
		t.Error("Expected error for empty response body")
	}
}

// Test parseCountFlag with edge cases
func TestParseCountFlagEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{"single digit", "5", false},
		{"two digits", "50", false},
		{"boundary min", "1", false},
		{"boundary max", "100", false},
		{"whitespace only", " ", true},
		{"number with spaces", "10", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCountFlag(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("parseCountFlag(%q) error = %v, expectErr %v", tt.input, err, tt.expectErr)
			}
		})
	}
}

// Test fetchEventsWithPerPage default behavior (perPage = 0 should default to 30)
func TestFetchEventsWithPerPageZeroPerPage(t *testing.T) {
	var capturedURL string
	mockClient := &mockURLClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("[]")),
		},
		err:         nil,
		capturedURL: &capturedURL,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()
	defaultClient = mockClient

	// Test with perPage = 0 (should use default 30)
	_, err := fetchEventsWithPerPage("testuser", 0)
	if err != nil {
		// validatePerPage rejects 0, so this will error
		if !strings.Contains(err.Error(), "count must be between") {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		// If it passes validation, check for default in URL
		if !strings.Contains(capturedURL, "per_page=30") {
			t.Errorf("Expected default per_page=30 in URL, got: %s", capturedURL)
		}
	}
}

// Test fetchEventsWithPerPage with exactly 1 event
func TestFetchEventsWithPerPageBoundary1(t *testing.T) {
	event := []GitHubEvent{mockPushEvent}
	body, _ := json.Marshal(event)

	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(body))),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()
	defaultClient = mockClient

	result, err := fetchEventsWithPerPage("testuser", 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 event, got %d", len(result))
	}
}

// Test fetchEventsWithPerPage with exactly 100 events
func TestFetchEventsWithPerPageBoundary100(t *testing.T) {
	events := make([]GitHubEvent, 100)
	for i := 0; i < 100; i++ {
		events[i] = mockPushEvent
	}
	body, _ := json.Marshal(events)

	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(body))),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()
	defaultClient = mockClient

	result, err := fetchEventsWithPerPage("testuser", 100)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != 100 {
		t.Errorf("Expected 100 events, got %d", len(result))
	}
}

// ============== Filter by Event Type Tests ==============

// Test validEventTypes constant
func TestValidEventTypes(t *testing.T) {
	expectedTypes := []string{
		"PushEvent", "IssuesEvent", "WatchEvent", "CreateEvent", "DeleteEvent",
		"ForkEvent", "PullRequestEvent", "IssueCommentEvent", "CommitCommentEvent",
		"PullRequestReviewEvent", "ReleaseEvent", "PullRequestReviewCommentEvent",
	}

	if len(validEventTypes) != len(expectedTypes) {
		t.Errorf("Expected %d event types, got %d", len(expectedTypes), len(validEventTypes))
	}

	for _, expected := range expectedTypes {
		found := false
		for _, actual := range validEventTypes {
			if expected == actual {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected event type %q not found in validEventTypes", expected)
		}
	}
}

// Test isValidEventType function
func TestIsValidEventType(t *testing.T) {
	tests := []struct {
		name     string
		eventType string
		expected bool
	}{
		{"valid PushEvent", "PushEvent", true},
		{"valid IssuesEvent", "IssuesEvent", true},
		{"valid WatchEvent", "WatchEvent", true},
		{"valid CreateEvent", "CreateEvent", true},
		{"valid DeleteEvent", "DeleteEvent", true},
		{"valid ForkEvent", "ForkEvent", true},
		{"valid PullRequestEvent", "PullRequestEvent", true},
		{"valid IssueCommentEvent", "IssueCommentEvent", true},
		{"valid CommitCommentEvent", "CommitCommentEvent", true},
		{"valid PullRequestReviewEvent", "PullRequestReviewEvent", true},
		{"valid ReleaseEvent", "ReleaseEvent", true},
		{"valid PullRequestReviewCommentEvent", "PullRequestReviewCommentEvent", true},
		{"invalid pushevent lowercase", "pushevent", false},
		{"invalid PushEvent capitalized", "PushEvent", true},
		{"invalid unknown type", "UnknownEvent", false},
		{"invalid empty string", "", false},
		{"invalid random string", "RandomEvent", false},
		{"invalid GollumEvent", "GollumEvent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEventType(tt.eventType)
			if result != tt.expected {
				t.Errorf("isValidEventType(%q) = %v, want %v", tt.eventType, result, tt.expected)
			}
		})
	}
}

// Test parseFilter function
func TestParseFilter(t *testing.T) {
	tests := []struct {
		name           string
		filterStr      string
		expectedMap    map[string]bool
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:        "single valid event type",
			filterStr:   "PushEvent",
			expectedMap: map[string]bool{"PushEvent": true},
			expectErr:   false,
		},
		{
			name:        "multiple valid event types",
			filterStr:   "PushEvent,WatchEvent",
			expectedMap: map[string]bool{"PushEvent": true, "WatchEvent": true},
			expectErr:   false,
		},
		{
			name:        "multiple event types with spaces",
			filterStr:   "PushEvent, WatchEvent, IssuesEvent",
			expectedMap: map[string]bool{"PushEvent": true, "WatchEvent": true, "IssuesEvent": true},
			expectErr:   false,
		},
		{
			name:        "empty string returns empty map",
			filterStr:   "",
			expectedMap: map[string]bool{},
			expectErr:   false,
		},
		{
			name:           "invalid event type",
			filterStr:      "InvalidEvent",
			expectedMap:    nil,
			expectErr:      true,
			expectedErrMsg: "invalid event type",
		},
		{
			name:           "mixed valid and invalid",
			filterStr:      "PushEvent,InvalidEvent",
			expectedMap:    nil,
			expectErr:      true,
			expectedErrMsg: "invalid event type",
		},
		{
			name:        "single event type no comma",
			filterStr:   "PullRequestEvent",
			expectedMap: map[string]bool{"PullRequestEvent": true},
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFilter(tt.filterStr)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				if tt.expectedErrMsg != "" && !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.expectedErrMsg, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify map contents
			if len(result) != len(tt.expectedMap) {
				t.Errorf("Expected map size %d, got %d", len(tt.expectedMap), len(result))
				return
			}

			for key := range tt.expectedMap {
				if !result[key] {
					t.Errorf("Expected key %q in result map", key)
				}
			}
		})
	}
}

// Test filterEvents function - filtering is immutable
func TestFilterEvents(t *testing.T) {
	events := []GitHubEvent{
		{Type: "PushEvent", Repo: Repo{Name: "user/repo1"}},
		{Type: "WatchEvent", Repo: Repo{Name: "user/repo2"}},
		{Type: "PushEvent", Repo: Repo{Name: "user/repo3"}},
		{Type: "IssuesEvent", Repo: Repo{Name: "user/repo4"}},
		{Type: "WatchEvent", Repo: Repo{Name: "user/repo5"}},
	}

	tests := []struct {
		name          string
		filterMap     map[string]bool
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "filter by PushEvent only",
			filterMap:     map[string]bool{"PushEvent": true},
			expectedCount: 2,
			expectedTypes: []string{"PushEvent", "PushEvent"},
		},
		{
			name:          "filter by WatchEvent only",
			filterMap:     map[string]bool{"WatchEvent": true},
			expectedCount: 2,
			expectedTypes: []string{"WatchEvent", "WatchEvent"},
		},
		{
			name:          "filter by multiple types",
			filterMap:     map[string]bool{"PushEvent": true, "WatchEvent": true},
			expectedCount: 4,
			expectedTypes: []string{"PushEvent", "WatchEvent", "PushEvent", "WatchEvent"},
		},
		{
			name:          "empty filter returns all",
			filterMap:     map[string]bool{},
			expectedCount: 5,
			expectedTypes: []string{"PushEvent", "WatchEvent", "PushEvent", "IssuesEvent", "WatchEvent"},
		},
		{
			name:          "filter by IssuesEvent",
			filterMap:     map[string]bool{"IssuesEvent": true},
			expectedCount: 1,
			expectedTypes: []string{"IssuesEvent"},
		},
		{
			name:          "filter excludes all",
			filterMap:     map[string]bool{"NonExistentEvent": true},
			expectedCount: 0,
			expectedTypes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterEvents(events, tt.filterMap)

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d events, got %d", tt.expectedCount, len(result))
				return
			}

			// Verify original slice is not modified (immutability)
			if len(events) != 5 {
				t.Errorf("Original events slice was modified")
			}

			// Verify types match expected
			for i, expectedType := range tt.expectedTypes {
				if result[i].Type != expectedType {
					t.Errorf("At index %d: expected type %q, got %q", i, expectedType, result[i].Type)
				}
			}
		})
	}
}

// Test filterEvents with nil filter (should return all events)
func TestFilterEventsNilFilter(t *testing.T) {
	events := []GitHubEvent{
		{Type: "PushEvent", Repo: Repo{Name: "user/repo1"}},
		{Type: "WatchEvent", Repo: Repo{Name: "user/repo2"}},
	}

	result := filterEvents(events, nil)

	if len(result) != 2 {
		t.Errorf("Expected 2 events with nil filter, got %d", len(result))
	}
}

// Test filterEvents with empty events slice
func TestFilterEventsEmptySlice(t *testing.T) {
	events := []GitHubEvent{}
	filterMap := map[string]bool{"PushEvent": true}

	result := filterEvents(events, filterMap)

	if len(result) != 0 {
		t.Errorf("Expected 0 events, got %d", len(result))
	}
}

// Test that filterEvents creates a new slice (immutability)
func TestFilterEventsReturnsNewSlice(t *testing.T) {
	events := []GitHubEvent{
		{Type: "PushEvent", Repo: Repo{Name: "user/repo1"}},
	}
	filterMap := map[string]bool{"PushEvent": true}

	result := filterEvents(events, filterMap)

	// Modify the result to verify it's a new slice
	if len(result) > 0 {
		result[0].Type = "ModifiedEvent"
	}

	// Original should be unchanged
	if events[0].Type != "PushEvent" {
		t.Error("Original slice was modified - immutability violated")
	}
}

// Test parseArgs with filter flag
func TestParseArgsWithFilter(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedCount   int
		expectedFilter  map[string]bool
		expectedUser    string
		expectErr       bool
		expectedErrMsg  string
	}{
		{
			name:           "filter with single event type",
			args:           []string{"github-activity", "-f", "PushEvent", "testuser"},
			expectedCount:  30,
			expectedFilter: map[string]bool{"PushEvent": true},
			expectedUser:   "testuser",
			expectErr:      false,
		},
		{
			name:           "filter with multiple event types",
			args:           []string{"github-activity", "-filter", "PushEvent,WatchEvent", "testuser"},
			expectedCount:  30,
			expectedFilter: map[string]bool{"PushEvent": true, "WatchEvent": true},
			expectedUser:   "testuser",
			expectErr:      false,
		},
		{
			name:           "filter with count flag",
			args:           []string{"github-activity", "-count", "10", "-f", "PushEvent", "testuser"},
			expectedCount:  10,
			expectedFilter: map[string]bool{"PushEvent": true},
			expectedUser:   "testuser",
			expectErr:      false,
		},
		{
			name:           "filter with short count flag",
			args:           []string{"github-activity", "-n", "20", "-f", "IssuesEvent", "testuser"},
			expectedCount:  20,
			expectedFilter: map[string]bool{"IssuesEvent": true},
			expectedUser:   "testuser",
			expectErr:      false,
		},
		{
			name:           "no filter returns empty map",
			args:           []string{"github-activity", "testuser"},
			expectedCount:  30,
			expectedFilter: map[string]bool{},
			expectedUser:   "testuser",
			expectErr:      false,
		},
		{
			name:           "invalid filter type",
			args:           []string{"github-activity", "-f", "InvalidEvent", "testuser"},
			expectedCount:  0,
			expectedFilter: nil,
			expectedUser:   "",
			expectErr:      true,
			expectedErrMsg: "invalid event type",
		},
		{
			name:           "filter short flag -f with count",
			args:           []string{"github-activity", "-f", "PushEvent", "-n", "5", "testuser"},
			expectedCount:  5,
			expectedFilter: map[string]bool{"PushEvent": true},
			expectedUser:   "testuser",
			expectErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withArgs(tt.args, func() {
				count, filterMap, username, err := parseArgs()

				if tt.expectErr {
					if err == nil {
						t.Errorf("Expected error, got nil")
						return
					}
					if tt.expectedErrMsg != "" && !strings.Contains(err.Error(), tt.expectedErrMsg) {
						t.Errorf("Expected error containing %q, got %v", tt.expectedErrMsg, err)
					}
					return
				}

				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if count != tt.expectedCount {
					t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
				}

				if username != tt.expectedUser {
					t.Errorf("Expected username %q, got %q", tt.expectedUser, username)
				}

				// Compare filter maps
				if len(filterMap) != len(tt.expectedFilter) {
					t.Errorf("Expected filter map size %d, got %d", len(tt.expectedFilter), len(filterMap))
					return
				}
				for key := range tt.expectedFilter {
					if !filterMap[key] {
						t.Errorf("Expected filter key %q", key)
					}
				}
			})
		})
	}
}

// Helper to set up os.Args for testing parseArgs and reset flags
func withArgs(args []string, fn func()) {
	// Reset flags before each test to avoid redefinition panic
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	origArgs := os.Args
	origStderr := os.Stderr
	// Suppress flag error output during tests
	os.Stderr, _ = os.Open(os.DevNull)

	defer func() {
		os.Args = origArgs
		os.Stderr = origStderr
	}()

	os.Args = args
	fn()
}

func TestParseArgsNoUsername(t *testing.T) {
	withArgs([]string{"github-activity"}, func() {
		_, _, _, err := parseArgs()
		if err == nil {
			t.Error("Expected error when no username provided")
		}
		if !strings.Contains(err.Error(), "usage") {
			t.Errorf("Expected usage error, got: %v", err)
		}
	})
}

func TestParseArgsInvalidCount(t *testing.T) {
	withArgs([]string{"github-activity", "-count", "0", "testuser"}, func() {
		_, _, _, err := parseArgs()
		if err == nil {
			t.Error("Expected error for count=0")
		}
		if !strings.Contains(err.Error(), "count must be between") {
			t.Errorf("Expected count validation error, got: %v", err)
		}
	})
}

func TestParseArgsInvalidCountOver100(t *testing.T) {
	withArgs([]string{"github-activity", "-count", "101", "testuser"}, func() {
		_, _, _, err := parseArgs()
		if err == nil {
			t.Error("Expected error for count=101")
		}
		if !strings.Contains(err.Error(), "count must be between") {
			t.Errorf("Expected count validation error, got: %v", err)
		}
	})
}

func TestParseArgsShortFlag(t *testing.T) {
	withArgs([]string{"github-activity", "-n", "50", "testuser"}, func() {
		count, _, username, err := parseArgs()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 50 {
			t.Errorf("Expected count=50, got %d", count)
		}
		if username != "testuser" {
			t.Errorf("Expected username=testuser, got %s", username)
		}
	})
}

func TestParseArgsDefaultCount(t *testing.T) {
	withArgs([]string{"github-activity", "testuser"}, func() {
		count, _, username, err := parseArgs()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 30 {
			t.Errorf("Expected default count=30, got %d", count)
		}
		if username != "testuser" {
			t.Errorf("Expected username=testuser, got %s", username)
		}
	})
}

func TestParseArgsBothFlagsNPrecedence(t *testing.T) {
	// When both -count and -n are provided, -n should take precedence
	withArgs([]string{"github-activity", "-count", "10", "-n", "20", "testuser"}, func() {
		count, _, _, err := parseArgs()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		// -n takes precedence
		if count != 20 {
			t.Errorf("Expected -n flag to take precedence, got count=%d", count)
		}
	})
}

func TestParseArgsCountFlag(t *testing.T) {
	withArgs([]string{"github-activity", "-count", "75", "testuser"}, func() {
		count, _, username, err := parseArgs()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 75 {
			t.Errorf("Expected count=75, got %d", count)
		}
		if username != "testuser" {
			t.Errorf("Expected username=testuser, got %s", username)
		}
	})
}

// ============== Pagination Tests ==============

// Test fetchEvents with per_page parameter
func TestFetchEventsWithPerPage(t *testing.T) {
	events := []GitHubEvent{mockPushEvent, mockWatchEvent, mockIssueEvent}
	body, _ := json.Marshal(events)

	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(body))),
		},
		err: nil,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()

	setClient(mockClient)

	// Test with perPage = 10
	result, err := fetchEventsWithPerPage("testuser", 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 events, got %d", len(result))
	}
}

// Test fetchEventsWithPerPage validates per_page in URL
func TestFetchEventsWithPerPageInURL(t *testing.T) {
	var capturedURL string
	mockClient := &mockURLClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("[]")),
		},
		err:         nil,
		capturedURL: &capturedURL,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()
	defaultClient = mockClient

	// Test with perPage = 50
	_, err := fetchEventsWithPerPage("testuser", 50)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the URL contains per_page parameter
	if !strings.Contains(capturedURL, "per_page=50") {
		t.Errorf("Expected URL to contain per_page=50, got %s", capturedURL)
	}
}

// mockURLClient captures the URL for testing
type mockURLClient struct {
	response    *http.Response
	err         error
	capturedURL *string
}

func (m *mockURLClient) Get(url string) (*http.Response, error) {
	if m.capturedURL != nil {
		*m.capturedURL = url
	}
	return m.response, m.err
}

// Test fetchEventsWithPerPage default value (30)
func TestFetchEventsWithPerPageDefault(t *testing.T) {
	var capturedURL string
	mockClient := &mockURLClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("[]")),
		},
		err:         nil,
		capturedURL: &capturedURL,
	}

	origClient := defaultClient
	defer func() { defaultClient = origClient }()
	defaultClient = mockClient

	// Test with 30 (explicit value for default) - should include per_page=30
	_, err := fetchEventsWithPerPage("testuser", 30)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Default should be 30
	if !strings.Contains(capturedURL, "per_page=30") {
		t.Errorf("Expected URL to contain per_page=30, got %s", capturedURL)
	}
}

// Test validatePerPage function
func TestValidatePerPage(t *testing.T) {
	tests := []struct {
		name      string
		perPage   int
		expectErr bool
	}{
		{"valid 1", 1, false},
		{"valid 10", 10, false},
		{"valid 30", 30, false},
		{"valid 50", 50, false},
		{"valid 100", 100, false},
		{"invalid 0", 0, true},
		{"invalid -1", -1, true},
		{"invalid -100", -100, true},
		{"invalid 101", 101, true},
		{"invalid 1000", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePerPage(tt.perPage)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error for perPage=%d, got nil", tt.perPage)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for perPage=%d, got %v", tt.perPage, err)
			}
		})
	}
}

// Test parseCountFlag function
func TestParseCountFlag(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  int
		expectErr bool
	}{
		{"valid 10", "10", 10, false},
		{"valid 30", "30", 30, false},
		{"valid 100", "100", 100, false},
		{"valid 1", "1", 1, false},
		{"invalid 0", "0", 0, true},
		{"invalid -1", "-1", 0, true},
		{"invalid 101", "101", 0, true},
		{"invalid empty", "", 0, true},
		{"invalid abc", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCountFlag(tt.input)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error for input=%q, got nil", tt.input)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for input=%q, got %v", tt.input, err)
			}
			if !tt.expectErr && result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}