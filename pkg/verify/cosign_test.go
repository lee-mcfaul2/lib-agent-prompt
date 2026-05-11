package verify

import (
	"path/filepath"
	"testing"
)

func TestVerifyCosign_MissingBundle(t *testing.T) {
	err := VerifyCosign(CosignOptions{
		BundlePath:               filepath.Join(t.TempDir(), "nonexistent.bundle"),
		CertificateIdentityRegex: "irrelevant",
	}, "")
	if err == nil {
		t.Fatal("expected error when bundle path missing")
	}
}

func TestVerifyCosign_NoBundlePath(t *testing.T) {
	if err := VerifyCosign(CosignOptions{}, ""); err == nil {
		t.Fatal("expected error when bundle path empty")
	}
}

func TestVerifyCosign_NoRegex(t *testing.T) {
	dir := t.TempDir()
	if err := VerifyCosign(CosignOptions{BundlePath: filepath.Join(dir, "x.bundle")}, ""); err == nil {
		t.Fatal("expected error when regex empty")
	}
}
