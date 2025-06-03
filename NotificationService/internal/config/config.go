package config

import (
	"fmt"
	"os"
	"strconv"

	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config основная структура конфигурации приложения
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	Logger   LoggerConfig   `yaml:"logger"`
	Auth     AuthConfig     `yaml:"auth"`
	SMTP     SMTPConfig     `yaml:"email"`
	Telegram TelegramConfig `yaml:"telegram"`
}

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// TelegramConfig — настройки для Telegram Bot API
type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

// DatabaseConfig конфигурация подключения к Postgres
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

// LoggerConfig уровень логирования
type LoggerConfig struct {
	Level       string `yaml:"level"`
	ServiceURL  string `yaml:"service_url"`
	ServiceName string `yaml:"service_name"`
}

// AuthConfig конфиг для JWT и т.п.
type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

// ExternalConfig внешние сервисы, например Telegram, OpenAI и др.

// LoadConfig читает YAML-файл конфигурации
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	overrideFromEnv(&cfg)

	return &cfg, nil
}

// SMTPConfig — настройки SMTP для отправки email
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

// PostgresDSN формирует строку подключения
func (d *DatabaseConfig) PostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

// overrideFromEnv позволяет переопределять значения из ENV
func overrideFromEnv(cfg *Config) {
	// Server config
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}

	// Logger config
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logger.Level = level
	}

	// Auth config
	if jwt := os.Getenv("JWT_SECRET"); jwt != "" {
		cfg.Auth.JWTSecret = jwt
	}

	// Database config
	if dbhost := os.Getenv("DB_HOST"); dbhost != "" {
		cfg.Database.Host = dbhost
	}
	if dbport := os.Getenv("DB_PORT"); dbport != "" {
		if p, err := strconv.Atoi(dbport); err == nil {
			cfg.Database.Port = p
		}
	}
	if dbuser := os.Getenv("DB_USER"); dbuser != "" {
		cfg.Database.User = dbuser
	}
	if dbpass := os.Getenv("DB_PASSWORD"); dbpass != "" {
		cfg.Database.Password = dbpass
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		cfg.Database.DBName = dbname
	}
	if sslmode := os.Getenv("DB_SSLMODE"); sslmode != "" {
		cfg.Database.SSLMode = sslmode
	}

	// SMTP config
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		cfg.SMTP.Host = smtpHost
	}
	if smtpPort := os.Getenv("SMTP_PORT"); smtpPort != "" {
		if p, err := strconv.Atoi(smtpPort); err == nil {
			cfg.SMTP.Port = p
		}
	}
	if smtpUser := os.Getenv("SMTP_USERNAME"); smtpUser != "" {
		cfg.SMTP.Username = smtpUser
	}
	if smtpPass := os.Getenv("SMTP_PASSWORD"); smtpPass != "" {
		cfg.SMTP.Password = smtpPass
	}
	if smtpFrom := os.Getenv("SMTP_FROM"); smtpFrom != "" {
		cfg.SMTP.From = smtpFrom
	}

	// Telegram config
	if botToken := os.Getenv("TELEGRAM_BOT_TOKEN"); botToken != "" {
		cfg.Telegram.BotToken = botToken
	}
	if chatID := os.Getenv("TELEGRAM_CHAT_ID"); chatID != "" {
		cfg.Telegram.ChatID = chatID
	}
}

// RabbitMQConfig конфигурация RabbitMQ
type RabbitMQConfig struct {
	URL string `yaml:"url"`
}
