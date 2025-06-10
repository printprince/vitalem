package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/printprince/vitalem/appointment_service/internal/config"
	"github.com/printprince/vitalem/appointment_service/internal/database"
	"github.com/printprince/vitalem/appointment_service/internal/handlers"
	"github.com/printprince/vitalem/appointment_service/internal/repository"
	"github.com/printprince/vitalem/appointment_service/internal/router"
	"github.com/printprince/vitalem/appointment_service/internal/service"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
)

var loggerClient *logger.Client

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация логгера
	if cfg.Logging.ServiceURL != "" {
		loggerClient = logger.NewClient(
			cfg.Logging.ServiceURL,
			"appointment_service",
			"",
			logger.WithAsync(3),
			logger.WithTimeout(3*time.Second),
		)
		if loggerClient != nil {
			defer loggerClient.Close()
			setupGracefulShutdown(loggerClient)

			// Тестовый лог
			loggerClient.Info("Appointment service started", map[string]interface{}{
				"config_loaded": true,
				"version":       cfg.App.Version,
				"environment":   cfg.App.Environment,
			})
		}
	}

	logInfo("Loaded config: %s v%s (%s)", cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Подключение к базе данных
	db, err := database.ConnectDB(cfg)
	if err != nil {
		logError("Failed to connect to database", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Выполнение миграций
	if err := database.RunMigrations(db); err != nil {
		logError("Failed to run migrations", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Создание индексов
	if err := database.CreateIndexes(db); err != nil {
		logError("Failed to create indexes", map[string]interface{}{
			"error": err.Error(),
		})
		log.Printf("Failed to create indexes: %v", err)
	}

	// Инициализация слоев
	repo := repository.NewAppointmentRepository(db)
	svc := service.NewAppointmentService(repo, loggerClient)
	handler := handlers.NewAppointmentHandler(svc)

	// Передаем логгер в хендлер
	if loggerClient != nil {
		handler.SetLogger(loggerClient)
	}

	// Настройка Echo сервера
	e := echo.New()
	e.HideBanner = true

	// Добавляем эндпоинт проверки здоровья
	e.HEAD("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// Настройка маршрутов
	router.SetupRoutes(e, handler, cfg.Auth.JWTSecret)

	// Запуск сервера
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	logInfo("Server starting on %s", serverAddr)
	logInfo("Health check: http://%s/health", serverAddr)
	logInfo("API documentation: http://%s/api", serverAddr)

	if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
		logError("Failed to start server", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to start server: %v", err)
	}
}

// logInfo отправляет информационный лог
func logInfo(format string, args ...interface{}) {
	message := format
	if len(args) > 0 {
		message = strings.ReplaceAll(format, "%s", "%v")
		message = strings.ReplaceAll(message, "%d", "%v")
	}

	// Отправляем в logger_service
	if loggerClient != nil {
		var metadata map[string]interface{}
		if len(args) > 0 {
			metadata = map[string]interface{}{
				"formatted_message": message,
				"args":              args,
			}
		}
		if err := loggerClient.Info(message, metadata); err != nil {
			log.Printf("Ошибка отправки лога: %v", err)
		}
	}

	// Проверяем, нужно ли выводить в консоль
	cfg := config.GetConfig()
	if cfg != nil && cfg.Logging.ConsoleLevel != "" {
		consoleLevel := strings.ToLower(cfg.Logging.ConsoleLevel)
		if consoleLevel == "debug" || consoleLevel == "info" {
			if len(args) > 0 {
				log.Printf(format, args...)
			} else {
				log.Println(format)
			}
		}
	} else {
		// Если нет конфигурации, выводим по умолчанию
		if len(args) > 0 {
			log.Printf(format, args...)
		} else {
			log.Println(format)
		}
	}
}

// logError отправляет лог об ошибке
func logError(message string, metadata map[string]interface{}) {
	// Отправляем в logger_service
	if loggerClient != nil {
		if err := loggerClient.Error(message, metadata); err != nil {
			log.Printf("Ошибка отправки лога: %v", err)
		}
	}

	// Ошибки всегда выводим в консоль
	cfg := config.GetConfig()
	if cfg == nil || cfg.Logging.ConsoleLevel == "" ||
		strings.ToLower(cfg.Logging.ConsoleLevel) != "none" {
		log.Printf("ОШИБКА: %s", message)
	}
}

// setupGracefulShutdown настраивает корректное завершение работы логгера
func setupGracefulShutdown(logger *logger.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Завершение работы, закрытие логгера...")
		if logger != nil {
			logger.Close()
		}
		os.Exit(0)
	}()
}
