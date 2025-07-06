package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Глобальная переменная для хранения конфигурации
var (
	config *Config
	once   sync.Once
)

// Config представляет конфигурацию API Gateway
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	JWT      JWTConfig      `yaml:"jwt"`
	Services ServicesConfig `yaml:"services"`
	Logging  LoggingConfig  `yaml:"logging"`
	CORS     CORSConfig     `yaml:"cors"`
}

// ServerConfig настройки HTTP сервера
type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// JWTConfig настройки JWT токенов
type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"`
}

// ServiceConfig настройки отдельного микросервиса
type ServiceConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

// ServicesConfig карта всех микросервисов
type ServicesConfig struct {
	Identity     ServiceConfig `yaml:"identity"`
	Logger       ServiceConfig `yaml:"logger"`
	Specialist   ServiceConfig `yaml:"specialist"`
	Patient      ServiceConfig `yaml:"patient"`
	Appointment  ServiceConfig `yaml:"appointment"`
	Notification ServiceConfig `yaml:"notification"`
	FileServer   ServiceConfig `yaml:"fileserver"`
}

// LoggingConfig настройки логирования
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// CORSConfig настройки CORS
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// LoadConfig загружает конфигурацию из файла с поддержкой переменных окружения
func LoadConfig(path string) (*Config, error) {
	// Чтение файла конфигурации
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Разбор YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	// Переопределение значений из переменных окружения
	overrideFromEnv(&cfg)

	// Сохраняем конфигурацию в глобальную переменную
	once.Do(func() {
		config = &cfg
	})

	return &cfg, nil
}

// GetConfig возвращает текущую конфигурацию
func GetConfig() *Config {
	return config
}

// overrideFromEnv переопределяет настройки из переменных окружения
func overrideFromEnv(cfg *Config) {
	// Server
	if host := os.Getenv("GATEWAY_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if portStr := os.Getenv("GATEWAY_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = port
		}
	}

	// JWT
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWT.Secret = jwtSecret
	}

	// Services URLs
	if identityURL := os.Getenv("IDENTITY_SERVICE_URL"); identityURL != "" {
		cfg.Services.Identity.URL = identityURL
	}
	if loggerURL := os.Getenv("LOGGER_SERVICE_URL"); loggerURL != "" {
		cfg.Services.Logger.URL = loggerURL
	}
	if specialistURL := os.Getenv("SPECIALIST_SERVICE_URL"); specialistURL != "" {
		cfg.Services.Specialist.URL = specialistURL
	}
	if patientURL := os.Getenv("PATIENT_SERVICE_URL"); patientURL != "" {
		cfg.Services.Patient.URL = patientURL
	}
	if appointmentURL := os.Getenv("APPOINTMENT_SERVICE_URL"); appointmentURL != "" {
		cfg.Services.Appointment.URL = appointmentURL
	}
	if notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL"); notificationURL != "" {
		cfg.Services.Notification.URL = notificationURL
	}
	if fileserverURL := os.Getenv("FILESERVER_SERVICE_URL"); fileserverURL != "" {
		cfg.Services.FileServer.URL = fileserverURL
	}

	// Logging level
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}
}

// GetServiceURL возвращает URL сервиса по имени
func (c *Config) GetServiceURL(serviceName string) string {
	switch serviceName {
	case "identity":
		return c.Services.Identity.URL
	case "logger":
		return c.Services.Logger.URL
	case "specialist":
		return c.Services.Specialist.URL
	case "patient":
		return c.Services.Patient.URL
	case "appointment":
		return c.Services.Appointment.URL
	case "notification":
		return c.Services.Notification.URL
	case "fileserver":
		return c.Services.FileServer.URL
	default:
		return ""
	}
}
