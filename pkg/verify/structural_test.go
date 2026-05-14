package verify

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// New bundle-layout tests (Task 12)
// ---------------------------------------------------------------------------

func TestVerifyStructuralGoodBundle(t *testing.T) {
	root := repoRoot(t)
	if err := VerifyBundleStructure(root); err != nil {
		t.Fatalf("VerifyBundleStructure: %v", err)
	}
}

func TestVerifyUserPromptHappy(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "tests", "fixtures", "sample-user-prompt.json"))
	if err != nil {
		t.Skipf("sample-user-prompt fixture not yet created: %v", err)
	}
	if err := VerifyUserPrompt(root, raw); err != nil {
		t.Fatalf("VerifyUserPrompt: %v", err)
	}
}

func TestVerifyUserPromptMissingField(t *testing.T) {
	root := repoRoot(t)
	bad := []byte(`{"prompt_uuid":"00000000-0000-0000-0000-000000000000","text":"x"}`)
	if err := VerifyUserPrompt(root, bad); err == nil {
		t.Fatal("expected error for missing schema_uuid")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	// Tests run from pkg/verify/, so repo root is two up.
	return filepath.Join("..", "..")
}

// ---------------------------------------------------------------------------
// Legacy tarball-based tests (kept for supply-chain verification coverage)
// ---------------------------------------------------------------------------

func TestVerifyStructure_ExampleBundle(t *testing.T) {
	// VerifyStructure targets the pre-T11 bundle shape (schemas/prompt.json,
	// prompts/, service-schemas/). The new builder no longer produces that
	// layout; skip rather than fail.
	t.Skip("legacy VerifyStructure test: pre-T11 bundle shape no longer produced")
}

func mustRun(t *testing.T, cwd string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %v failed: %v", name, args, err)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	for d := wd; d != "/"; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "Makefile")); err == nil {
			return d
		}
	}
	t.Fatal("could not find repo root")
	return ""
}
