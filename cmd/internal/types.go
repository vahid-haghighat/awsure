package internal

var configFileVersion = "1.0.0"

type configuration struct {
	AzureTenantId        string `yaml:"azure_tenant_id"`
	AzureAppIdUri        string `yaml:"azure_app_id_uri"`
	AzureUsername        string `yaml:"azure_username"`
	OktaUsername         string `yaml:"okta_username"`
	RememberMe           bool   `yaml:"remember_me"`
	DefaultRoleArn       string `yaml:"default_role_arn"`
	DefaultDurationHours int    `yaml:"default_duration_hours"`
}

type configurationFile struct {
	Version string                    `yaml:"version"`
	Configs map[string]*configuration `yaml:"configs"`
}
