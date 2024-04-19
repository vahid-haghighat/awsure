package cmd

import (
	"awsure/cmd/internal"
	"awsure/cmd/types"
	"os"

	"github.com/spf13/cobra"
)

var configuration types.Configuration

var rootCmd = &cobra.Command{
	Use:   "awsure",
	Short: "Helps setting aws cli credentials with azure login",
	Long:  `Helps setting aws cli credentials with azure login`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.Login(configuration)
	},
	Args: func(cmd *cobra.Command, args []string) error {
		return validateProfileAndAllProfileInput(cmd)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configuration.Profile, "profile", "p", "default", "The name of the profile to log in with or configure (This is mutually exclusive with --all-profiles)")
	rootCmd.PersistentFlags().BoolVarP(&configuration.AllProfiles, "all-profiles", "a", false, "Log in or configure all profiles (This is mutually exclusive with --profile)")
}
