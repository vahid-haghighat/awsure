package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func validateProfileAndAllProfileInput(cmd *cobra.Command) error {
	if cmd.Flag("profile").Changed && cmd.Flag("all-profiles").Changed {
		return fmt.Errorf("you cannot specify both --profile and --all-profiles")
	}
	return nil
}
