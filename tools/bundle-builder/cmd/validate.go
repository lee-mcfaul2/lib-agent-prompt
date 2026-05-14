package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/validator"
	"github.com/spf13/cobra"
)

var validateSourceDir string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the schema tree is internally consistent",
	RunE: func(cmd *cobra.Command, args []string) error {
		// The validator expects root/services/; canonical layout puts services
		// under <source-dir>/schemas/services/, so pass schemas/ as the root.
		schemaRoot := filepath.Join(validateSourceDir, "schemas")
		if err := validator.ValidateSchemaTree(schemaRoot); err != nil {
			return err
		}
		fmt.Println("OK: schema tree validated")
		return nil
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateSourceDir, "source-dir", ".", "Root of the lib-agent-prompt source tree")
}
