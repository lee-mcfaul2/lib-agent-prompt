package agentprompt

import (
	"context"
	"fmt"

	verify "github.com/lee-mcfaul2/lib-agent-prompt/pkg/verify"
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

// Load fetches a bundle from a local tarball path OR (Task 18) an OCI reference.
func (l *BundleLoader) Load(ctx context.Context, ref string) (*Bundle, error) {
	if l.registry == "" {
		b, hash, err := verify.LoadAndHash(ref)
		if err != nil {
			return nil, err
		}
		return &Bundle{inner: b, hash: hash}, nil
	}
	return nil, fmt.Errorf("OCI fetch not yet implemented in library; see Task 18")
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

// Prompts returns every prompt in the bundle.
func (b *Bundle) Prompts() map[string]map[string]any { return b.inner.Prompts }

// Services returns every embedded service schema in the bundle.
func (b *Bundle) Services() map[string]map[string]any { return b.inner.Services }

func (b *Bundle) ManifestVersion() string {
	if v, ok := b.inner.Manifest["bundle_version"].(string); ok {
		return v
	}
	return ""
}

func (b *Bundle) Manifest() map[string]any { return b.inner.Manifest }
