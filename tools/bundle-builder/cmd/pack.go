package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use:   "pack [bundle-dir]",
	Short: "Pack an assembled bundle directory into a reproducible tar.gz",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}
