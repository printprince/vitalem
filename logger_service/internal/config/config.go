package config

import (
	"fmt"
	"os"
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
	JWT struct {
		Secret string `yaml:"jwt_secret"`
		Expire int    `yaml:"jwt_expire"`
	} `yaml:"jwt"`
	Logging *LoggingConfig `yaml:"logging"`
}

// LoggingConfig содержит настройки логирования
type LoggingConfig struct {
	Level            string `yaml:"level"`             // Уровень логов (error, warn, info, debug)
	ElasticsearchURL string `yaml:"elasticsearch_url"` // URL для Elasticsearch
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	// Переопределение значений из переменных окружения
	if elasticURL := os.Getenv("ELASTICSEARCH_URL"); elasticURL != "" {
		if cfg.Logging == nil {
			cfg.Logging = &LoggingConfig{}
		}
		cfg.Logging.ElasticsearchURL = elasticURL
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
