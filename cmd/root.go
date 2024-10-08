package cmd

import (
	"fmt"
	"github.com/vahid-haghighat/awsure/cmd/internal"
	"github.com/vahid-haghighat/awsure/cmd/types"
	"github.com/vahid-haghighat/awsure/version"
	"os"

	"github.com/spf13/cobra"
)

var configuration types.Configuration
var versionFlag bool

var rootCmd = &cobra.Command{
	Use:   "awsure",
	Short: "Helps setting aws cli credentials with azure login",
	Long:  `Helps setting aws cli credentials with azure login`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionFlag {
			fmt.Println(version.Version)
			return nil
		}

		if cmd.Flags().Changed("profile") {
			return internal.Login(configuration.Profile, nil)
		}

		return internal.LoginAll()
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
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print the version and exit")
}
