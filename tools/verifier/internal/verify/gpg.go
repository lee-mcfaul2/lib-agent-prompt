package verify

import (
	"bytes"
	"fmt"
	"os/exec"
)

type GPGOptions struct {
	TagName string
	RepoDir string
}

// VerifyGPGTag invokes `git verify-tag <tag>` from the given repo dir.
// Assumes the verifier's GPG keyring contains the maintainer public keys.
func VerifyGPGTag(opts GPGOptions) error {
	if opts.TagName == "" {
		return fmt.Errorf("tag name required")
	}
	cmd := exec.Command("git", "-C", opts.RepoDir, "verify-tag", opts.TagName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git verify-tag %s: %w: %s", opts.TagName, err, stderr.String())
	}
	return nil
}
