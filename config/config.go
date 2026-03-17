package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Project Project
	DB      RDB
}

type Project struct {
	Name        string
	Description string
	Version     string
}

type RDB struct {
	User string
	PWD  string
	Host string
	Port int
	Name string
	SSL  string
}

// DSN returns the Data Source Name for connecting to the database.
func (r RDB) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s application_name=dating-mcp connect_timeout=5",
		r.Host,
		r.Port,
		r.User,
		r.PWD,
		r.Name,
		r.SSL,
	)
}

func Load() (*Config, error) {
	cfg := &Config{
		DB: RDB{
			User: getEnv("DB_USER", ""),
			PWD:  getEnv("DB_PWD", ""),
			Host: getEnv("DB_HOST", "localhost"),
			Port: getEnvInt("DB_PORT", 5432),
			Name: getEnv("DB_NAME", ""),
			SSL:  getEnv("DB_SSL", "disable"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DB.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DB.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DB.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
