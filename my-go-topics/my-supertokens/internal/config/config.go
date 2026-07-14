package config

// Config holds all application configuration.
type Config struct {
	Server     ServerConfig
	SuperTokens SuperTokensConfig
}

type ServerConfig struct {
	Port string
}

type SuperTokensConfig struct {
	ConnectionURI string
	APIKey        string
	APIDomain     string
	WebsiteDomain string
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			Port: "8080",
		},
		SuperTokens: SuperTokensConfig{
			ConnectionURI: "http://localhost:3567",
			APIKey:        "",
			APIDomain:     "http://localhost:8080",
			WebsiteDomain: "http://localhost:3000",
		},
	}
}
