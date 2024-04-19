package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vahid-haghighat/awsure/cmd/internal"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
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

func init() {
	rootCmd.AddCommand(configCmd)
}
