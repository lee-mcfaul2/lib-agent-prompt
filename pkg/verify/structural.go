package verify

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	v5 "github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/santhosh-tekuri/jsonschema/v6"
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
	c := v5.NewCompiler()
	c.Draft = v5.Draft2020

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
			return v5.LoadURL("file://" + filepath.Join(tmp, "schemas", rel))
		}
		if strings.HasPrefix(s, "file://") {
			return v5.LoadURL(s)
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

// VerifyBundleStructure performs cheap shape checks on a bundle root.
// The bundle is expected to have:
//   - schemas/user-prompt.json, final-response.json, tool-result.json, bundle-manifest.json
//   - schemas/shared/uuid.json (referenced by the envelopes)
//   - schemas/services/<mcp>/<tool>.{request,response}.json pairs
//   - bundle-manifest.json at the root (after build) — optional for source trees
func VerifyBundleStructure(root string) error {
	requiredFiles := []string{
		"schemas/user-prompt.json",
		"schemas/final-response.json",
		"schemas/tool-result.json",
		"schemas/bundle-manifest.json",
		"schemas/shared/uuid.json",
	}
	for _, p := range requiredFiles {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			return fmt.Errorf("missing required file %s: %w", p, err)
		}
	}

	servicesRoot := filepath.Join(root, "schemas", "services")
	if _, err := os.Stat(servicesRoot); err != nil {
		return fmt.Errorf("missing schemas/services/: %w", err)
	}

	mcps, err := os.ReadDir(servicesRoot)
	if err != nil {
		return err
	}
	if len(mcps) == 0 {
		return fmt.Errorf("schemas/services/ has no MCPs")
	}
	for _, mcp := range mcps {
		if !mcp.IsDir() {
			continue
		}
		if err := verifyMCPDir(filepath.Join(servicesRoot, mcp.Name()), mcp.Name()); err != nil {
			return err
		}
	}
	return nil
}

func verifyMCPDir(dir, mcp string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	tools := map[string]map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		base := strings.TrimSuffix(name, ".json")
		// Expected format: <tool>.<kind>.json where kind is request, response, or meta.
		// We only require request + response pairs.
		parts := strings.SplitN(base, ".", 2)
		if len(parts) != 2 {
			return fmt.Errorf("%s/%s: bad filename", mcp, name)
		}
		tool, kind := parts[0], parts[1]
		if kind == "meta" {
			continue
		}
		if tools[tool] == nil {
			tools[tool] = map[string]bool{}
		}
		tools[tool][kind] = true
	}
	for tool, kinds := range tools {
		if !kinds["request"] || !kinds["response"] {
			return fmt.Errorf("%s.%s: incomplete (need request + response)", mcp, tool)
		}
	}
	return nil
}

// VerifyUserPrompt validates a user-prompt instance against the bundle's
// user-prompt.json schema.
func VerifyUserPrompt(root string, instance []byte) error {
	return validateInstance(filepath.Join(root, "schemas", "user-prompt.json"), instance)
}

// VerifyToolRequest validates a tool-call args payload against the bundle's
// services/<mcp>/<tool>.request.json schema.
func VerifyToolRequest(root, mcp, tool string, args []byte) error {
	path := filepath.Join(root, "schemas", "services", mcp, tool+".request.json")
	return validateInstance(path, args)
}

// VerifyToolResponse validates a tool-call response payload against the
// bundle's services/<mcp>/<tool>.response.json schema.
func VerifyToolResponse(root, mcp, tool string, response []byte) error {
	path := filepath.Join(root, "schemas", "services", mcp, tool+".response.json")
	return validateInstance(path, response)
}

func validateInstance(schemaPath string, instance []byte) error {
	c := jsonschema.NewCompiler()

	// Determine the schemas/ root so we can pre-load all shared schemas.
	// schemaPath is either:
	//   <root>/schemas/<name>.json           → schemasRoot = filepath.Dir(schemaPath)
	//   <root>/schemas/services/<mcp>/<tool>.<kind>.json → schemasRoot = four dirs up from file
	dir := filepath.Dir(schemaPath)
	var schemasRoot string
	switch filepath.Base(dir) {
	case "schemas":
		schemasRoot = dir
	default:
		// Walk up until we find the "schemas" directory.
		schemasRoot = dir
		for filepath.Base(schemasRoot) != "schemas" && schemasRoot != "/" {
			schemasRoot = filepath.Dir(schemasRoot)
		}
	}

	sharedDir := filepath.Join(schemasRoot, "shared")
	if entries, err := os.ReadDir(sharedDir); err == nil {
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			full := filepath.Join(sharedDir, e.Name())
			raw, err := os.ReadFile(full)
			if err != nil {
				continue
			}
			doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
			if err != nil {
				continue
			}
			// Register under the $id URI used in the schemas (e.g. "https://ai-security.io/schemas/shared/uuid.json")
			// and also under the relative path used in $ref values (e.g. "shared/uuid.json").
			_ = c.AddResource("shared/"+e.Name(), doc)

			// Also parse the $id from the schema and register under that URI so the
			// compiler can resolve absolute $ref URIs.
			var schemaDoc map[string]any
			if json.Unmarshal(raw, &schemaDoc) == nil {
				if id, ok := schemaDoc["$id"].(string); ok && id != "" {
					rawDoc, _ := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
					_ = c.AddResource(id, rawDoc)
				}
			}
		}
	}

	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema %s: %w", schemaPath, err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("parse schema %s: %w", schemaPath, err)
	}
	if err := c.AddResource(schemaPath, doc); err != nil {
		return fmt.Errorf("add schema %s: %w", schemaPath, err)
	}
	s, err := c.Compile(schemaPath)
	if err != nil {
		return fmt.Errorf("compile schema %s: %w", schemaPath, err)
	}
	var inst any
	if err := json.Unmarshal(instance, &inst); err != nil {
		return fmt.Errorf("parse instance: %w", err)
	}
	return s.Validate(inst)
}
