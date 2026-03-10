package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// CLI E2E Tests - Test the compiled binary end-to-end
// These tests verify the CLI works correctly from a user's perspective

const binaryName = "github-activity-test"

func buildBinary(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, out)
	}
}

func cleanupBinary() {
	os.Remove(binaryName)
}

func TestCLI_NoArguments(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	// Run without arguments
	cmd := exec.Command("./" + binaryName)
	out, err := cmd.CombinedOutput()

	// Should fail with usage message
	if err == nil {
		t.Error("Expected error when no arguments provided")
	}
	output := string(out)
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage message, got: %s", output)
	}
}

func TestCLI_InvalidUsername_QueryString(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	cmd := exec.Command("./"+binaryName, "test?url=http://evil.com")
	out, err := cmd.CombinedOutput()

	// Should fail with validation error
	if err == nil {
		t.Error("Expected error for invalid username with query string")
	}
	output := string(out)
	if !strings.Contains(output, "invalid") && !strings.Contains(output, "format") {
		t.Errorf("Expected invalid format error, got: %s", output)
	}
}

func TestCLI_InvalidUsername_PathTraversal(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	cmd := exec.Command("./"+binaryName, "../../../etc/passwd")
	out, err := cmd.CombinedOutput()

	// Should fail with validation error
	if err == nil {
		t.Error("Expected error for path traversal username")
	}
	output := string(out)
	if !strings.Contains(output, "invalid") && !strings.Contains(output, "format") {
		t.Errorf("Expected invalid format error, got: %s", output)
	}
}

func TestCLI_InvalidUsername_XSS(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	cmd := exec.Command("./"+binaryName, "<script>alert(1)</script>")
	out, err := cmd.CombinedOutput()

	// Should fail with validation error
	if err == nil {
		t.Error("Expected error for XSS attempt")
	}
	output := string(out)
	if !strings.Contains(output, "invalid") && !strings.Contains(output, "format") {
		t.Errorf("Expected invalid format error, got: %s", output)
	}
}

func TestCLI_InvalidUsername_Space(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	cmd := exec.Command("./"+binaryName, "test username")
	out, err := cmd.CombinedOutput()

	// Should fail with validation error
	if err == nil {
		t.Error("Expected error for username with space")
	}
	output := string(out)
	if !strings.Contains(output, "invalid") && !strings.Contains(output, "format") {
		t.Errorf("Expected invalid format error, got: %s", output)
	}
}

func TestCLI_ValidUsername_TooLong(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	// GitHub usernames max 39 chars
	longUsername := strings.Repeat("a", 40)
	cmd := exec.Command("./"+binaryName, longUsername)
	out, err := cmd.CombinedOutput()

	// Should fail with validation error
	if err == nil {
		t.Error("Expected error for username > 39 chars")
	}
	output := string(out)
	if !strings.Contains(output, "invalid") && !strings.Contains(output, "format") {
		t.Errorf("Expected invalid format error, got: %s", output)
	}
}