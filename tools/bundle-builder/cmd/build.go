package cmd

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/builder"
	"github.com/spf13/cobra"
)

var (
	buildSchemaLib        string
	buildPrompts          string
	buildServices         string
	buildOutput           string
	buildVersion          string
	buildSchemaLibVersion string
	buildAllowPlaceholder bool
	buildRegistry         string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Assemble a bundle directory tree from the source repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		commit := gitCommit()
		return builder.Build(context.Background(), builder.Options{
			SchemaLib:        buildSchemaLib,
			Prompts:          buildPrompts,
			Services:         buildServices,
			Output:           buildOutput,
			Version:          buildVersion,
			SchemaLibVersion: buildSchemaLibVersion,
			BuildTime:        time.Now().UTC(),
			SourceCommit:     commit,
			BuilderID:        "local-bundle-builder",
			RegistryBase:     buildRegistry,
			AllowPlaceholder: buildAllowPlaceholder,
		})
	},
}

func gitCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short=12", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func init() {
	buildCmd.Flags().StringVar(&buildSchemaLib, "schema-lib", "./schemas", "Path to the canonical schemas/ directory")
	buildCmd.Flags().StringVar(&buildPrompts, "prompts", "./prompts", "Path to the prompts/ directory")
	buildCmd.Flags().StringVar(&buildServices, "services", "./schemas/service-references", "Path to the service-references directory")
	buildCmd.Flags().StringVar(&buildOutput, "output", "./out/bundle", "Output directory for the assembled bundle tree")
	buildCmd.Flags().StringVar(&buildVersion, "version", "", "Bundle version (semver)")
	buildCmd.Flags().StringVar(&buildSchemaLibVersion, "schema-lib-version", "", "Schema-library version (defaults to bundle version)")
	buildCmd.Flags().BoolVar(&buildAllowPlaceholder, "allow-placeholder", false, "Permit all-zero source_digest placeholders (example bundle only)")
	buildCmd.Flags().StringVar(&buildRegistry, "registry", "ghcr.io/lee-mcfaul2/ai-security/services", "Base registry for mcp-schema artifacts")
	_ = buildCmd.MarkFlagRequired("version")
}
