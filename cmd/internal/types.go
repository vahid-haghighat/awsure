package internal

var configFileVersion = "1.0.0"

type azureConfig struct {
	azureTenantId        string `yaml:"azure_tenant_id"`
	azureAppIdUri        string `yaml:"azure_app_id_uri"`
	azureUsername        string `yaml:"azure_username"`
	oktaUsername         string `yaml:"okta_username"`
	rememberMe           bool   `yaml:"remember_me"`
	defaultRoleArn       string `yaml:"default_role_arn"`
	defaultDurationHours int    `yaml:"default_duration_hours"`
}

type azureConfigFile struct {
	Version string      `yaml:"version"`
	Config  azureConfig `yaml:"config"`
}
