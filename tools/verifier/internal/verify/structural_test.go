package verify

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestVerifyStructure_ExampleBundle(t *testing.T) {
	repoRoot := findRepoRoot(t)
	out := filepath.Join(repoRoot, "out", "test-bundle")
	tarball := out + ".tar.gz"

	mustRun(t, repoRoot,
		"go", "run", "./tools/bundle-builder", "build",
		"--schema-lib", "./schemas",
		"--prompts", "./prompts/example",
		"--services", "./schemas/service-references",
		"--output", out,
		"--version", "0.1.0-test",
		"--allow-placeholder",
	)
	mustRun(t, repoRoot, "go", "run", "./tools/bundle-builder", "pack", out)

	b, hash, err := LoadAndHash(tarball)
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" {
		t.Fatal("empty hash")
	}
	if err := VerifyStructure(b); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
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
