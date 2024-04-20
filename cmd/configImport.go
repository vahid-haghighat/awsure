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
		var err error
		importPath, err = absolutePath(importPath)
		if err != nil {
			return err
		}

		return internal.ConfigImport(importPath)
	},
}

func init() {
	importCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		hideFlagsExcept(rootCmd)
		command.Parent().HelpFunc()(command, strings)
	})
	importCmd.Flags().StringVarP(&importPath, "file", "f", "", "Path to the config file")
	_ = importCmd.MarkFlagRequired("file")
	configCmd.AddCommand(importCmd)
}
