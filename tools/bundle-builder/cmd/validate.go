package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var validateSchemaLib string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the schema library is internally consistent and self-validating",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateSchemaLib, "schema-lib", "./schemas", "Path to schemas/")
}
