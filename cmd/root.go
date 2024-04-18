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
	rootCmd.Flags().BoolVarP(&configuration.ForceRefresh, "force-refresh", "f", false, "Force refresh of the access tokens")
	rootCmd.Flags().StringVarP(&configuration.Mode, "mode", "m", "cli", "'cli' to hide the login page and perform the login through the CLI, 'gui' to perform the login through the Azure GUI (more reliable but only works on GUI operating system and also the default behaviour)")
	rootCmd.Flags().BoolVarP(&configuration.NoVerifySSL, "no-verify-ssl", "s", false, "Disable SSL certificate verification")
	rootCmd.Flags().BoolVarP(&configuration.NoPrompt, "no-prompt", "t", false, "Disable prompting and use default configuration values")
}
