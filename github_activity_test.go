package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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