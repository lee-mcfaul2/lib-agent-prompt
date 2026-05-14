package agentprompt

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestCompileRequestValidator(t *testing.T) {
	b, _ := LoadFromDir(filepath.Join("..", ""))
	v, err := b.CompileRequestValidator("kb", "search")
	if err != nil {
		t.Fatalf("CompileRequestValidator: %v", err)
	}
	doc := map[string]any{"q": "hello"}
	raw, _ := json.Marshal(doc)
	if err := v.Validate(raw); err != nil {
		t.Errorf("Validate happy: %v", err)
	}
	bad, _ := json.Marshal(map[string]any{})
	if err := v.Validate(bad); err == nil {
		t.Error("expected error for missing 'q'")
	}
}

func TestCompileResponseValidator(t *testing.T) {
	b, _ := LoadFromDir(filepath.Join("..", ""))
	v, err := b.CompileResponseValidator("kb", "search")
	if err != nil {
		t.Fatalf("CompileResponseValidator: %v", err)
	}
	doc := map[string]any{"rows": []any{map[string]any{"id": "1", "title": "x"}}}
	raw, _ := json.Marshal(doc)
	if err := v.Validate(raw); err != nil {
		t.Errorf("Validate happy: %v", err)
	}
}

func TestCompileUnknownTool(t *testing.T) {
	b, _ := LoadFromDir(filepath.Join("..", ""))
	if _, err := b.CompileRequestValidator("nope", "absent"); err == nil {
		t.Fatal("expected error")
	}
}
