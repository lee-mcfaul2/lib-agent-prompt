// Package loader walks a lib-agent-prompt source tree and produces a
// canonical list of services with per-tool SHA-256 digests. Inputs are
// schemas/services/<mcp>/<tool>.{request,response,meta}.json files.
package loader

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/jcs"
)

// Service is one MCP's catalog entry in the bundle manifest.
type Service struct {
	MCP   string
	Tools []Tool
}

// Tool is one (request, response) pair plus its optional meta.
type Tool struct {
	Name                string
	RequestDigest       string // "sha256:..."
	ResponseDigest      string
	Write               bool
	RequiresPermissions []string
}

type meta struct {
	RequiresPermissions []string `json:"requires_permissions"`
	Write               bool     `json:"write"`
}

// LoadServices walks <root>/services/<mcp>/<tool>.{request,response,meta}.json
// and returns one Service per MCP, with tools sorted alphabetically by name
// and services sorted alphabetically by MCP. Each request/response file is
// canonicalized (JCS) before digesting so trailing whitespace doesn't drift
// the digest across machines.
func LoadServices(root string) ([]Service, error) {
	servicesRoot := filepath.Join(root, "services")
	entries, err := os.ReadDir(servicesRoot)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", servicesRoot, err)
	}

	var services []Service
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mcp := e.Name()
		s, err := loadOneMCP(servicesRoot, mcp)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no services found under %s", servicesRoot)
	}

	sort.Slice(services, func(i, j int) bool { return services[i].MCP < services[j].MCP })
	return services, nil
}

func loadOneMCP(servicesRoot, mcp string) (Service, error) {
	mcpDir := filepath.Join(servicesRoot, mcp)
	entries, err := os.ReadDir(mcpDir)
	if err != nil {
		return Service{}, fmt.Errorf("read %s: %w", mcpDir, err)
	}

	// Group files by tool name.
	byTool := map[string]map[string]string{} // tool → kind → path
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
			return Service{}, fmt.Errorf("malformed filename %s/%s: expected <tool>.<kind>.json", mcp, name)
		}
		tool, kind := parts[0], parts[1]
		if byTool[tool] == nil {
			byTool[tool] = map[string]string{}
		}
		byTool[tool][kind] = filepath.Join(mcpDir, name)
	}

	if len(byTool) == 0 {
		return Service{}, fmt.Errorf("%s: no tools", mcp)
	}

	// Build each tool.
	var tools []Tool
	for name, files := range byTool {
		t, err := loadOneTool(mcp, name, files)
		if err != nil {
			return Service{}, err
		}
		tools = append(tools, t)
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })

	return Service{MCP: mcp, Tools: tools}, nil
}

func loadOneTool(mcp, name string, files map[string]string) (Tool, error) {
	reqPath, ok := files["request"]
	if !ok {
		return Tool{}, fmt.Errorf("%s.%s: missing request schema", mcp, name)
	}
	respPath, ok := files["response"]
	if !ok {
		return Tool{}, fmt.Errorf("%s.%s: missing response schema", mcp, name)
	}

	reqDigest, err := digestFile(reqPath)
	if err != nil {
		return Tool{}, err
	}
	respDigest, err := digestFile(respPath)
	if err != nil {
		return Tool{}, err
	}

	t := Tool{
		Name:           name,
		RequestDigest:  reqDigest,
		ResponseDigest: respDigest,
	}

	if metaPath, ok := files["meta"]; ok {
		raw, err := os.ReadFile(metaPath)
		if err != nil {
			return Tool{}, fmt.Errorf("read %s: %w", metaPath, err)
		}
		var m meta
		if err := json.Unmarshal(raw, &m); err != nil {
			return Tool{}, fmt.Errorf("%s: %w", metaPath, err)
		}
		t.Write = m.Write
		t.RequiresPermissions = m.RequiresPermissions
	}

	return t, nil
}

// digestFile reads a JSON file, re-encodes it via JCS for canonical form,
// and returns a "sha256:<hex>" digest of the canonical bytes.
func digestFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	// Unmarshal into a generic value so we can re-encode canonically.
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}
	canon, err := jcs.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("canonicalize %s: %w", path, err)
	}
	sum := sha256.Sum256(canon)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
