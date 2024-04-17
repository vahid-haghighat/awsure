package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configures the cli",
	Long:  `Configures the cli`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(configuration.Profile)
	},
}
var configCmd = configureCmd

func init() {
	configCmd.Use = "config"
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(configCmd)
}
