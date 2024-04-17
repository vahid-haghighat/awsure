package cmd

import (
	"awsure/cmd/internal"
	"github.com/spf13/cobra"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configures the cli",
	Long:  `Configures the cli`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.Configure(configuration)
	},
}
var configCmd = configureCmd

func init() {
	configCmd.Use = "config"
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(configCmd)
}
