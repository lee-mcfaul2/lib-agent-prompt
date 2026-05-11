// Top-level manifest inside every bundle. Pins bundle version, schema-library version,
// prompts (by UUID + digest), services (by SPIFFE + source digest + embedded file).
type BundleManifest struct {
	Build                Build            `json:"build"`
	BundleVersion        string           `json:"bundle_version"`
	EnvelopeCostCaps     EnvelopeCostCaps `json:"envelope_cost_caps"`
	Prompts              []Prompt         `json:"prompts"`
	SchemaLibraryVersion string           `json:"schema_library_version"`
	Services             []Service        `json:"services"`
}

import "time"

type Build struct {
	BuilderID    string    `json:"builder_id"`
	SourceCommit string    `json:"source_commit"`
	Timestamp    time.Time `json:"timestamp"`
}

type EnvelopeCostCaps struct {
	MaxCostUsd     float64 `json:"max_cost_usd"`
	MaxIterations  int64   `json:"max_iterations"`
	MaxWallclockMS int64   `json:"max_wallclock_ms"`
}

type Prompt struct {
	Digest string `json:"digest"`
	File   string `json:"file"`
	ID     string `json:"id"`
}

type Service struct {
	EmbeddedFile string `json:"embedded_file"`
	Name         string `json:"name"`
	SourceDigest string `json:"source_digest"`
	Spiffe       string `json:"spiffe"`
}
