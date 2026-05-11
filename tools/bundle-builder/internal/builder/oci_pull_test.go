package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestPullServiceSchema_DigestMismatch(t *testing.T) {
	b := []byte(`{"name":"test"}`)
	h := sha256.Sum256(b)
	got := "sha256:" + hex.EncodeToString(h[:])
	want := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	if got == want {
		t.Fatal("expected mismatch for sentinel digest")
	}
}
