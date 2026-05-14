// Package validator checks that a lib-agent-prompt source tree is internally
// consistent before the loader/builder run on it. Cheap sanity checks: every
// tool has a request + response pair, every JSON file parses, every schema
// declares draft-2020-12.
package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateSchemaTree walks root/services/<mcp>/<tool>.<kind>.json and ensures
// each tool has both a request and a response schema and every file parses.
func ValidateSchemaTree(root string) error {
	servicesRoot := filepath.Join(root, "services")
	entries, err := os.ReadDir(servicesRoot)
	if err != nil {
		return fmt.Errorf("read %s: %w", servicesRoot, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no MCPs under %s", servicesRoot)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mcp := e.Name()
		if err := validateOneMCP(filepath.Join(servicesRoot, mcp), mcp); err != nil {
			return err
		}
	}
	return nil
}

func validateOneMCP(mcpDir, mcp string) error {
	entries, err := os.ReadDir(mcpDir)
	if err != nil {
		return fmt.Errorf("read %s: %w", mcpDir, err)
	}

	byTool := map[string]map[string]bool{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		base := strings.TrimSuffix(name, ".json")
		parts := strings.SplitN(base, ".", 2)
		if len(parts) != 2 {
			return fmt.Errorf("%s/%s: expected <tool>.<kind>.json", mcp, name)
		}
		tool, kind := parts[0], parts[1]
		if kind != "request" && kind != "response" && kind != "meta" {
			return fmt.Errorf("%s/%s: unknown kind %q (expected request|response|meta)", mcp, name, kind)
		}
		if byTool[tool] == nil {
			byTool[tool] = map[string]bool{}
		}
		byTool[tool][kind] = true

		// Verify the file parses + (for request/response) declares draft-2020-12.
		raw, err := os.ReadFile(filepath.Join(mcpDir, name))
		if err != nil {
			return fmt.Errorf("read %s/%s: %w", mcp, name, err)
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return fmt.Errorf("parse %s/%s: %w", mcp, name, err)
		}
		if kind == "request" || kind == "response" {
			schemaURI, _ := doc["$schema"].(string)
			if schemaURI != "https://json-schema.org/draft/2020-12/schema" {
				return fmt.Errorf("%s/%s: $schema must be draft-2020-12 (got %q)", mcp, name, schemaURI)
			}
		}
	}

	if len(byTool) == 0 {
		return fmt.Errorf("%s: no tools", mcp)
	}

	for tool, kinds := range byTool {
		if !kinds["request"] {
			return fmt.Errorf("%s.%s: missing request schema", mcp, tool)
		}
		if !kinds["response"] {
			return fmt.Errorf("%s.%s: missing response schema", mcp, tool)
		}
	}
	return nil
}
