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
		if cmd.Flag("all-profiles").Changed {
			_, err := cmd.Flags().GetBool("all-profiles")
			if err != nil {
				return err
			}
			return internal.ConfigAll()
		}

		profile, err := cmd.Flags().GetString("profile")
		if err != nil {
			return err
		}

		return internal.ConfigProfile(profile)
	},
	Args: func(cmd *cobra.Command, args []string) error {
		return validateProfileAndAllProfileInput(cmd)
	},
}
var configCmd = configureCmd

func init() {
	configCmd.Use = "config"
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(configCmd)
}
