package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"db_host"`
		Port     int    `yaml:"db_port"`
		User     string `yaml:"db_user"`
		Password string `yaml:"db_pass"`
		DBName   string `yaml:"db_name"`
		SSLMode  string `yaml:"db_ssl_mode"`
	} `yaml:"database"`
	JWT struct {
		Secret string `yaml:"jwt_secret"`
		Expire int    `yaml:"jwt_expire"`
	} `yaml:"jwt"`
}

func LoadConfig(path string) (*Config, error) {
	// Чтение файла конфигурации
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Разбор YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	// Переопределение значений из переменных окружения
	// Server
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}

	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Server.Port = port
		}
	}

	// Database
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}

	if dbPortStr := os.Getenv("DB_PORT"); dbPortStr != "" {
		if dbPort, err := strconv.Atoi(dbPortStr); err == nil {
			config.Database.Port = dbPort
		}
	}

	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.Database.User = dbUser
	}

	if dbPass := os.Getenv("DB_PASS"); dbPass != "" {
		config.Database.Password = dbPass
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.DBName = dbName
	}

	if dbSSLMode := os.Getenv("DB_SSL_MODE"); dbSSLMode != "" {
		config.Database.SSLMode = dbSSLMode
	}

	// JWT
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWT.Secret = jwtSecret
	}

	if jwtExpireStr := os.Getenv("JWT_EXPIRE"); jwtExpireStr != "" {
		if jwtExpire, err := strconv.Atoi(jwtExpireStr); err == nil {
			config.JWT.Expire = jwtExpire
		}
	}

	return &config, nil
}
