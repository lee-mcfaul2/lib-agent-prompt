package push

import (
	"context"
	"fmt"
	"os"
	"strings"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
)

const BundleMediaType = "application/vnd.ai-security.bundle.v1+tar"

// Push uploads tarballPath as a single-layer OCI artifact to ref (registry/repo:tag).
// Returns the descriptor digest.
func Push(ctx context.Context, tarballPath, ref string) (string, error) {
	if _, err := os.Stat(tarballPath); err != nil {
		return "", fmt.Errorf("stat tarball: %w", err)
	}
	fs, err := file.New("")
	if err != nil {
		return "", err
	}
	defer fs.Close()

	desc, err := fs.Add(ctx, tarballPath, BundleMediaType, "")
	if err != nil {
		return "", fmt.Errorf("file.Add: %w", err)
	}
	manifestDesc, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1, BundleMediaType,
		oras.PackManifestOptions{Layers: []v1.Descriptor{desc}})
	if err != nil {
		return "", fmt.Errorf("PackManifest: %w", err)
	}
	if err := fs.Tag(ctx, manifestDesc, "latest"); err != nil {
		return "", err
	}

	repo, err := remote.NewRepository(ref)
	if err != nil {
		return "", err
	}
	// Use plain HTTP for localhost/127.0.0.1 registries (e.g., local zot in tests).
	if strings.HasPrefix(ref, "localhost") || strings.HasPrefix(ref, "127.0.0.1") {
		repo.PlainHTTP = true
	}
	pushed, err := oras.Copy(ctx, fs, "latest", repo, "latest", oras.DefaultCopyOptions)
	if err != nil {
		return "", err
	}
	return pushed.Digest.String(), nil
}
