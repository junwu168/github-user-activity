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
	if !strings.Contains(output, "usage:") {
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

// ============== Pagination E2E Tests ==============

func TestCLI_CountFlag_Valid(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	// Valid count should not cause validation error
	cmd := exec.Command("./"+binaryName, "-count", "10", "testuser")
	out, err := cmd.CombinedOutput()

	// Either succeeds with data or fails with user not found (network error)
	// Should not fail with "count must be between" error
	output := string(out)
	if strings.Contains(output, "count must be between") {
		t.Errorf("Expected valid count, got error: %s", output)
	}
	if err != nil && !strings.Contains(output, "not found") && !strings.Contains(output, "Error") {
		// Network errors are acceptable, but validation errors are not
		t.Logf("Command output: %s", output)
	}
}

func TestCLI_CountFlag_Short(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	// Short flag -n should also work
	cmd := exec.Command("./"+binaryName, "-n", "50", "testuser")
	out, err := cmd.CombinedOutput()

	output := string(out)
	if strings.Contains(output, "count must be between") {
		t.Errorf("Expected valid count with -n flag, got error: %s", output)
	}
	_ = err // May fail with network/user not found
}

func TestCLI_CountFlag_Invalid_Zero(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	cmd := exec.Command("./"+binaryName, "-count", "0", "testuser")
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for count=0")
	}
	output := string(out)
	if !strings.Contains(output, "count must be between") {
		t.Errorf("Expected count validation error, got: %s", output)
	}
}

func TestCLI_CountFlag_Invalid_Negative(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	cmd := exec.Command("./"+binaryName, "-count", "-5", "testuser")
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for negative count")
	}
	output := string(out)
	if !strings.Contains(output, "count must be between") && !strings.Contains(output, "invalid") {
		t.Errorf("Expected count validation error, got: %s", output)
	}
}

func TestCLI_CountFlag_Invalid_Over100(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	cmd := exec.Command("./"+binaryName, "-count", "101", "testuser")
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for count > 100")
	}
	output := string(out)
	if !strings.Contains(output, "count must be between") {
		t.Errorf("Expected count validation error, got: %s", output)
	}
}

func TestCLI_CountFlag_Invalid_NonNumber(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	cmd := exec.Command("./"+binaryName, "-count", "abc", "testuser")
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for non-numeric count")
	}
	output := string(out)
	if !strings.Contains(output, "invalid") && !strings.Contains(output, "count") {
		t.Errorf("Expected count parsing error, got: %s", output)
	}
}

func TestCLI_CountFlag_Default(t *testing.T) {
	buildBinary(t)
	defer cleanupBinary()

	// Default count should work (uses 30)
	cmd := exec.Command("./"+binaryName, "testuser")
	out, err := cmd.CombinedOutput()

	output := string(out)
	// Should either get events or "not found" error
	if err != nil && !strings.Contains(output, "not found") && !strings.Contains(output, "Error") {
		t.Errorf("Unexpected error: %s", output)
	}
}