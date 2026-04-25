package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/C5383717/my-todo/internal/constants"
	"github.com/spf13/viper"
)

// Config holds all application configuration parameters
type Config struct {
	HTTP     HTTPServer `yaml:"http"`
	LogLevel string     `yaml:"logLevel"`
	Database Database   `yaml:"database"`
}

// HTTPServer holds http server config
type HTTPServer struct {
	Address         string        `yaml:"address" default:":8080"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout" default:"5s"`
}

// Database holds the PostgreSQL connection config
type Database struct {
	DSN string `yaml:"dsn"`
}

// LoadConfig reads and validates configuration from the specified config paths
func LoadConfig(paths ...string) (*Config, error) {
	if err := initViper(paths...); err != nil {
		return nil, err
	}

	return unmarshalConfig()
}

func unmarshalConfig() (*Config, error) {
	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func initViper(paths ...string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	for _, path := range paths {
		viper.AddConfigPath(path)
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	viper.WatchConfig()
	viper.SetEnvPrefix(constants.APIName)
	viper.AutomaticEnv()

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	return nil
}
