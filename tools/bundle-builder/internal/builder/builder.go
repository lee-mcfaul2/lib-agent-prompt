package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/jcs"
	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/loader"
)

type Options struct {
	SchemaLib        string
	Prompts          string
	Services         string
	Output           string
	Version          string
	SchemaLibVersion string
	BuildTime        time.Time
	SourceCommit    string
	BuilderID       string
	// Reserved for Task 10:
	RegistryBase     string
	AllowPlaceholder bool
}

type prompt struct {
	ID   string `json:"id"`
	File string `json:"file"`
	Path string `json:"-"`
}

type service struct {
	Name         string `json:"name"`
	Spiffe       string `json:"spiffe"`
	SourceDigest string `json:"source_digest"`
	EmbeddedFile string `json:"embedded_file"`
}

// Build assembles the bundle directory tree.
func Build(ctx context.Context, opts Options) error {
	if err := os.MkdirAll(opts.Output, 0o755); err != nil {
		return err
	}

	if err := copyDir(opts.SchemaLib, filepath.Join(opts.Output, "schemas")); err != nil {
		return fmt.Errorf("snapshot schemas: %w", err)
	}

	prompts, err := copyPrompts(opts.Prompts, filepath.Join(opts.Output, "prompts"))
	if err != nil {
		return fmt.Errorf("copy prompts: %w", err)
	}

	services, err := embedServices(ctx, opts.Services, filepath.Join(opts.Output, "service-schemas"), opts.RegistryBase, opts.AllowPlaceholder)
	if err != nil {
		return fmt.Errorf("embed services: %w", err)
	}

	return writeManifest(opts, prompts, services)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyPrompts(src, dst string) ([]prompt, error) {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return nil, err
	}
	docs, err := loader.LoadAllJSON(src)
	if err != nil {
		return nil, err
	}
	var out []prompt
	for rel, doc := range docs {
		if strings.HasSuffix(rel, "README.md") || strings.HasPrefix(rel, ".") {
			continue
		}
		id, _ := doc["id"].(string)
		if id == "" {
			return nil, fmt.Errorf("prompt %s missing id", rel)
		}
		fname := id + ".json"
		if err := copyFile(filepath.Join(src, rel), filepath.Join(dst, fname)); err != nil {
			return nil, err
		}
		out = append(out, prompt{
			ID:   id,
			File: filepath.Join("prompts", fname),
			Path: filepath.Join(dst, fname),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func embedServices(ctx context.Context, src, dst, registryBase string, allowPlaceholder bool) ([]service, error) {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return nil, err
	}
	docs, err := loader.LoadAllJSON(src)
	if err != nil {
		return nil, err
	}
	const placeholder = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	var out []service
	for rel, doc := range docs {
		name, _ := doc["name"].(string)
		if name == "" {
			return nil, fmt.Errorf("service-reference %s missing name", rel)
		}
		spiffe, _ := doc["spiffe"].(string)
		digest, _ := doc["source_digest"].(string)

		var payload []byte
		if digest == placeholder {
			if !allowPlaceholder {
				return nil, fmt.Errorf("service %s has placeholder digest; pass --allow-placeholder to permit", name)
			}
			payload = []byte("{}")
		} else {
			ref := registryBase + "/" + name
			payload, err = PullServiceSchema(ctx, ref, digest)
			if err != nil {
				return nil, fmt.Errorf("pull %s: %w", name, err)
			}
		}

		embeddedRel := filepath.Join("service-schemas", name+".json")
		if err := os.WriteFile(filepath.Join(dst, name+".json"), payload, 0o644); err != nil {
			return nil, err
		}
		out = append(out, service{
			Name:         name,
			Spiffe:       spiffe,
			SourceDigest: digest,
			EmbeddedFile: embeddedRel,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func writeManifest(opts Options, prompts []prompt, services []service) error {
	type promptEntry struct {
		ID     string `json:"id"`
		File   string `json:"file"`
		Digest string `json:"digest"`
	}
	var promptEntries []promptEntry
	for _, p := range prompts {
		data, err := os.ReadFile(p.Path)
		if err != nil {
			return err
		}
		var doc any
		if err := json.Unmarshal(data, &doc); err != nil {
			return err
		}
		d, err := jcs.Sha256OfJCS(doc)
		if err != nil {
			return err
		}
		promptEntries = append(promptEntries, promptEntry{ID: p.ID, File: p.File, Digest: d})
	}

	if opts.SchemaLibVersion == "" {
		opts.SchemaLibVersion = opts.Version
	}

	manifest := map[string]any{
		"bundle_version":         opts.Version,
		"schema_library_version": opts.SchemaLibVersion,
		"build": map[string]any{
			"timestamp":     opts.BuildTime.UTC().Format(time.RFC3339),
			"source_commit": opts.SourceCommit,
			"builder_id":    opts.BuilderID,
		},
		"envelope_cost_caps": map[string]any{
			"max_iterations":   50,
			"max_wallclock_ms": 600000,
			"max_cost_usd":     10.0,
		},
		"prompts":  promptEntries,
		"services": services,
	}

	b, err := jcs.Marshal(manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(opts.Output, "bundle-manifest.json"), b, 0o644)
}
