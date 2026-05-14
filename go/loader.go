package agentprompt

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
)

// Bundle is a loaded lib-agent-prompt bundle. Held in memory as raw bytes;
// the consumer compiles validators on demand from these bytes (see validators.go).
type Bundle struct {
	// Digest is "sha256:<hex>" of the canonicalized schema set. Deterministic
	// across machines if the source tree is byte-identical (after JCS).
	Digest string

	// Schemas maps "<mcp>/<tool>.<direction>" → raw JSON Schema bytes.
	// Direction is "request" or "response". Example key: "kb/search.request".
	Schemas map[string][]byte

	userPrompt    []byte
	finalResponse []byte
	toolResult    []byte
	manifest      []byte

	// legacy fields populated when loading from a packed OCI/tarball bundle
	legacyManifest map[string]any
	legacyPrompts  map[string]map[string]any
	legacyServices map[string]map[string]any
	hash           string
}

// UserPromptSchema returns the raw bytes of schemas/user-prompt.json.
func (b *Bundle) UserPromptSchema() []byte { return b.userPrompt }

// FinalResponseSchema returns the raw bytes of schemas/final-response.json.
func (b *Bundle) FinalResponseSchema() []byte { return b.finalResponse }

// ToolResultSchema returns the raw bytes of schemas/tool-result.json.
func (b *Bundle) ToolResultSchema() []byte { return b.toolResult }

// ManifestJSON returns the raw bytes of bundle-manifest.json.
func (b *Bundle) ManifestJSON() []byte { return b.manifest }

// RequestSchema returns the raw JSON Schema bytes for mcp.tool's request side.
func (b *Bundle) RequestSchema(mcp, tool string) ([]byte, bool) {
	v, ok := b.Schemas[mcp+"/"+tool+".request"]
	return v, ok
}

// ResponseSchema returns the raw JSON Schema bytes for mcp.tool's response side.
func (b *Bundle) ResponseSchema(mcp, tool string) ([]byte, bool) {
	v, ok := b.Schemas[mcp+"/"+tool+".response"]
	return v, ok
}

// Hash returns the sha256 digest of the bundle tarball (populated for OCI/tarball loads).
func (b *Bundle) Hash() string { return b.hash }

// ManifestVersion returns bundle_version from the bundle manifest JSON (legacy OCI compat).
func (b *Bundle) ManifestVersion() string {
	if b.legacyManifest != nil {
		if v, ok := b.legacyManifest["bundle_version"].(string); ok {
			return v
		}
	}
	return ""
}

// Manifest returns the parsed bundle manifest map (legacy OCI compat).
func (b *Bundle) Manifest() map[string]any { return b.legacyManifest }

// Prompts returns the parsed prompt documents (legacy OCI compat).
func (b *Bundle) Prompts() map[string]map[string]any { return b.legacyPrompts }

// Services returns the parsed service documents (legacy OCI compat).
func (b *Bundle) Services() map[string]map[string]any { return b.legacyServices }

// Prompt returns the prompt document with the given UUID, or an error if absent (legacy OCI compat).
func (b *Bundle) Prompt(uuid string) (map[string]any, error) {
	for _, doc := range b.legacyPrompts {
		if doc["id"] == uuid {
			return doc, nil
		}
	}
	return nil, fmt.Errorf("prompt %s not found", uuid)
}

// LoadFromDir reads a bundle source tree from the given directory. The
// directory must contain schemas/user-prompt.json, schemas/final-response.json,
// schemas/tool-result.json, and schemas/services/.
func LoadFromDir(root string) (*Bundle, error) {
	return loadFromFS(os.DirFS(root), ".")
}

// LoadFromFS reads a bundle from an arbitrary fs.FS rooted at root.
func LoadFromFS(fsys fs.FS, root string) (*Bundle, error) {
	return loadFromFS(fsys, root)
}

func loadFromFS(fsys fs.FS, root string) (*Bundle, error) {
	b := &Bundle{Schemas: map[string][]byte{}}

	join := func(rel string) string {
		if root == "." {
			return rel
		}
		return filepath.Join(root, rel)
	}

	readOne := func(rel string, dst *[]byte) error {
		raw, err := fs.ReadFile(fsys, join(rel))
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		*dst = raw
		return nil
	}
	if err := readOne("schemas/user-prompt.json", &b.userPrompt); err != nil {
		return nil, err
	}
	if err := readOne("schemas/final-response.json", &b.finalResponse); err != nil {
		return nil, err
	}
	if err := readOne("schemas/tool-result.json", &b.toolResult); err != nil {
		return nil, err
	}
	// Manifest is optional in a source tree (only present after build);
	// load if it exists, skip otherwise.
	if raw, err := fs.ReadFile(fsys, join("bundle-manifest.json")); err == nil {
		b.manifest = raw
	}

	servicesRoot := join("schemas/services")
	if err := fs.WalkDir(fsys, servicesRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".request.json") && !strings.HasSuffix(p, ".response.json") {
			return nil
		}
		raw, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		// Strip the services root prefix + separator to get "<mcp>/<tool>.<direction>"
		prefix := servicesRoot + "/"
		rel := strings.TrimPrefix(p, prefix)
		rel = strings.TrimSuffix(rel, ".json")
		b.Schemas[rel] = raw
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk services: %w", err)
	}
	if len(b.Schemas) == 0 {
		return nil, fmt.Errorf("no per-tool schemas found under %s", servicesRoot)
	}

	keys := make([]string, 0, len(b.Schemas))
	for k := range b.Schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{0})
		h.Write(b.Schemas[k])
		h.Write([]byte{0})
	}
	h.Write([]byte("user-prompt"))
	h.Write([]byte{0})
	h.Write(b.userPrompt)
	h.Write([]byte("final-response"))
	h.Write([]byte{0})
	h.Write(b.finalResponse)
	h.Write([]byte("tool-result"))
	h.Write([]byte{0})
	h.Write(b.toolResult)
	b.Digest = "sha256:" + hex.EncodeToString(h.Sum(nil))

	return b, nil
}

// BundleLoader fetches and loads bundles from OCI registries or local tarballs.
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
		return loadFromTarball(ref)
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
	var ociManifest struct {
		Layers []struct {
			MediaType string `json:"mediaType"`
			Digest    string `json:"digest"`
			Size      int64  `json:"size"`
		} `json:"layers"`
	}
	if err := json.Unmarshal(manifestBytes, &ociManifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if len(ociManifest.Layers) == 0 {
		return nil, fmt.Errorf("manifest has no layers")
	}

	// Fetch the single bundle layer (we always pack as a single-layer artifact)
	layer := ociManifest.Layers[0]
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

	return loadFromTarball(tmpFile.Name())
}

// loadFromTarball extracts a packed bundle tarball to a temp directory and
// loads it with LoadFromDir. It also populates legacy fields (Manifest, Prompts, Services)
// for backward compatibility.
func loadFromTarball(tarballPath string) (*Bundle, error) {
	// Compute hash of the tarball itself.
	hash, err := sha256OfFile(tarballPath)
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "agentprompt-extract-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	legacyManifest, legacyPrompts, legacyServices, err := extractTarball(tarballPath, tmpDir)
	if err != nil {
		return nil, err
	}

	b, err := LoadFromDir(tmpDir)
	if err != nil {
		return nil, err
	}
	b.hash = hash
	b.legacyManifest = legacyManifest
	b.legacyPrompts = legacyPrompts
	b.legacyServices = legacyServices
	return b, nil
}

// extractTarball extracts a bundle tarball into destDir and returns the legacy
// parsed data (manifest, prompts, services).
func extractTarball(tarballPath, destDir string) (
	manifest map[string]any,
	prompts map[string]map[string]any,
	services map[string]map[string]any,
	err error,
) {
	f, err := os.Open(tarballPath)
	if err != nil {
		return nil, nil, nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, nil, nil, err
	}
	defer gz.Close()

	prompts = map[string]map[string]any{}
	services = map[string]map[string]any{}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, nil, nil, err
		}

		// Write into destDir for LoadFromDir.
		dest := filepath.Join(destDir, hdr.Name)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return nil, nil, nil, err
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return nil, nil, nil, err
		}

		// Also parse legacy fields.
		switch {
		case hdr.Name == "bundle-manifest.json":
			if err := json.Unmarshal(data, &manifest); err != nil {
				return nil, nil, nil, fmt.Errorf("parse manifest: %w", err)
			}
		case strings.HasPrefix(hdr.Name, "prompts/"):
			var doc map[string]any
			if err := json.Unmarshal(data, &doc); err != nil {
				return nil, nil, nil, fmt.Errorf("parse %s: %w", hdr.Name, err)
			}
			prompts[hdr.Name] = doc
		case strings.HasPrefix(hdr.Name, "service-schemas/"):
			var doc map[string]any
			if err := json.Unmarshal(data, &doc); err != nil {
				return nil, nil, nil, fmt.Errorf("parse %s: %w", hdr.Name, err)
			}
			name := strings.TrimSuffix(filepath.Base(hdr.Name), ".json")
			services[name] = doc
		}
	}
	return manifest, prompts, services, nil
}

func sha256OfFile(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// mkDesc constructs a minimal descriptor for Fetch.
func mkDesc(mediaType, dgst string, size int64) ocispec.Descriptor {
	return ocispec.Descriptor{MediaType: mediaType, Digest: digest.Digest(dgst), Size: size}
}
