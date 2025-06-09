package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"gopkg.in/yaml.v3"
)

// Глобальная переменная для хранения конфигурации
var (
	config *Config
	once   sync.Once
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
	RabbitMQ struct {
		Host       string `yaml:"rmq_host"`
		Port       string `yaml:"rmq_port"`
		User       string `yaml:"rmq_user"`
		Password   string `yaml:"rmq_pass"`
		Exchange   string `yaml:"rmq_exchange"`
		UserQueue  string `yaml:"rmq_user_queue"`
		RoutingKey string `yaml:"rmq_routing_key"`
	} `yaml:"rabbitmq"`
	Logging *LoggingConfig `yaml:"logging"`
}

// LoggingConfig содержит настройки логирования
type LoggingConfig struct {
	ConsoleLevel string `yaml:"console_level"` // Уровень логов для консоли (error, warn, info, debug)
	ServiceLevel string `yaml:"service_level"` // Уровень логов для logger_service
	ServiceURL   string `yaml:"service_url"`   // URL logger_service
}

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
	// Server
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = port
		}
	}

	// Database
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}

	if dbPortStr := os.Getenv("DB_PORT"); dbPortStr != "" {
		if dbPort, err := strconv.Atoi(dbPortStr); err == nil {
			cfg.Database.Port = dbPort
		}
	}

	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}

	if dbPass := os.Getenv("DB_PASS"); dbPass != "" {
		cfg.Database.Password = dbPass
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Database.DBName = dbName
	}

	if dbSSLMode := os.Getenv("DB_SSL_MODE"); dbSSLMode != "" {
		cfg.Database.SSLMode = dbSSLMode
	}

	// JWT
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWT.Secret = jwtSecret
	}

	if jwtExpireStr := os.Getenv("JWT_EXPIRE"); jwtExpireStr != "" {
		if jwtExpire, err := strconv.Atoi(jwtExpireStr); err == nil {
			cfg.JWT.Expire = jwtExpire
		}
	}

	// RabbitMQ
	if rmqHost := os.Getenv("RMQ_HOST"); rmqHost != "" {
		cfg.RabbitMQ.Host = rmqHost
	}

	if rmqPort := os.Getenv("RMQ_PORT"); rmqPort != "" {
		cfg.RabbitMQ.Port = rmqPort
	}

	if rmqUser := os.Getenv("RMQ_USER"); rmqUser != "" {
		cfg.RabbitMQ.User = rmqUser
	}

	if rmqPass := os.Getenv("RMQ_PASS"); rmqPass != "" {
		cfg.RabbitMQ.Password = rmqPass
	}

	if rmqExchange := os.Getenv("RMQ_EXCHANGE"); rmqExchange != "" {
		cfg.RabbitMQ.Exchange = rmqExchange
	}

	if rmqUserQueue := os.Getenv("RMQ_USER_QUEUE"); rmqUserQueue != "" {
		cfg.RabbitMQ.UserQueue = rmqUserQueue
	}

	if rmqRoutingKey := os.Getenv("RMQ_ROUTING_KEY"); rmqRoutingKey != "" {
		cfg.RabbitMQ.RoutingKey = rmqRoutingKey
	}

	// Logging
	if loggerURL := os.Getenv("LOGGER_SERVICE_URL"); loggerURL != "" {
		if cfg.Logging == nil {
			cfg.Logging = &LoggingConfig{}
		}
		cfg.Logging.ServiceURL = loggerURL
	}

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
