package verify

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

type Bundle struct {
	Manifest map[string]any
	Prompts  map[string]map[string]any
	Services map[string]map[string]any
	Schemas  map[string][]byte
}

// LoadAndHash reads a packed tarball, returns sha256, and parses contents into Bundle.
func LoadAndHash(tarballPath string) (*Bundle, string, error) {
	hash, err := sha256OfFile(tarballPath)
	if err != nil {
		return nil, "", err
	}
	b, err := extractBundle(tarballPath)
	if err != nil {
		return nil, "", err
	}
	return b, hash, nil
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

func extractBundle(tarballPath string) (*Bundle, error) {
	f, err := os.Open(tarballPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	b := &Bundle{
		Prompts:  map[string]map[string]any{},
		Services: map[string]map[string]any{},
		Schemas:  map[string][]byte{},
	}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		switch {
		case hdr.Name == "bundle-manifest.json":
			if err := json.Unmarshal(data, &b.Manifest); err != nil {
				return nil, fmt.Errorf("parse manifest: %w", err)
			}
		case strings.HasPrefix(hdr.Name, "prompts/"):
			var doc map[string]any
			if err := json.Unmarshal(data, &doc); err != nil {
				return nil, fmt.Errorf("parse %s: %w", hdr.Name, err)
			}
			b.Prompts[hdr.Name] = doc
		case strings.HasPrefix(hdr.Name, "service-schemas/"):
			var doc map[string]any
			if err := json.Unmarshal(data, &doc); err != nil {
				return nil, fmt.Errorf("parse %s: %w", hdr.Name, err)
			}
			name := strings.TrimSuffix(filepath.Base(hdr.Name), ".json")
			b.Services[name] = doc
		case strings.HasPrefix(hdr.Name, "schemas/"):
			b.Schemas[hdr.Name] = data
		}
	}
	if b.Manifest == nil {
		return nil, fmt.Errorf("bundle-manifest.json missing")
	}
	return b, nil
}

// VerifyStructure runs the structural checks: manifest validates, every prompt validates, services consistency.
func VerifyStructure(b *Bundle) error {
	c := jsonschema.NewCompiler()
	c.Draft = jsonschema.Draft2020

	tmp, err := os.MkdirTemp("", "verify-schemas-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	for rel, data := range b.Schemas {
		out := filepath.Join(tmp, rel)
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return err
		}
	}

	c.LoadURL = func(s string) (io.ReadCloser, error) {
		const prefix = "https://ai-security.io/schemas/"
		if strings.HasPrefix(s, prefix) {
			rel := strings.TrimPrefix(s, prefix)
			return jsonschema.LoadURL("file://" + filepath.Join(tmp, "schemas", rel))
		}
		if strings.HasPrefix(s, "file://") {
			return jsonschema.LoadURL(s)
		}
		return nil, fmt.Errorf("refusing external URL: %s", s)
	}

	manifestSchema, err := c.Compile(filepath.Join(tmp, "schemas", "bundle-manifest.json"))
	if err != nil {
		return fmt.Errorf("compile manifest schema: %w", err)
	}
	if err := manifestSchema.Validate(b.Manifest); err != nil {
		return fmt.Errorf("manifest validation: %w", err)
	}

	promptSchema, err := c.Compile(filepath.Join(tmp, "schemas", "prompt.json"))
	if err != nil {
		return fmt.Errorf("compile prompt schema: %w", err)
	}

	manifestServices := map[string]bool{}
	for _, s := range b.Manifest["services"].([]any) {
		sm := s.(map[string]any)
		name := sm["name"].(string)
		manifestServices[name] = true
		if _, ok := b.Services[name]; !ok {
			return fmt.Errorf("manifest lists service %s but service-schemas/%s.json is missing from bundle", name, name)
		}
	}

	for file, doc := range b.Prompts {
		if err := promptSchema.Validate(doc); err != nil {
			return fmt.Errorf("prompt %s: %w", file, err)
		}
		for _, s := range doc["services"].([]any) {
			sm := s.(map[string]any)
			name := sm["name"].(string)
			if !manifestServices[name] {
				return fmt.Errorf("prompt %s declares service %s not in bundle manifest", file, name)
			}
		}
	}

	return nil
}
