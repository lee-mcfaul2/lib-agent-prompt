package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "agent-prompt-verify",
	Short: "Reference verifier for lib-agent-prompt bundles",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
