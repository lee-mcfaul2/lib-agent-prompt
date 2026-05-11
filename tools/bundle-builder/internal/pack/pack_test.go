package pack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackReproducible_Deterministic(t *testing.T) {
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0o644)
	os.WriteFile(filepath.Join(src, "b.json"), []byte(`{"k":1}`), 0o644)
	os.MkdirAll(filepath.Join(src, "sub"), 0o755)
	os.WriteFile(filepath.Join(src, "sub", "c.txt"), []byte("world"), 0o644)

	dst1 := filepath.Join(t.TempDir(), "out1.tar.gz")
	dst2 := filepath.Join(t.TempDir(), "out2.tar.gz")

	d1, err := PackReproducible(src, dst1, 1700000000)
	if err != nil {
		t.Fatal(err)
	}
	d2, err := PackReproducible(src, dst2, 1700000000)
	if err != nil {
		t.Fatal(err)
	}
	if d1 != d2 {
		t.Errorf("expected reproducible digests, got %s vs %s", d1, d2)
	}
}
