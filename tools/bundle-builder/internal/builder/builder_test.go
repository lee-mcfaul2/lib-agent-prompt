package builder

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuild_ExampleBundle(t *testing.T) {
	repoRoot := findRepoRoot(t)
	out := t.TempDir()

	opts := Options{
		SchemaLib:        filepath.Join(repoRoot, "schemas"),
		Prompts:          filepath.Join(repoRoot, "prompts", "example"),
		Services:         filepath.Join(repoRoot, "schemas", "service-references"),
		Output:           out,
		Version:          "0.1.0-test",
		SchemaLibVersion: "0.1.0",
		BuildTime:        time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC),
		SourceCommit:     "deadbeef",
		BuilderID:        "test-builder",
		AllowPlaceholder: true,
	}

	if err := Build(context.Background(), opts); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(out, "bundle-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["bundle_version"] != "0.1.0-test" {
		t.Errorf("bundle_version wrong: %v", m["bundle_version"])
	}
	if prompts := m["prompts"].([]any); len(prompts) == 0 {
		t.Error("expected at least one prompt in manifest")
	}
	if services := m["services"].([]any); len(services) == 0 {
		t.Error("expected at least one service in manifest")
	}

	if _, err := os.Stat(filepath.Join(out, "schemas", "prompt.json")); err != nil {
		t.Errorf("schemas/prompt.json not snapshotted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "service-schemas", "postgresql-service.json")); err != nil {
		t.Errorf("service-schemas/postgresql-service.json missing: %v", err)
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
