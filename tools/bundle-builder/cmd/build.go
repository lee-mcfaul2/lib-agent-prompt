package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/builder"
	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/loader"
	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/validator"
	"github.com/spf13/cobra"
)

var (
	buildSourceDir           string
	buildOutDir              string
	buildVersion             string
	buildSchemaLibVersion    string
	buildSourceCommit        string
	buildBuilderID           string
	buildMaxIterations       int
	buildMaxWallclockMs      int
	buildMaxCostUSD          float64
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Assemble a bundle directory tree from the source repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		// The loader/validator treat their root as the dir that contains services/.
		// The canonical layout is <source-dir>/schemas/services/…, so we pass
		// <source-dir>/schemas as the schema root for those two packages.
		schemaRoot := filepath.Join(buildSourceDir, "schemas")

		// 1. Validate schema tree
		if err := validator.ValidateSchemaTree(schemaRoot); err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}

		// 2. Load services
		services, err := loader.LoadServices(schemaRoot)
		if err != nil {
			return fmt.Errorf("loading services: %w", err)
		}

		// 3. Resolve source commit
		commit := buildSourceCommit
		if commit == "" {
			commit = gitCommit()
		}

		// 4. Resolve schema-lib version
		schemaLibVersion := buildSchemaLibVersion
		if schemaLibVersion == "" {
			schemaLibVersion = buildVersion
		}

		// 5. Build manifest
		m, err := builder.BuildManifest(builder.BuildInput{
			BundleVersion:        buildVersion,
			SchemaLibraryVersion: schemaLibVersion,
			Build: builder.BuildProvenance{
				Timestamp:    time.Now().UTC(),
				SourceCommit: commit,
				BuilderID:    buildBuilderID,
			},
			EnvelopeCostCaps: builder.EnvelopeCostCaps{
				MaxIterations:  buildMaxIterations,
				MaxWallclockMs: buildMaxWallclockMs,
				MaxCostUSD:     buildMaxCostUSD,
			},
			Services: services,
		})
		if err != nil {
			return fmt.Errorf("building manifest: %w", err)
		}

		// 6. Write bundle output
		if err := writeBundle(buildOutDir, m, buildSourceDir); err != nil {
			return fmt.Errorf("writing bundle: %w", err)
		}

		fmt.Printf("bundle written to %s (%d service(s))\n", buildOutDir, len(m.Services))
		return nil
	},
}

func gitCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short=12", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func writeBundle(outDir string, m *builder.Manifest, sourceDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	manifestBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "bundle-manifest.json"), manifestBytes, 0o644); err != nil {
		return err
	}

	srcSchemas := filepath.Join(sourceDir, "schemas")
	dstSchemas := filepath.Join(outDir, "schemas")
	return copyDir(srcSchemas, dstSchemas)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	})
}

func init() {
	buildCmd.Flags().StringVar(&buildSourceDir, "source-dir", ".", "Root of the lib-agent-prompt source tree")
	buildCmd.Flags().StringVar(&buildOutDir, "out-dir", "./out", "Output directory for the assembled bundle")
	buildCmd.Flags().StringVar(&buildVersion, "bundle-version", "", "Bundle version (semver)")
	buildCmd.Flags().StringVar(&buildSchemaLibVersion, "schema-library-version", "", "Schema-library version (defaults to bundle-version)")
	buildCmd.Flags().StringVar(&buildSourceCommit, "source-commit", "", "Git commit SHA (defaults to git rev-parse HEAD)")
	buildCmd.Flags().StringVar(&buildBuilderID, "builder-id", "local-bundle-builder", "Builder identifier recorded in provenance")
	buildCmd.Flags().IntVar(&buildMaxIterations, "max-iterations", 8, "Maximum agent loop iterations")
	buildCmd.Flags().IntVar(&buildMaxWallclockMs, "max-wallclock-ms", 300000, "Maximum wall-clock time in milliseconds")
	buildCmd.Flags().Float64Var(&buildMaxCostUSD, "max-cost-usd", 1.0, "Maximum cost cap in USD")
	_ = buildCmd.MarkFlagRequired("bundle-version")
}
