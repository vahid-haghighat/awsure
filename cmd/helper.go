package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
	"strings"
)

func contains[T comparable](b []T, i T) bool {
	for _, s := range b {
		if s == i {
			return true
		}
	}
	return false
}

func hideFlagsExcept(command *cobra.Command, unhidden ...string) {
	command.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		name := flag.Name
		if !contains(unhidden, name) {
			flag.Hidden = true
		}
	})
}

func absolutePath(targetPath string) (string, error) {
	if strings.Contains(targetPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		targetPath = strings.ReplaceAll(targetPath, "~", home)
	}
	return filepath.Abs(targetPath)
}
