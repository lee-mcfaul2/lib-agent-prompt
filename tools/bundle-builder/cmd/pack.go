package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/pack"
	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use:   "pack [bundle-dir]",
	Short: "Pack an assembled bundle directory into a reproducible tar.gz",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		dst := src + ".tar.gz"
		digest, err := pack.PackReproducible(src, dst, pack.ReadSourceDateEpoch())
		if err != nil {
			return err
		}
		fmt.Printf("packed: %s\ndigest: %s\n", filepath.Base(dst), digest)
		return nil
	},
}
