package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSONFile_Good(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.json")
	if err := os.WriteFile(p, []byte(`{"a":1,"b":"two"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := LoadJSONFile(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc["a"].(float64) != 1 || doc["b"].(string) != "two" {
		t.Fatalf("bad doc: %v", doc)
	}
}

func TestLoadJSONFile_Malformed(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.json")
	if err := os.WriteFile(p, []byte(`{not json`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadJSONFile(p); err == nil {
		t.Fatal("expected error on malformed JSON")
	}
}

func TestLoadAllJSON_Recursive(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "a.json"), []byte(`{"x":1}`), 0o644)
	os.WriteFile(filepath.Join(dir, "sub", "b.json"), []byte(`{"y":2}`), 0o644)
	os.WriteFile(filepath.Join(dir, "skip.txt"), []byte("ignore me"), 0o644)

	docs, err := LoadAllJSON(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("want 2 docs, got %d", len(docs))
	}
	if _, ok := docs["a.json"]; !ok {
		t.Fatal("missing a.json")
	}
	if _, ok := docs[filepath.Join("sub", "b.json")]; !ok {
		t.Fatal("missing sub/b.json")
	}
}
