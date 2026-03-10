package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// GitHub username validation regex
// Usernames: 1-39 chars, alphanumeric, hyphens, underscores, cannot start with -
var validUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$|^[a-zA-Z0-9]$`)

// GitHub API types
type GitHubEvent struct {
	Type      string  `json:"type"`
	Repo      Repo    `json:"repo"`
	CreatedAt string  `json:"created_at"`
	Payload   Payload `json:"payload"`
}

type Repo struct {
	Name string `json:"name"`
}

type Payload struct {
	Commits     []Commit     `json:"commits"`
	Action      string       `json:"action"`
	Issue       Issue       `json:"issue"`
	RefType     string       `json:"ref_type"`
	Ref         string       `json:"ref"`
	Forkee      Forkee       `json:"forkee"`
	PullRequest PullRequest  `json:"pull_request"`
}

type Commit struct {
	Message string `json:"message"`
}

type Issue struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Forkee struct {
	FullName string `json:"full_name"`
}

type PullRequest struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: github-activity <username>")
		os.Exit(1)
	}

	username := os.Args[1]
	events, err := fetchEvents(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(events) == 0 {
		fmt.Println("No recent activity found.")
		return
	}

	for _, event := range events {
		fmt.Println(formatEvent(event))
	}
}

// HTTPClient interface for making HTTP requests (allows mocking)
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// defaultClient is the default HTTP client used in production
var defaultClient HTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// setClient allows injecting a custom client (used for testing)
func setClient(client HTTPClient) {
	if client != nil {
		defaultClient = client
	}
}

func fetchEvents(username string) ([]GitHubEvent, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	// Validate username format to prevent URL injection
	// GitHub usernames: alphanumeric, hyphens, underscores, max 39 chars
	if !isValidUsername(username) {
		return nil, fmt.Errorf("invalid username format: %s", username)
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/events", url.PathEscape(username))

	resp, err := defaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user %s not found", username)
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

func formatEvent(event GitHubEvent) string {
	switch event.Type {
	case "PushEvent":
		count := len(event.Payload.Commits)
		if count == 0 {
			return fmt.Sprintf("Pushed to %s", event.Repo.Name)
		}
		if count == 1 {
			return fmt.Sprintf("Pushed %d commit to %s", count, event.Repo.Name)
		}
		return fmt.Sprintf("Pushed %d commits to %s", count, event.Repo.Name)

	case "IssuesEvent":
		action := event.Payload.Action
		if action == "" {
			action = "unknown"
		}
		if action == "opened" {
			return fmt.Sprintf("Opened a new issue in %s", event.Repo.Name)
		}
		return fmt.Sprintf("%s issue in %s", capitalize(action), event.Repo.Name)

	case "WatchEvent":
		return fmt.Sprintf("Starred %s", event.Repo.Name)

	case "CreateEvent":
		refType := event.Payload.RefType
		ref := event.Payload.Ref
		if ref == "" {
			return fmt.Sprintf("Created %s in %s", refType, event.Repo.Name)
		}
		return fmt.Sprintf("Created %s %s in %s", refType, ref, event.Repo.Name)

	case "DeleteEvent":
		refType := event.Payload.RefType
		ref := event.Payload.Ref
		return fmt.Sprintf("Deleted %s %s in %s", refType, ref, event.Repo.Name)

	case "ForkEvent":
		if event.Payload.Forkee.FullName != "" {
			return fmt.Sprintf("Forked %s to %s", event.Repo.Name, event.Payload.Forkee.FullName)
		}
		return fmt.Sprintf("Forked %s", event.Repo.Name)

	case "PullRequestEvent":
		action := event.Payload.Action
		return fmt.Sprintf("%s pull request in %s", capitalize(action), event.Repo.Name)

	case "IssueCommentEvent":
		action := event.Payload.Action
		return fmt.Sprintf("%s comment on issue in %s", capitalize(action), event.Repo.Name)

	case "CommitCommentEvent":
		return fmt.Sprintf("Commented on commit in %s", event.Repo.Name)

	case "PullRequestReviewEvent":
		action := event.Payload.Action
		return fmt.Sprintf("%s pull request review in %s", capitalize(action), event.Repo.Name)

	case "ReleaseEvent":
		action := event.Payload.Action
		return fmt.Sprintf("%s release in %s", capitalize(action), event.Repo.Name)

	case "PullRequestReviewCommentEvent":
		action := event.Payload.Action
		return fmt.Sprintf("%s pull request comment in %s", capitalize(action), event.Repo.Name)

	default:
		// Generic format for unknown event types
		return fmt.Sprintf("%s in %s", event.Type, event.Repo.Name)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// isValidUsername validates GitHub username format
// GitHub usernames: alphanumeric, hyphens, underscores, 1-39 chars, cannot start with -
func isValidUsername(username string) bool {
	if username == "" || len(username) > 39 {
		return false
	}
	return validUsernameRegex.MatchString(username)
}