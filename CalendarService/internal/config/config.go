package config

import (
	"os"

	"fmt"
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`

	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"name"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`

	Notification struct {
		URL string `yaml:"url"`
	} `yaml:"notification"`

	Logger struct {
		Level       string `yaml:"level"`
		ServiceURL  string `yaml:"service_url"`
		ServiceName string `yaml:"service_name"`
	} `yaml:"logger"`

	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`
}

// LoadConfig читает YAML и позволяет override через env-переменные
func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	// Load YAML config
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(yamlFile, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	// Override server config
	if v := os.Getenv("SERVER_HOST"); v != "" {
		config.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.Server.Port = port
		}
	}

	// Override database config
	if v := os.Getenv("DB_HOST"); v != "" {
		config.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		config.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		config.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		config.Database.DBName = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		config.Database.SSLMode = v
	}

	// Override notification.url
	if v := os.Getenv("NOTIFICATION_URL"); v != "" {
		config.Notification.URL = v
	}
	// Override logger.level
	if v := os.Getenv("LOGGER_LEVEL"); v != "" {
		config.Logger.Level = v
	}

	return config, nil
}

// PostgresDSN формирует строку подключения
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.DBName, c.Database.SSLMode)
}
