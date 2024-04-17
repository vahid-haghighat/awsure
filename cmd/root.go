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
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configuration.Profile, "profile", "p", "default", "The name of the profile to log in with or configure")
	rootCmd.PersistentFlags().BoolVarP(&configuration.AllProfiles, "all-profiles", "a", false, "Log in or configure all profiles")
	rootCmd.Flags().BoolVarP(&configuration.ForceRefresh, "force-refresh", "f", false, "Force refresh of the access tokens")
	rootCmd.Flags().StringVarP(&configuration.Mode, "mode", "m", "cli", "'cli' to hide the login page and perform the login through the CLI (default behavior), 'gui' to perform the login through the Azure GUI (more reliable but only works on GUI operating system), 'debug' to show the login page but perform the login through the CLI (useful to debug issues with the CLI login)")
	rootCmd.Flags().BoolVarP(&configuration.NoVerifySSL, "no-verify-ssl", "s", false, "Disable SSL certificate verification")
	rootCmd.Flags().BoolVarP(&configuration.NoPrompt, "no-prompt", "t", false, "Disable prompting and use default configuration values")
}
