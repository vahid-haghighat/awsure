package types

type Configuration struct {
	Profile      string
	AllProfiles  bool
	ForceRefresh bool
	Mode         string
	NoVerifySSL  bool
	NoPrompt     bool
}
