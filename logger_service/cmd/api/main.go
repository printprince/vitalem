package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/printprince/vitalem/logger_service/internal/config"
	"github.com/printprince/vitalem/logger_service/internal/handlers"
	"github.com/printprince/vitalem/logger_service/internal/service"
	"github.com/printprince/vitalem/logger_service/pkg/elasticsearch"
	"github.com/printprince/vitalem/utils/middleware"
)

// LoggerService - это сервис для логирования событий в наших микросервисах
func main() {
	// Загружаем конфигурации
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Настраиваем уровень логирования
	logLevel := slog.LevelInfo // По умолчанию INFO
	if cfg.Logging != nil && cfg.Logging.Level != "" {
		switch strings.ToLower(cfg.Logging.Level) {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		}
	}

	// Инициализируем логгер
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}))
	logger.Info("Starting logger service")

	// Инициализируем компоненты приложения
	logService := service.NewLogService()

	// Получаем URL для Elasticsearch
	esURL := ""
	if cfg.Logging != nil && cfg.Logging.ElasticsearchURL != "" {
		esURL = cfg.Logging.ElasticsearchURL
	} else {
		// Если нет в конфигурации, проверяем переменную окружения
		esURL = os.Getenv("ELASTICSEARCH_URL")
	}

	if esURL == "" {
		logger.Error("Elasticsearch URL not set in config or environment")
		os.Exit(1)
	}

	// Инициализируем es клиент
	esClient := elasticsearch.NewESClient(esURL, "vitalem-logs")
	logService.SetElasticsearchClient(esClient)

	// Проверяем подключение к ES, если не получится то приложение падает
	if err := esClient.Ping(); err != nil {
		logger.Error("Failed to connect to Elasticsearch", "error", err)
		os.Exit(1)
	}

	// Инициализируем echo
	e := echo.New()
	e.Validator = middleware.NewValidator()
	e.Use(middleware.LoggerMiddleware()) // Кастомное middleware без health check логирования
	e.Use(echomiddleware.Recover())
	e.Use(middleware.CORSMiddleware())

	// Добавляем эндпоинт проверки здоровья
	e.HEAD("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// Регистрируем маршруты
	handlers.RegisterRoutes(e, logService, logger)

	// Защищенные маршруты (требуют JWT аутентификации)
	protectedGroup := e.Group("")
	protectedGroup.Use(middleware.JWTMiddleware(cfg.JWT.Secret))

	// Настройка защищенных маршрутов
	protectedGroup.GET("/logs/stats", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "Authenticated access",
		})
	})

	// Запускаем сервер
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Starting server", "address", serverAddr)
	if err := e.Start(serverAddr); err != nil {
		logger.Error("Server shutdown", "error", err)
	}
}
