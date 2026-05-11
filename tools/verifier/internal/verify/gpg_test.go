package verify

import "testing"

func TestVerifyGPGTag_BogusTag(t *testing.T) {
	err := VerifyGPGTag(GPGOptions{
		TagName: "v99.99.99-does-not-exist",
		RepoDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected failure for bogus tag in non-repo")
	}
}

func TestVerifyGPGTag_EmptyTag(t *testing.T) {
	if err := VerifyGPGTag(GPGOptions{}); err == nil {
		t.Fatal("expected failure for empty tag")
	}
}
