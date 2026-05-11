package agentprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	verify "github.com/lee-mcfaul2/lib-agent-prompt/pkg/verify"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
)

type Bundle struct {
	inner *verify.Bundle
	hash  string
}

type BundleLoader struct {
	registry string
	verifier *Verifier
}

func NewBundleLoader(registry string, v *Verifier) *BundleLoader {
	return &BundleLoader{registry: registry, verifier: v}
}

// Load fetches a bundle from a local tarball path OR an OCI reference.
// When BundleLoader has a registry, ref is interpreted as "<repo>:<tag>" or "<repo>@<digest>".
// When BundleLoader has no registry, ref is interpreted as a local tarball path.
func (l *BundleLoader) Load(ctx context.Context, ref string) (*Bundle, error) {
	if l.registry == "" {
		b, hash, err := verify.LoadAndHash(ref)
		if err != nil {
			return nil, err
		}
		return &Bundle{inner: b, hash: hash}, nil
	}

	repoRef := l.registry + "/" + strings.SplitN(ref, ":", 2)[0]
	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return nil, fmt.Errorf("new repository: %w", err)
	}
	if strings.HasPrefix(l.registry, "localhost") || strings.HasPrefix(l.registry, "127.0.0.1") {
		repo.PlainHTTP = true
	}
	tag := "latest"
	if parts := strings.SplitN(ref, ":", 2); len(parts) == 2 {
		tag = parts[1]
	}

	manifestDesc, err := oras.Resolve(ctx, repo, tag, oras.DefaultResolveOptions)
	if err != nil {
		return nil, fmt.Errorf("resolve %s:%s: %w", repoRef, tag, err)
	}
	rc, err := repo.Fetch(ctx, manifestDesc)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	manifestBytes, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return nil, err
	}
	var manifest struct {
		Layers []struct {
			MediaType string `json:"mediaType"`
			Digest    string `json:"digest"`
			Size      int64  `json:"size"`
		} `json:"layers"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("manifest has no layers")
	}

	// Fetch the single bundle layer (we always pack as a single-layer artifact)
	layer := manifest.Layers[0]
	layerRC, err := repo.Fetch(ctx, mkDesc(layer.MediaType, layer.Digest, layer.Size))
	if err != nil {
		return nil, fmt.Errorf("fetch layer: %w", err)
	}
	defer layerRC.Close()

	tmpFile, err := os.CreateTemp("", "agentprompt-bundle-*.tar.gz")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	if _, err := io.Copy(tmpFile, layerRC); err != nil {
		tmpFile.Close()
		return nil, err
	}
	tmpFile.Close()

	b, hash, err := verify.LoadAndHash(tmpFile.Name())
	if err != nil {
		return nil, err
	}
	return &Bundle{inner: b, hash: hash}, nil
}

func (b *Bundle) Hash() string { return b.hash }

// Prompt returns the prompt document with the given UUID, or an error if absent.
func (b *Bundle) Prompt(uuid string) (map[string]any, error) {
	for _, doc := range b.inner.Prompts {
		if doc["id"] == uuid {
			return doc, nil
		}
	}
	return nil, fmt.Errorf("prompt %s not found", uuid)
}

func (b *Bundle) Prompts() map[string]map[string]any  { return b.inner.Prompts }
func (b *Bundle) Services() map[string]map[string]any { return b.inner.Services }

func (b *Bundle) ManifestVersion() string {
	if v, ok := b.inner.Manifest["bundle_version"].(string); ok {
		return v
	}
	return ""
}

func (b *Bundle) Manifest() map[string]any { return b.inner.Manifest }

// mkDesc constructs a minimal descriptor for Fetch.
func mkDesc(mediaType, digest string, size int64) ocispec.Descriptor {
	return ocispec.Descriptor{MediaType: mediaType, Digest: digestParse(digest), Size: size}
}

func digestParse(s string) digest.Digest { return digest.Digest(s) }
