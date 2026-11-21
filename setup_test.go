package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The D9 setup command for this project is `go mod download`.
// These tests verify the four D9 specs:
//   1. A fresh checkout is made ready by running it once.
//   2. The setup command is idempotent.
//   3. The setup command is offline-friendly (uses local module cache).
//   4. The setup command fails loudly when it cannot make a checkout ready.

// TestSetupMakesFreshCheckoutReady verifies that after running `go mod download`,
// the project can build. This is the "made ready" criterion.
func TestSetupMakesFreshCheckoutReady(t *testing.T) {
	// Run the setup command.
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = testdataDir(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod download failed: %v\noutput: %s", err, out)
	}

	// After setup, the project should build.
	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = testdataDir(t)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./... failed after setup: %v\noutput: %s", err, out)
	}
}

// TestSetupIsIdempotent verifies that running `go mod download` a second time
// on an already-ready checkout changes nothing.
func TestSetupIsIdempotent(t *testing.T) {
	dir := testdataDir(t)

	// First run: make the checkout ready.
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("first go mod download failed: %v\noutput: %s", err, out)
	}

	// Snapshot the module cache state (list of modules).
	listBefore := moduleCacheList(t)

	// Second run: should be a no-op.
	cmd = exec.Command("go", "mod", "download")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("second go mod download failed: %v\noutput: %s", err, out)
	}

	// The module cache should be unchanged.
	listAfter := moduleCacheList(t)
	if listBefore != listAfter {
		t.Errorf("module cache changed after second run:\nbefore: %s\nafter: %s", listBefore, listAfter)
	}
}

// TestSetupIsOfflineFriendly verifies that once modules are in the local cache,
// `go mod download` succeeds with GOPROXY=off (no network access).
func TestSetupIsOfflineFriendly(t *testing.T) {
	dir := testdataDir(t)

	// First: populate the cache with network access.
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("initial go mod download failed: %v\noutput: %s", err, out)
	}

	// Second: run with GOPROXY=off to simulate no network.
	// If the cache is complete, this should succeed.
	cmd = exec.Command("go", "mod", "download")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOPROXY=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod download with GOPROXY=off failed (cache incomplete?): %v\noutput: %s", err, out)
	}
}

// TestSetupFailsLoudly verifies that when a module in go.sum cannot be fetched,
// `go mod download` fails with a non-zero exit and names the missing module.
func TestSetupFailsLoudly(t *testing.T) {
	dir := t.TempDir()

	// Create a minimal go.mod that references a non-existent module.
	gomod := "module example.com/fake\n\ngo 1.25.0\n\nrequire example.com/nonexistent-module v1.0.0\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("go mod download succeeded for a non-existent module; expected failure")
	}

	// The error output should name the missing module.
	outStr := string(out)
	if !strings.Contains(outStr, "nonexistent-module") {
		t.Errorf("error output does not name the missing module:\n%s", outStr)
	}
}

// testdataDir returns the project root (directory containing go.mod).
func testdataDir(t *testing.T) string {
	t.Helper()
	// This test lives at the project root.
	return "."
}

// moduleCacheList returns a deterministic listing of the module cache for
// the modules this project depends on. Used to detect changes between runs.
func moduleCacheList(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-m", "all")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list -m all failed: %v\noutput: %s", err, out)
	}
	return string(out)
}
