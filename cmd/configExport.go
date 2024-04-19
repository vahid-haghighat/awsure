package cmd

import (
	"github.com/vahid-haghighat/awsure/cmd/internal"

	"github.com/spf13/cobra"
)

var exportPath string

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports config file",
	Long:  `Exports config file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.ConfigExport(exportPath)
	},
}

func init() {
	configCmd.AddCommand(exportCmd)
}
