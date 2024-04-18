package internal

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
)

func Configure(profile string, allProfiles bool) error {
	if allProfiles {
		return configAll()
	}

	return configProfile(profile)
}

func configAll() error {
	configs, err := loadConfig()
	if err != nil {
		if errors.Is(err, fileNotFoundError) {
			return fmt.Errorf("cannot configure all profiles when no profile is configured")
		} else {
			return err
		}
	}

	prompter := Prompter{}

	fmt.Println("Leaving any of the following as empty will keep them unchanged on profiles")
	azureTenantId, err := prompter.Prompt("Azure Tenant Id", "")
	azureAppIdUri, err := prompter.Prompt("Azure App Id Uri", "")
	azureUsername, err := prompter.Prompt("Azure Username", "")
	oktaUsername, err := prompter.Prompt("Okta Username", "")
	defaultRoleArn, err := prompter.Prompt("Default Role Arn", "")
	defaultDurationHours, err := prompter.Prompt("Default Duration (Hour)", "")

	for profile, config := range configs {
		fmt.Printf("Updating %s profile\n", profile)

		if azureTenantId != "" {
			config.AzureTenantId = azureTenantId
		}

		if azureAppIdUri != "" {
			config.AzureAppIdUri = azureAppIdUri
		}

		if azureUsername != "" {
			config.AzureUsername = azureUsername
		}

		if oktaUsername != "" {
			config.OktaUsername = oktaUsername
		}

		if defaultRoleArn != "" {
			config.DefaultRoleArn = defaultRoleArn
		}

		if defaultDurationHours != "" {
			config.DefaultDurationHours, err = strconv.Atoi(defaultDurationHours)
			if err != nil {
				fmt.Println("Not a valid duration was entered. Will set the duration to 1.")
				config.DefaultDurationHours = 1
			} else if config.DefaultDurationHours < 1 {
				fmt.Println("Duration cannot be less than 1. Setting it to 1")
				config.DefaultDurationHours = 1
			} else if config.DefaultDurationHours > 12 {
				fmt.Println("Duration cannot be greater than 12. Setting it to 12")
				config.DefaultDurationHours = 12
			}
		}
	}

	return saveConfig(configs)
}

func configProfile(profile string) error {
	configs, err := loadConfig()
	if profile == "" {
		profile = "default"
	}

	if err != nil {
		if errors.Is(err, fileNotFoundError) {
			configs = map[string]*configuration{
				profile: {},
			}
		} else {
			return err
		}
	}

	if _, ok := configs[profile]; !ok {
		configs[profile] = &configuration{}
	}

	prompter := Prompter{}
	configs[profile].AzureTenantId, err = prompter.Prompt("Azure Tenant Id", configs[profile].AzureTenantId)
	configs[profile].AzureAppIdUri, err = prompter.Prompt("Azure App Id Uri", configs[profile].AzureAppIdUri)
	configs[profile].AzureUsername, err = prompter.Prompt("Azure Username", configs[profile].AzureUsername)
	configs[profile].OktaUsername, err = prompter.Prompt("Okta Username", configs[profile].OktaUsername)
	configs[profile].DefaultRoleArn, err = prompter.Prompt("Default Role Arn", configs[profile].DefaultRoleArn)
	defaultDurationHours, err := prompter.Prompt("Default Duration (Hour)", strconv.Itoa(configs[profile].DefaultDurationHours))

	configs[profile].DefaultDurationHours, err = strconv.Atoi(defaultDurationHours)
	if err != nil {
		fmt.Println("Not a valid duration was entered. Will set the duration to 1.")
		configs[profile].DefaultDurationHours = 1
	} else if configs[profile].DefaultDurationHours < 1 {
		fmt.Println("Duration cannot be less than 1. Setting it to 1")
		configs[profile].DefaultDurationHours = 1
	} else if configs[profile].DefaultDurationHours > 12 {
		fmt.Println("Duration cannot be greater than 12. Setting it to 12")
		configs[profile].DefaultDurationHours = 12
	}

	return saveConfig(configs)
}

func loadConfig() (map[string]*configuration, error) {
	_, err := os.Stat(defaultConfigLocation)
	if os.IsNotExist(err) {
		return nil, fileNotFoundError
	}

	content, err := os.ReadFile(defaultConfigLocation)
	if err != nil {
		return nil, err
	}

	configFile := configurationFile{}
	err = yaml.Unmarshal(content, &configFile)
	return configFile.Configs, nil
}

func saveConfig(configs map[string]*configuration) error {
	_, err := os.Stat(defaultConfigLocation)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(defaultConfigLocation), 0700)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	configFile := configurationFile{
		Version: configFileVersion,
		Configs: configs,
	}
	content, err := yaml.Marshal(configFile)
	if err != nil {
		return err
	}

	return os.WriteFile(defaultConfigLocation, content, 0600)
}
