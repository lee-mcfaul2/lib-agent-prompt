package agentprompt

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestValidateIterationOutput_ServiceCall(t *testing.T) {
	bundle := buildAndLoad(t)

	var promptID string
	for _, doc := range bundle.inner.Prompts {
		promptID = doc["id"].(string)
		break
	}

	good := map[string]any{
		"service_calls": []map[string]any{
			{"id": "call-postgres-get-users", "args": map[string]any{}},
		},
		"terminate": nil,
	}
	b, _ := json.Marshal(good)
	if err := bundle.ValidateIterationOutput(promptID, b); err != nil {
		t.Fatalf("expected pass, got: %v", err)
	}

	bad := map[string]any{
		"service_calls": []map[string]any{
			{"id": "call-unknown-service", "args": map[string]any{}},
		},
		"terminate": nil,
	}
	bb, _ := json.Marshal(bad)
	if err := bundle.ValidateIterationOutput(promptID, bb); err == nil {
		t.Fatal("expected rejection of unknown service call")
	}
}

func buildAndLoad(t *testing.T) *Bundle {
	t.Helper()
	repoRoot := findRepoRoot(t)
	out := filepath.Join(repoRoot, "out", "go-iter-test")
	tarball := out + ".tar.gz"
	c1 := exec.Command("go", "run", "./tools/bundle-builder", "build",
		"--schema-lib", "./schemas",
		"--prompts", "./prompts/example",
		"--services", "./schemas/service-references",
		"--output", out,
		"--version", "0.1.0-iter-test",
		"--allow-placeholder")
	c1.Dir = repoRoot
	c1.Stdout = os.Stderr
	c1.Stderr = os.Stderr
	if err := c1.Run(); err != nil {
		t.Fatal(err)
	}
	c2 := exec.Command("go", "run", "./tools/bundle-builder", "pack", out)
	c2.Dir = repoRoot
	c2.Stdout = os.Stderr
	c2.Stderr = os.Stderr
	if err := c2.Run(); err != nil {
		t.Fatal(err)
	}
	loader := NewBundleLoader("", NewVerifier(TrustRelaxed, "", ""))
	b, err := loader.Load(context.Background(), tarball)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
