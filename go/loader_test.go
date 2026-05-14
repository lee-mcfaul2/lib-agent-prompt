package agentprompt

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBundleSchemasKeys(t *testing.T) {
	b, err := LoadFromDir(filepath.Join("..", ""))
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	for _, key := range []string{
		"kb/search.request", "kb/search.response",
		"kb/fetch.request", "kb/fetch.response",
		"audit_db/search.request", "audit_db/search.response",
		"customer/search_customer.request", "customer/search_customer.response",
	} {
		if _, ok := b.Schemas[key]; !ok {
			t.Errorf("missing schema key: %s", key)
		}
	}
}

func TestBundleEnvelopeSchemas(t *testing.T) {
	b, err := LoadFromDir(filepath.Join("..", ""))
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if len(b.UserPromptSchema()) == 0 {
		t.Error("UserPromptSchema empty")
	}
	if len(b.FinalResponseSchema()) == 0 {
		t.Error("FinalResponseSchema empty")
	}
	if len(b.ToolResultSchema()) == 0 {
		t.Error("ToolResultSchema empty")
	}
}

func TestBundleAccessors(t *testing.T) {
	b, _ := LoadFromDir(filepath.Join("..", ""))
	req, ok := b.RequestSchema("kb", "search")
	if !ok || len(req) == 0 {
		t.Error("RequestSchema kb.search missing")
	}
	resp, ok := b.ResponseSchema("kb", "search")
	if !ok || len(resp) == 0 {
		t.Error("ResponseSchema kb.search missing")
	}
	_, ok = b.RequestSchema("nope", "absent")
	if ok {
		t.Error("RequestSchema returned true for unknown tool")
	}
}

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
