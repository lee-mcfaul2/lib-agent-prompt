package agentprompt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	verify "github.com/lee-mcfaul2/lib-agent-prompt/pkg/verify"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
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

	repo, err := remote.NewRepository(l.registry + "/" + strings.SplitN(ref, ":", 2)[0])
	if err != nil {
		return nil, fmt.Errorf("new repository: %w", err)
	}
	tag := "latest"
	if parts := strings.SplitN(ref, ":", 2); len(parts) == 2 {
		tag = parts[1]
	}

	tmpDir, err := os.MkdirTemp("", "agentprompt-fetch-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	fs, err := file.New(tmpDir)
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	if _, err := oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions); err != nil {
		return nil, fmt.Errorf("oras copy: %w", err)
	}

	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".gz") {
			b, hash, err := verify.LoadAndHash(filepath.Join(tmpDir, name))
			if err != nil {
				return nil, err
			}
			return &Bundle{inner: b, hash: hash}, nil
		}
	}
	return nil, fmt.Errorf("no tar.gz layer found in fetched artifact at %s", l.registry+"/"+ref)
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
