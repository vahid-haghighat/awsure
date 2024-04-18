package internal

import (
	"os"
	"path/filepath"
)

var homeDir string
var defaultConfigLocation string

func init() {
	var err error
	homeDir, err = os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	defaultConfigLocation = filepath.Join(homeDir, ".config", "awsure", "config.yml")
}
