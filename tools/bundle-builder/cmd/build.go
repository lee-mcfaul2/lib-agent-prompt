package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	buildSchemaLib string
	buildPrompts   string
	buildServices  string
	buildOutput    string
	buildVersion   string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Assemble a bundle directory tree from the source repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	buildCmd.Flags().StringVar(&buildSchemaLib, "schema-lib", "./schemas", "Path to the canonical schemas/ directory")
	buildCmd.Flags().StringVar(&buildPrompts, "prompts", "./prompts", "Path to the prompts/ directory")
	buildCmd.Flags().StringVar(&buildServices, "services", "./schemas/service-references", "Path to the service-references directory")
	buildCmd.Flags().StringVar(&buildOutput, "output", "./out/bundle", "Output directory for the assembled bundle tree")
	buildCmd.Flags().StringVar(&buildVersion, "version", "", "Bundle version (semver)")
	_ = buildCmd.MarkFlagRequired("version")
}
