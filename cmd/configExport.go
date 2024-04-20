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
		var err error
		exportPath, err = absolutePath(exportPath)
		if err != nil {
			return err
		}

		return internal.ConfigExport(exportPath)
	},
}

func init() {
	importCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		hideFlagsExcept(rootCmd)
		command.Parent().HelpFunc()(command, strings)
	})
	exportCmd.Flags().StringVarP(&exportPath, "file", "f", "", "The path to export the config file to")
	_ = exportCmd.MarkFlagRequired("file")
	configCmd.AddCommand(exportCmd)
}
