package cmd

import (
	"fmt"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/loader"
	"github.com/lee-mcfaul2/lib-agent-prompt/tools/bundle-builder/internal/validator"
	"github.com/spf13/cobra"
)

var validateSchemaLib string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the schema library is internally consistent and self-validating",
	RunE: func(cmd *cobra.Command, args []string) error {
		docs, err := loader.LoadAllJSON(validateSchemaLib)
		if err != nil {
			return err
		}
		if err := validator.ValidateSchemaSelfConsistent(validateSchemaLib, docs); err != nil {
			return err
		}
		fmt.Printf("OK: %d schemas validated\n", len(docs))
		return nil
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateSchemaLib, "schema-lib", "./schemas", "Path to schemas/")
}
