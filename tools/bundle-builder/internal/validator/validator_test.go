package validator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSchemaSelfConsistent_Repo(t *testing.T) {
	repoRoot := findRepoRoot(t)
	schemaRoot := filepath.Join(repoRoot, "schemas")
	c, err := Compiler(schemaRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		"shared/uuid.json",
		"shared/error-envelope.json",
		"shared/trace-context.json",
		"shared/pii-types.json",
		"external/mcp-prompt.json",
		"allowed-response/service-call.json",
		"allowed-response/terminate.json",
		"service-reference.json",
		"prompt.json",
		"bundle-manifest.json",
		"iteration-output.json",
		"response-envelope.json",
	} {
		if _, err := c.Compile(filepath.Join(schemaRoot, p)); err != nil {
			t.Errorf("compile %s: %v", p, err)
		}
	}
}

func TestValidatePrompt_Example(t *testing.T) {
	repoRoot := findRepoRoot(t)
	schemaRoot := filepath.Join(repoRoot, "schemas")

	matches, err := filepath.Glob(filepath.Join(repoRoot, "prompts", "example", "*.json"))
	if err != nil || len(matches) == 0 {
		t.Skip("no example prompt found")
	}
	// Pick the UUID-named file (skip README.md, etc., which Glob will not match for *.json anyway).
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePrompt(schemaRoot, doc); err != nil {
		t.Fatalf("example prompt failed validation: %v", err)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	for d := wd; d != "/"; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil && filepath.Base(d) == "bundle-builder" {
			return filepath.Join(d, "..", "..")
		}
	}
	t.Fatal("could not find repo root")
	return ""
}
