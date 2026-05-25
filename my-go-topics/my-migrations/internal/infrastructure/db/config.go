package db

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"
)

type Config struct {
	DatabaseURL string

	MaxConns int32
	MinConns int32

	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration

	HealthCheckPeriod time.Duration
}

func (c Config) Validate() error {

	if c.DatabaseURL == "" {
		return fmt.Errorf("database url is required")
	}

	if c.MaxConns <= 0 {
		return fmt.Errorf("max_conns must be > 0")
	}

	if c.MinConns < 0 {
		return fmt.Errorf("min_conns must be >= 0")
	}

	return nil
}

func DefaultConfig() Config {
	return Config{
		MaxConns: 20,
		MinConns: 5,

		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,

		HealthCheckPeriod: time.Minute,
	}
}

func BuildDSNFromEnv() string {
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASS")),
		Host:     net.JoinHostPort(os.Getenv("DB_HOST"), dbPort),
		Path:     "/" + os.Getenv("DB_NAME"),
		RawQuery: "sslmode=disable",
	}
	return u.String()
}