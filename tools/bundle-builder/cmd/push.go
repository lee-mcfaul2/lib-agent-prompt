package cmd

import (
	"context"
	"fmt"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/push"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [bundle.tar.gz] [registry/repo:tag]",
	Short: "Push a packed bundle to an OCI registry via oras",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		digest, err := push.Push(context.Background(), args[0], args[1])
		if err != nil {
			return err
		}
		fmt.Printf("pushed: %s\ndigest: %s\n", args[1], digest)
		return nil
	},
}
