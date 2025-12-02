package myyaml

type Config struct {
	BaseConfig `mapstructure:",squash" yaml:",inline"`
}

type BaseConfig struct {
	Application Application `yaml:"application" json:"application"`
	Status      Status      `yaml:"status" json:"status"`
	Logger      Logger      `yaml:"logger" json:"logger"`
}

type Application struct {
	Name        string            `yaml:"name" json:"name"`
	Environment string            `yaml:"environment" json:"environment"`
	Labels      map[string]string `yaml:"labels" json:"labels"`
	BuildInfo   BuildInfo
}

type BuildInfo struct {
	Component `mapstructure:",squash" yaml:",inline"`

	Components []Component `json:"components,omitempty"`
}

type Component struct {
	Branch    string `json:"branch,omitempty"`
	Org       string `json:"org,omitempty"`
	Product   string `json:"product,omitempty"`
	Repo      string `json:"repo,omitempty"`
	SHA       string `json:"sha,omitempty"`
	Version   string `json:"version,omitempty"`
	BuildTime string `json:"buildTime,omitempty"`
}

type Status struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Status.Address is the address to listen on for status reporting
	Address string `yaml:"address" json:"address" default:":8888"`
}

type Logger struct {
	Source bool   `yaml:"source" json:"source"`
	Level  string `yaml:"level" json:"level" default:"info"`
}
