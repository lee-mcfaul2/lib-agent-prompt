package builder

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/loader"
)

func sampleServices() []loader.Service {
	return []loader.Service{
		{
			MCP: "kb",
			Tools: []loader.Tool{
				{
					Name:           "search",
					RequestDigest:  "sha256:" + hex64(),
					ResponseDigest: "sha256:" + hex64(),
					RequiresPermissions: []string{"kb:read"},
				},
				{
					Name:           "fetch",
					RequestDigest:  "sha256:" + hex64(),
					ResponseDigest: "sha256:" + hex64(),
				},
			},
		},
	}
}

func hex64() string {
	out := make([]byte, 64)
	for i := range out {
		out[i] = 'a'
	}
	return string(out)
}

func TestBuildManifestShape(t *testing.T) {
	m, err := BuildManifest(BuildInput{
		BundleVersion:        "1.0.0",
		SchemaLibraryVersion: "1.0.0",
		Build: BuildProvenance{
			Timestamp:     time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC),
			SourceCommit:  "abc1234",
			BuilderID:     "bundle-builder-test",
		},
		EnvelopeCostCaps: EnvelopeCostCaps{
			MaxIterations:  8,
			MaxWallclockMs: 300000,
			MaxCostUSD:     1.0,
		},
		Services: sampleServices(),
	})
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}

	raw, _ := json.MarshalIndent(m, "", "  ")
	t.Logf("manifest:\n%s", raw)

	if m.BundleVersion != "1.0.0" {
		t.Errorf("BundleVersion = %q", m.BundleVersion)
	}
	if len(m.Services) != 1 {
		t.Fatalf("services len = %d", len(m.Services))
	}
	if m.Services[0].MCP != "kb" {
		t.Errorf("MCP = %q", m.Services[0].MCP)
	}
	if len(m.Services[0].Tools) != 2 {
		t.Fatalf("tools len = %d", len(m.Services[0].Tools))
	}
	if m.Services[0].Tools[0].Name != "fetch" {
		t.Errorf("tools[0].Name = %q (expected alphabetized: fetch before search)", m.Services[0].Tools[0].Name)
	}
}

func TestBuildManifestWriteFlagDefaultsFalse(t *testing.T) {
	in := BuildInput{
		BundleVersion:        "1.0.0",
		SchemaLibraryVersion: "1.0.0",
		Build:                BuildProvenance{Timestamp: time.Now(), SourceCommit: "abc1234", BuilderID: "x"},
		EnvelopeCostCaps:     EnvelopeCostCaps{MaxIterations: 1, MaxWallclockMs: 1, MaxCostUSD: 0},
		Services: []loader.Service{
			{MCP: "kb", Tools: []loader.Tool{{Name: "search", RequestDigest: "sha256:" + hex64(), ResponseDigest: "sha256:" + hex64()}}},
		},
	}
	m, _ := BuildManifest(in)
	if m.Services[0].Tools[0].Write {
		t.Error("default Write should be false")
	}
	if m.Services[0].Tools[0].RequiresPermissions != nil {
		t.Errorf("default RequiresPermissions should be nil, got %v", m.Services[0].Tools[0].RequiresPermissions)
	}
}

func TestBuildManifestEmptyServicesError(t *testing.T) {
	_, err := BuildManifest(BuildInput{
		BundleVersion:        "1.0.0",
		SchemaLibraryVersion: "1.0.0",
		Build:                BuildProvenance{Timestamp: time.Now(), SourceCommit: "abc1234", BuilderID: "x"},
		EnvelopeCostCaps:     EnvelopeCostCaps{MaxIterations: 1, MaxWallclockMs: 1, MaxCostUSD: 0},
	})
	if err == nil {
		t.Fatal("expected error for empty services")
	}
}
