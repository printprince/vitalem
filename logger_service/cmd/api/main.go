package main

import (
	"fmt"
	"log/slog"
	"logger_service/internal/config"
	"logger_service/internal/handlers"
	"logger_service/internal/middleware"
	"logger_service/internal/service"
	"logger_service/pkg/elasticsearch"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// LoggerService - это сервис для логирования событий в наших микросервисах
func main() {
	// Инициализируем логгер
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))
	logger.Info("Starting logger service")

	// Загружаем конфигурации
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Инициализируем компоненты приложения
	logService := service.NewLogService()

	// Берем переменную окружения для es, если нету то не запускаем приложение
	esURL := os.Getenv("ELASTICSEARCH_URL")
	if esURL == "" {
		logger.Error("ELASTICSEARCH_URL environment variable not set")
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
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-API-Key"},
		ExposeHeaders:    []string{echo.HeaderContentLength},
		AllowCredentials: true,
		MaxAge:           86400, // 24 часа
	}))

	// Регистрируем маршруты
	handlers.RegisterRoutes(e, logService, logger)

	// Защищенные маршруты (требуют JWT аутентификации)
	protectedGroup := e.Group("/protected")
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
