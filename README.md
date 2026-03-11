# GitHub User Activity CLI

A lightweight command-line tool that fetches and displays a user's recent GitHub activity.

**Project:** https://roadmap.sh/projects/github-user-activity

## Why I Built This

I needed a quick way to check what someone was working on without opening GitHub in a browser. A simple CLI that shows recent pushes, PRs, issues, and stars — right in the terminal.

## What It Does

```bash
$ ./github-activity kamranahmedse

Pushed to kamranahmedse/claude-queue
Created comment on issue in kamranahmedse/claude-statusline
Closed pull request in kamranahmedse/slim
Created comment on issue in kamranahmedse/slim
Closed issue in kamranahmedse/slim
```

Supports these event types:
- PushEvent (commits)
- IssuesEvent (opened, closed, reopened)
- PullRequestEvent
- WatchEvent (stars)
- ForkEvent
- CreateEvent / DeleteEvent
- Comment events

## Security First

The implementation includes input validation to prevent URL injection attacks:

```go
// Validates: alphanumeric, hyphens, underscores, 1-39 chars
func isValidUsername(username string) bool {
    if username == "" || len(username) > 39 {
        return false
    }
    return validUsernameRegex.MatchString(username)
}
```

Plus defense-in-depth with `url.PathEscape()` on the API URL.

## Testing

82% test coverage with unit tests and CLI E2E tests:

```bash
$ go test -v ./...
=== RUN   TestFetchEventsInvalidUsernameFormat
--- PASS: TestFetchEventsInvalidUsernameFormat (0.00s)
=== RUN   TestCLI_InvalidUsername_XSS
--- PASS: TestCLI_InvalidUsername_XSS (0.54s)
...
PASS
ok      github-activity    6.297s
```

## Tech Stack

- Go 1.18+ (standard library only — no external dependencies)
- GitHub Events API

## How to Run

```bash
# Build
go build -o github-activity .

# Run
./github-activity <username>

# Example
./github-activity torvalds
```

## Options

```bash
# Specify number of events (1-100, default 30)
./github-activity torvalds --count 50
./github-activity torvalds -n 25
```

## Authentication (Optional)

GitHub limits unauthenticated requests to 60/hour. To increase the limit to 5,000/hour:

1. Generate a Personal Access Token at https://github.com/settings/tokens
2. Set the environment variable:

```bash
export GITHUB_TOKEN=your_token_here
./github-activity torvalds
```

If you hit the rate limit without a token, you'll see a helpful message explaining how to set one up.

## What's Next

Potential improvements:
- Filter by event type
- Structured JSON output
- Cache results

The code is ~270 lines. Clean, focused, and done.