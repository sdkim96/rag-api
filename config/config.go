package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Project        Project
	DB             RDB
	AzureBlobStore AzureBlobStore
	AzureCU        AzureCU
	OpenAI         OpenAI
}

type Project struct {
	Name        string
	Description string
	Version     string
}

type AzureBlobStore struct {
	AccountName   string
	ConnString    string
	ContainerName string
}

type AzureCU struct {
	Endpoint string
	APIKey   string
}

type OpenAI struct {
	APIKey string
}

type RDB struct {
	User string
	PWD  string
	Host string
	Port int
	Name string
	SSL  string
}

func (r RDB) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s application_name=rag-api connect_timeout=5",
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
		AzureBlobStore: AzureBlobStore{
			AccountName:   getEnv("AZURE_BLOB_ACCOUNT_NAME", ""),
			ConnString:    getEnv("AZURE_BLOB_CONN_STRING", ""),
			ContainerName: getEnv("AZURE_BLOB_CONTAINER_NAME", ""),
		},
		AzureCU: AzureCU{
			Endpoint: getEnv("AZURE_CU_ENDPOINT", ""),
			APIKey:   getEnv("AZURE_CU_API_KEY", ""),
		},
		OpenAI: OpenAI{
			APIKey: getEnv("OPENAI_API_KEY", ""),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	// DB
	if c.DB.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DB.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DB.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	// Azure Blob
	if c.AzureBlobStore.AccountName == "" {
		return fmt.Errorf("AZURE_BLOB_ACCOUNT_NAME is required")
	}
	if c.AzureBlobStore.ConnString == "" {
		return fmt.Errorf("AZURE_BLOB_CONN_STRING is required")
	}
	if c.AzureBlobStore.ContainerName == "" {
		return fmt.Errorf("AZURE_BLOB_CONTAINER_NAME is required")
	}

	// Azure CU
	if c.AzureCU.Endpoint == "" {
		return fmt.Errorf("AZURE_CU_ENDPOINT is required")
	}
	if c.AzureCU.APIKey == "" {
		return fmt.Errorf("AZURE_CU_API_KEY is required")
	}

	// OpenAI
	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
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
