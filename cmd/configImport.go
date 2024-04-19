package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vahid-haghighat/awsure/cmd/internal"
)

var importPath string

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports configurations from config file",
	Long:  `Imports configurations from config file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.ConfigImport(importPath)
	},
}

func init() {
	configCmd.AddCommand(importCmd)
}
