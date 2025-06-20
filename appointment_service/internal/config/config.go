package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config - конфигурация приложения
type Config struct {
	Server   ServerConfig   `yaml:"server" json:"server"`
	Database DatabaseConfig `yaml:"database" json:"database"`
	Logging  LoggingConfig  `yaml:"logging" json:"logging"`
	App      AppConfig      `yaml:"app" json:"app"`
	Auth     AuthConfig     `yaml:"auth" json:"auth"`
	Meeting  MeetingConfig  `yaml:"meeting" json:"meeting"`
}

// ServerConfig - конфигурация сервера
type ServerConfig struct {
	Host string `yaml:"host" json:"host"`
	Port string `yaml:"port" json:"port"`
}

// DatabaseConfig - конфигурация базы данных
type DatabaseConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	DBName   string `yaml:"db_name" json:"db_name"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`
}

// LoggingConfig - конфигурация логирования
type LoggingConfig struct {
	ConsoleLevel string `yaml:"console_level" json:"console_level"`
	ServiceLevel string `yaml:"service_level" json:"service_level"`
	ServiceURL   string `yaml:"service_url" json:"service_url"`
}

// AppConfig - конфигурация приложения
type AppConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Environment string `yaml:"environment" json:"environment"`
}

// AuthConfig - конфигурация авторизации
type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret" env:"JWT_SECRET"`
}

// MeetingConfig - конфигурация онлайн встреч
type MeetingConfig struct {
	PlatformURL string `yaml:"platform_url" env:"MEETING_PLATFORM_URL" envDefault:"https://meet.vitalem.kz"`
}

var (
	config *Config
	once   sync.Once
)

// LoadConfig - загрузка конфигурации из YAML файла
func LoadConfig() (*Config, error) {
	// Определяем путь к конфигурационному файлу
	configPath := getConfigPath()

	// Читаем файл
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Парсим YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Переопределяем значения из переменных окружения, если они есть
	overrideFromEnv(&cfg)

	// Сохраняем конфигурацию в глобальную переменную
	once.Do(func() {
		config = &cfg
	})

	return &cfg, nil
}

// GetConfig - возвращает текущую конфигурацию
func GetConfig() *Config {
	return config
}

// getConfigPath - определение пути к конфигурационному файлу
func getConfigPath() string {
	// Проверяем переменную окружения
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}

	// Ищем в стандартных местах
	possiblePaths := []string{
		"config.yaml",
		"configs/config.yaml",
		"./config.yaml",
		"../config.yaml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// По умолчанию возвращаем config.yaml в текущей директории
	return "config.yaml"
}

// overrideFromEnv - переопределение значений из переменных окружения
func overrideFromEnv(config *Config) {
	// Server
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}

	// Database
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		config.Database.Port = port
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.DBName = dbName
	}
	if sslMode := os.Getenv("DB_SSL_MODE"); sslMode != "" {
		config.Database.SSLMode = sslMode
	}

	// Logging
	if consoleLevel := os.Getenv("CONSOLE_LOG_LEVEL"); consoleLevel != "" {
		config.Logging.ConsoleLevel = consoleLevel
	}
	if serviceLevel := os.Getenv("SERVICE_LOG_LEVEL"); serviceLevel != "" {
		config.Logging.ServiceLevel = serviceLevel
	}
	if serviceURL := os.Getenv("LOGGER_SERVICE_URL"); serviceURL != "" {
		config.Logging.ServiceURL = serviceURL
	}

	// App
	if env := os.Getenv("APP_ENVIRONMENT"); env != "" {
		config.App.Environment = env
	}

	// Auth
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.Auth.JWTSecret = jwtSecret
	}

	// Meeting
	if platformURL := os.Getenv("MEETING_PLATFORM_URL"); platformURL != "" {
		config.Meeting.PlatformURL = platformURL
	}
}
