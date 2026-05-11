package cmd

import (
	"fmt"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/verifier/internal/verify"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify [bundle.tar.gz]",
	Short: "Verify a packed bundle's structure + content hash",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, hash, err := verify.LoadAndHash(args[0])
		if err != nil {
			return err
		}
		if err := verify.VerifyStructure(b); err != nil {
			return err
		}
		fmt.Printf("OK: %s\n  digest: %s\n  prompts: %d\n  services: %d\n",
			args[0], hash, len(b.Prompts), len(b.Services))
		return nil
	},
}
