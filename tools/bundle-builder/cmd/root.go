package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "bundle-builder",
	Short: "Assemble, pack, validate, and push lib-agent-prompt bundles",
	Long:  "bundle-builder is the build tool for lib-agent-prompt bundles. It validates schemas + prompts, embeds service-schema artifacts, produces a reproducible tarball, and pushes it as an OCI artifact.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(buildCmd, packCmd, validateCmd, pushCmd)
}
