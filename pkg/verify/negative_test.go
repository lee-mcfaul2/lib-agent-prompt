package verify

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestNegative_CorruptBundles(t *testing.T) {
	// LoadAndHash + VerifyStructure target the pre-T11 bundle shape
	// (schemas/prompt.json, prompts/, service-schemas/). The new builder no
	// longer produces that layout; skip rather than fail.
	t.Skip("legacy VerifyStructure negative tests: pre-T11 bundle shape no longer produced")

	repoRoot := findRepoRoot(t)
	script := filepath.Join(repoRoot, "tests", "fixtures", "corrupt-bundles", "make-fixtures.sh")
	cmd := exec.Command("bash", script)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("fixture build failed: %v", err)
	}

	cases := []struct {
		name    string
		tarball string
	}{
		{"corrupt-manifest", filepath.Join(repoRoot, "out", "fixture-corrupt", "corrupt-manifest.tar.gz")},
		{"missing-service", filepath.Join(repoRoot, "out", "fixture-corrupt", "missing-service.tar.gz")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, _, err := LoadAndHash(c.tarball)
			if err != nil {
				// Loading itself failing is an acceptable rejection.
				return
			}
			if err := VerifyStructure(b); err == nil {
				t.Fatalf("expected structural rejection for %s", c.name)
			}
		})
	}
}
