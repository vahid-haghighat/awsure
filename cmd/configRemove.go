package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vahid-haghighat/awsure/cmd/internal"
)

// deleteCmd represents the delete command
var configRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes a configuration profile",
	Long:  `Removes a configuration profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := cmd.Flags().GetString("profile")
		if err != nil {
			return err
		}

		return internal.ConfigRemove(profile)
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("profile") {
			return fmt.Errorf("the profile name to remove should explicitly be specified")
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configRemoveCmd)
}
