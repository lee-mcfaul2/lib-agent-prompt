package validator

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Compiler returns a configured jsonschema.Compiler that resolves $refs relative to a local schemas/ root.
func Compiler(schemaRoot string) (*jsonschema.Compiler, error) {
	absRoot, err := filepath.Abs(schemaRoot)
	if err != nil {
		return nil, fmt.Errorf("abs schema-root: %w", err)
	}
	c := jsonschema.NewCompiler()
	c.Draft = jsonschema.Draft2020

	// Map https://ai-security.io/schemas/<path> → file://<absRoot>/<path>.
	// file:// URLs pass through to the default loader. http(s):// for other hosts is refused.
	c.LoadURL = func(s string) (io.ReadCloser, error) {
		const prefix = "https://ai-security.io/schemas/"
		if strings.HasPrefix(s, prefix) {
			rel := strings.TrimPrefix(s, prefix)
			local := filepath.Join(absRoot, rel)
			return jsonschema.LoadURL("file://" + local)
		}
		if strings.HasPrefix(s, "file://") {
			return jsonschema.LoadURL(s)
		}
		return nil, fmt.Errorf("refusing external URL: %s", s)
	}
	return c, nil
}

// ValidatePrompt validates one prompt JSON document against schemas/prompt.json.
func ValidatePrompt(schemaRoot string, promptDoc any) error {
	c, err := Compiler(schemaRoot)
	if err != nil {
		return err
	}
	sch, err := c.Compile(filepath.Join(schemaRoot, "prompt.json"))
	if err != nil {
		return fmt.Errorf("compile prompt schema: %w", err)
	}
	return sch.Validate(promptDoc)
}

// ValidateSchemaSelfConsistent compiles every *.json in schemas/ to ensure they're well-formed JSON Schemas and resolve their own $refs.
func ValidateSchemaSelfConsistent(schemaRoot string, docs map[string]map[string]any) error {
	c, err := Compiler(schemaRoot)
	if err != nil {
		return err
	}
	for rel := range docs {
		path := filepath.Join(schemaRoot, rel)
		if _, err := c.Compile(path); err != nil {
			return fmt.Errorf("schema %s: %w", rel, err)
		}
	}
	return nil
}
