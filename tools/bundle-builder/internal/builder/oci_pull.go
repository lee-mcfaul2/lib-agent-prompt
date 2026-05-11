package builder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
)

// PullServiceSchema fetches a service's mcp-schema.json blob from an OCI registry,
// verifies the content hash matches the expected source_digest, and returns the raw bytes.
func PullServiceSchema(ctx context.Context, registryRef string, expectedDigest string) ([]byte, error) {
	repo, err := remote.NewRepository(registryRef)
	if err != nil {
		return nil, fmt.Errorf("registry ref %s: %w", registryRef, err)
	}
	desc, err := oras.Resolve(ctx, repo, expectedDigest, oras.DefaultResolveOptions)
	if err != nil {
		return nil, fmt.Errorf("resolve %s@%s: %w", registryRef, expectedDigest, err)
	}
	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer rc.Close()
	buf, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	h := sha256.Sum256(buf)
	got := "sha256:" + hex.EncodeToString(h[:])
	if got != expectedDigest {
		return nil, fmt.Errorf("digest mismatch: got %s want %s", got, expectedDigest)
	}
	return buf, nil
}
