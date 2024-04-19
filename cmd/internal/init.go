package internal

import (
	"os"
	"path/filepath"
	"time"
)

var homeDir string
var defaultConfigLocation string
var defaultAwsCredentialsFileLocation string
var timeFormat string

func init() {
	var err error
	homeDir, err = os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	defaultConfigLocation = filepath.Join(homeDir, ".config", "awsure", "config.yml")
	defaultAwsCredentialsFileLocation = filepath.Join(homeDir, ".aws", "credentials")
	timeFormat = time.RFC3339
}
