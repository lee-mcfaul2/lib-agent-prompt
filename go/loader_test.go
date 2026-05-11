package agentprompt

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLoad_ExampleBundle(t *testing.T) {
	repoRoot := findRepoRoot(t)
	out := filepath.Join(repoRoot, "out", "go-test-bundle")
	tarball := out + ".tar.gz"

	must(t, repoRoot, "go", "run", "./tools/bundle-builder", "build",
		"--schema-lib", "./schemas",
		"--prompts", "./prompts/example",
		"--services", "./schemas/service-references",
		"--output", out,
		"--version", "0.1.0-go-test",
		"--allow-placeholder",
	)
	must(t, repoRoot, "go", "run", "./tools/bundle-builder", "pack", out)

	v := NewVerifier(TrustRelaxed, "", "")
	loader := NewBundleLoader("", v)
	b, err := loader.Load(context.Background(), tarball)
	if err != nil {
		t.Fatal(err)
	}
	if b.ManifestVersion() != "0.1.0-go-test" {
		t.Errorf("unexpected version: %s", b.ManifestVersion())
	}
}

func must(t *testing.T, cwd, name string, args ...string) {
	t.Helper()
	c := exec.Command(name, args...)
	c.Dir = cwd
	c.Stdout = os.Stderr
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		t.Fatalf("%s %v: %v", name, args, err)
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
	t.Fatal("repo root not found")
	return ""
}
