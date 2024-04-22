package internal

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func ConfigAll() error {
	configs, err := loadConfigs()
	if err != nil {
		if errors.Is(err, fileNotFoundError) {
			return fmt.Errorf("cannot configure all profiles when no profile is configured")
		} else {
			return err
		}
	}

	fmt.Println("Leaving any of the following as empty will keep them unchanged on profiles")
	c, err := askConfig(configuration{}, true)
	if err != nil {
		return err
	}

	for profile, config := range configs {
		fmt.Printf("Updating %s profile\n", profile)
		config.Merge(c)
	}

	return saveConfig(configs)
}

func ConfigProfile(profile string) error {
	configs, err := loadConfigs()
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

	configs[profile], err = askConfig(*configs[profile], false)
	if err != nil {
		return err
	}

	return saveConfig(configs)
}

func ConfigRemove(profile string) error {
	configs, err := loadConfigs()
	if err != nil {
		fmt.Println("There is no configuration file present, so you're probably good to go!")
		return nil
	}

	if _, ok := configs[profile]; !ok {
		fmt.Println("We couldn't find the profile you specified, so you're probably good to go!")
		return nil
	}

	delete(configs, profile)
	fmt.Printf("Removed %s profile\n", profile)

	return saveConfig(configs)
}

func ConfigImport(importPath string) error {
	stat, err := os.Stat(importPath)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is a directory. please provide a valid config file", importPath)
	}

	_, err = os.Stat(defaultConfigLocation)
	if err == nil {
		fmt.Println("Backing up existing config file")
		err = os.Rename(defaultConfigLocation, fmt.Sprintf("%s.bkp", defaultConfigLocation))
		if err != nil {
			return err
		}
	}

	configs, err := loadConfigsFrom(importPath)
	if err != nil {
		return err
	}

	err = saveConfig(configs)
	if err != nil {
		return err
	}

	fmt.Println("Config has been imported")

	return nil
}

func ConfigExport(exportPath string) error {
	stat, err := os.Stat(exportPath)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		updatedExportPath := filepath.Join(exportPath, "configs.yaml")
		fmt.Printf("%s is a directory. configs will be stored as %s\n", exportPath, updatedExportPath)
		exportPath = updatedExportPath
	}

	_, err = loadConfigs()
	if err != nil {
		return err
	}

	srcStat, err := os.Stat(defaultConfigLocation)
	if err != nil {
		return err
	}

	src, err := os.Open(defaultConfigLocation)
	if err != nil {
		return err
	}
	defer func(src *os.File) {
		err = src.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(src)

	dest, err := os.Create(exportPath)
	if err != nil {
		return err
	}
	defer func(dest *os.File) {
		err = dest.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(dest)

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	err = dest.Sync()
	if err != nil {
		return err
	}

	return os.Chmod(exportPath, srcStat.Mode())
}

func loadConfigs() (map[string]*configuration, error) {
	return loadConfigsFrom(defaultConfigLocation)
}

func loadConfigsFrom(path string) (map[string]*configuration, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fileNotFoundError
	}

	content, err := os.ReadFile(path)
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
		err = os.MkdirAll(filepath.Dir(defaultConfigLocation), 0755)
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

	return os.WriteFile(defaultConfigLocation, content, 0644)
}

func askConfig(config configuration, allowEmpty bool) (*configuration, error) {
	prompter := Prompter{}

	var err error
	config.AzureTenantId, err = prompter.Prompt("Azure Tenant Id", config.AzureTenantId)
	config.AzureAppIdUri, err = prompter.Prompt("Azure App Id Uri", config.AzureAppIdUri)
	config.AzureUsername, err = prompter.Prompt("Azure Username", config.AzureUsername)
	config.OktaUsername, err = prompter.Prompt("Okta Username", config.OktaUsername)
	config.Region, err = prompter.Prompt("Region", "")
	config.DefaultJumpRole, err = prompter.Prompt("Default Jump Role", config.DefaultJumpRole)
	defaultDurationHours, err := prompter.Prompt("Default Duration (Hour)", strconv.Itoa(config.DefaultDurationHours))

	config.DefaultDurationHours, err = strconv.Atoi(defaultDurationHours)
	if err != nil {
		if allowEmpty {
			config.DefaultDurationHours = -1
		} else {
			fmt.Println("Not a valid duration was entered. Will set the duration to 1.")
			config.DefaultDurationHours = 1
		}
	} else if config.DefaultDurationHours < 1 {
		fmt.Println("Duration cannot be less than 1. Setting it to 1")
		config.DefaultDurationHours = 1
	} else if config.DefaultDurationHours > 12 {
		fmt.Println("Duration cannot be greater than 12. Setting it to 12")
		config.DefaultDurationHours = 12
	}

	return &config, nil
}

func loadJumpRoleCredentials() (map[string]*jumpRoleCredentials, error) {
	_, err := os.Stat(defaultJumpRoleCredentialsFileLocation)
	if os.IsNotExist(err) {
		return nil, fileNotFoundError
	}

	content, err := os.ReadFile(defaultJumpRoleCredentialsFileLocation)
	if err != nil {
		return nil, err
	}

	file := jumpRoleCredentialsFile{}
	err = yaml.Unmarshal(content, &file)
	return file.Credentials, nil
}

func saveJumpRoleCredentials(credentials map[string]*jumpRoleCredentials) error {
	_, err := os.Stat(defaultJumpRoleCredentialsFileLocation)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(defaultJumpRoleCredentialsFileLocation), 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	file := jumpRoleCredentialsFile{
		Version:     configFileVersion,
		Credentials: credentials,
	}
	content, err := yaml.Marshal(file)
	if err != nil {
		return err
	}

	return os.WriteFile(defaultJumpRoleCredentialsFileLocation, content, 0644)
}
