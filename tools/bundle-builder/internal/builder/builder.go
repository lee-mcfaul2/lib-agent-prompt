// Package builder assembles the bundle manifest from a loaded service catalog.
package builder

import (
	"fmt"
	"sort"
	"time"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/loader"
)

// Manifest is the JSON shape written to bundle-manifest.json at the
// bundle root. The schema lives at schemas/bundle-manifest.json.
type Manifest struct {
	BundleVersion        string            `json:"bundle_version"`
	SchemaLibraryVersion string            `json:"schema_library_version"`
	Build                BuildProvenance   `json:"build"`
	EnvelopeCostCaps     EnvelopeCostCaps  `json:"envelope_cost_caps"`
	Services             []ManifestService `json:"services"`
}

type BuildProvenance struct {
	Timestamp    time.Time `json:"timestamp"`
	SourceCommit string    `json:"source_commit"`
	BuilderID    string    `json:"builder_id"`
}

type EnvelopeCostCaps struct {
	MaxIterations  int     `json:"max_iterations"`
	MaxWallclockMs int     `json:"max_wallclock_ms"`
	MaxCostUSD     float64 `json:"max_cost_usd"`
}

type ManifestService struct {
	MCP   string         `json:"mcp"`
	Tools []ManifestTool `json:"tools"`
}

type ManifestTool struct {
	Name                string   `json:"name"`
	RequestDigest       string   `json:"request_digest"`
	ResponseDigest      string   `json:"response_digest"`
	Write               bool     `json:"write,omitempty"`
	RequiresPermissions []string `json:"requires_permissions,omitempty"`
}

// BuildInput is everything the builder needs to assemble a manifest.
type BuildInput struct {
	BundleVersion        string
	SchemaLibraryVersion string
	Build                BuildProvenance
	EnvelopeCostCaps     EnvelopeCostCaps
	Services             []loader.Service
}

// BuildManifest produces a Manifest from the loaded service catalog. The
// returned manifest has services alphabetized by MCP and each service's
// tools alphabetized by name (defense-in-depth — the loader already sorts,
// but unit-test callers may pass unsorted input).
func BuildManifest(in BuildInput) (*Manifest, error) {
	if len(in.Services) == 0 {
		return nil, fmt.Errorf("at least one service required")
	}
	if in.BundleVersion == "" || in.SchemaLibraryVersion == "" {
		return nil, fmt.Errorf("bundle_version and schema_library_version required")
	}

	out := &Manifest{
		BundleVersion:        in.BundleVersion,
		SchemaLibraryVersion: in.SchemaLibraryVersion,
		Build:                in.Build,
		EnvelopeCostCaps:     in.EnvelopeCostCaps,
	}

	// Copy services, sorting tools within each.
	services := make([]loader.Service, len(in.Services))
	copy(services, in.Services)
	sort.Slice(services, func(i, j int) bool { return services[i].MCP < services[j].MCP })

	for _, s := range services {
		tools := make([]loader.Tool, len(s.Tools))
		copy(tools, s.Tools)
		sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })

		ms := ManifestService{MCP: s.MCP}
		for _, t := range tools {
			mt := ManifestTool{
				Name:                t.Name,
				RequestDigest:       t.RequestDigest,
				ResponseDigest:      t.ResponseDigest,
				Write:               t.Write,
				RequiresPermissions: t.RequiresPermissions,
			}
			ms.Tools = append(ms.Tools, mt)
		}
		out.Services = append(out.Services, ms)
	}

	return out, nil
}
