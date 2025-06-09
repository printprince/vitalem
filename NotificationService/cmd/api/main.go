package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"NotificationService/internal/config"
	"NotificationService/internal/delivery/http/router"
	"NotificationService/internal/domain/models"
	"NotificationService/internal/domain/repository"
	"NotificationService/internal/infrastructure/codegen"
	"NotificationService/internal/infrastructure/email"
	"NotificationService/internal/infrastructure/messaging"
	"NotificationService/internal/infrastructure/telegram"
	"NotificationService/internal/service"

	"github.com/printprince/vitalem/logger_service/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// LoggerInterface - интерфейс для совместимости с логгером
type LoggerInterface interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	Sugar() SugarInterface
}

// SugarInterface - интерфейс для Sugar логгера
type SugarInterface interface {
	Infow(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
}

var loggerClient *logger.Client

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Инициализация логгера
	if cfg.Logging.ServiceURL != "" {
		loggerClient = logger.NewClient(
			cfg.Logging.ServiceURL,
			"notification_service",
			"",
			logger.WithAsync(3),
			logger.WithTimeout(3*time.Second),
		)
		if loggerClient != nil {
			defer loggerClient.Close()
			setupGracefulShutdown(loggerClient)

			// Тестовый лог
			loggerClient.Info("Notification service started", map[string]interface{}{
				"config_loaded": true,
			})
		}
	}

	// 3. Подключение к базе данных через GORM
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password,
		cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logError("Failed to connect to database", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("failed to connect to database: %v", err)
	}

	logInfo("Successfully connected to database")

	// 4. Выполняем автомиграции
	err = db.AutoMigrate(&models.Notification{})
	if err != nil {
		logError("Failed to run migrations", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("failed to run migrations: %v", err)
	}

	logInfo("Database migrations completed successfully")

	// 5. Инициализация инфраструктуры отправки
	emailSender := email.NewSMTPEmailSender(&cfg.SMTP)
	telegramSender := telegram.NewTelegramSender(&cfg.Telegram)
	codeGenerator := codegen.NewCodeGenerator()

	// 6. Инициализация репозитория и сервиса
	notifRepo := repository.NewGormNotificationRepository(db)
	notifService := service.NewNotificationService(notifRepo, emailSender, telegramSender, codeGenerator, createServiceLogger())

	// 7. Инициализация Echo и роутера
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	router.SetupRoutes(e, notifService)

	// Initialize and start RabbitMQ consumer
	consumer, err := messaging.NewConsumer(cfg.RabbitMQ.URL, notifService, createMessagingLogger())
	if err != nil {
		logError("Failed to create RabbitMQ consumer", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}
	defer consumer.Close()

	ctx := context.Background()
	if err := consumer.StartConsumer(ctx); err != nil {
		logError("Failed to start RabbitMQ consumer", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to start RabbitMQ consumer: %v", err)
	}

	// 8. Запуск HTTP сервера с graceful shutdown
	serverAddr := ":" + cfg.Server.Port
	go func() {
		logInfo("Starting API server on %s", serverAddr)
		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			logError("API server error", map[string]interface{}{
				"error": err.Error(),
			})
			log.Fatalf("shutting down api due to error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logInfo("Shutting down API...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctxShutdown); err != nil {
		logError("API shutdown failed", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logInfo("API stopped gracefully")
	}
}

// logInfo отправляет информационный лог
func logInfo(format string, args ...interface{}) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
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

	// Проверяем, нужно ли выводить в консоль (пока выводим всегда)
	if len(args) > 0 {
		log.Printf(format, args...)
	} else {
		log.Println(format)
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
	log.Printf("ОШИБКА: %s", message)
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

// createCompatibleLogger создает обертку для совместимости со старым интерфейсом
func createCompatibleLogger() LoggerInterface {
	return &CompatibleLogger{}
}

// createServiceLogger создает логгер для service пакета
func createServiceLogger() service.LoggerInterface {
	return &ServiceCompatibleLogger{}
}

// createMessagingLogger создает логгер для messaging пакета
func createMessagingLogger() messaging.LoggerInterface {
	return &MessagingCompatibleLogger{}
}

// CompatibleLogger - обертка для совместимости со старым zap логгером
type CompatibleLogger struct{}

func (l *CompatibleLogger) Info(msg string, keysAndValues ...interface{}) {
	metadata := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			metadata[key] = keysAndValues[i+1]
		}
	}

	if loggerClient != nil {
		loggerClient.Info(msg, metadata)
	} else {
		log.Printf("INFO: %s", msg)
	}
}

func (l *CompatibleLogger) Error(msg string, keysAndValues ...interface{}) {
	metadata := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			metadata[key] = keysAndValues[i+1]
		}
	}

	if loggerClient != nil {
		loggerClient.Error(msg, metadata)
	} else {
		log.Printf("ERROR: %s", msg)
	}
}

func (l *CompatibleLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.Error(msg, keysAndValues...)
	os.Exit(1)
}

func (l *CompatibleLogger) Sugar() SugarInterface {
	return &SugarLogger{logger: l}
}

// SugarLogger - обертка для Sugar методов
type SugarLogger struct {
	logger LoggerInterface
}

func (s *SugarLogger) Infow(msg string, keysAndValues ...interface{}) {
	s.logger.Info(msg, keysAndValues...)
}

func (s *SugarLogger) Errorw(msg string, keysAndValues ...interface{}) {
	s.logger.Error(msg, keysAndValues...)
}

func (s *SugarLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	s.logger.Fatal(msg, keysAndValues...)
}

func (s *SugarLogger) Warnw(msg string, keysAndValues ...interface{}) {
	// Обрабатываем как info для упрощения
	s.logger.Info("WARN: "+msg, keysAndValues...)
}

// ServiceCompatibleLogger - обертка для service пакета
type ServiceCompatibleLogger struct{}

func (l *ServiceCompatibleLogger) Info(msg string, keysAndValues ...interface{}) {
	metadata := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			metadata[key] = keysAndValues[i+1]
		}
	}

	if loggerClient != nil {
		loggerClient.Info(msg, metadata)
	} else {
		log.Printf("INFO: %s", msg)
	}
}

func (l *ServiceCompatibleLogger) Error(msg string, keysAndValues ...interface{}) {
	metadata := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			metadata[key] = keysAndValues[i+1]
		}
	}

	if loggerClient != nil {
		loggerClient.Error(msg, metadata)
	} else {
		log.Printf("ERROR: %s", msg)
	}
}

func (l *ServiceCompatibleLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.Error(msg, keysAndValues...)
	os.Exit(1)
}

func (l *ServiceCompatibleLogger) Sugar() service.SugarInterface {
	return &ServiceSugarLogger{logger: l}
}

// ServiceSugarLogger - обертка для Sugar методов service пакета
type ServiceSugarLogger struct {
	logger service.LoggerInterface
}

func (s *ServiceSugarLogger) Infow(msg string, keysAndValues ...interface{}) {
	s.logger.Info(msg, keysAndValues...)
}

func (s *ServiceSugarLogger) Errorw(msg string, keysAndValues ...interface{}) {
	s.logger.Error(msg, keysAndValues...)
}

func (s *ServiceSugarLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	s.logger.Fatal(msg, keysAndValues...)
}

func (s *ServiceSugarLogger) Warnw(msg string, keysAndValues ...interface{}) {
	// Обрабатываем как info для упрощения
	s.logger.Info("WARN: "+msg, keysAndValues...)
}

// MessagingCompatibleLogger - обертка для messaging пакета
type MessagingCompatibleLogger struct{}

func (l *MessagingCompatibleLogger) Info(msg string, keysAndValues ...interface{}) {
	metadata := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			metadata[key] = keysAndValues[i+1]
		}
	}

	if loggerClient != nil {
		loggerClient.Info(msg, metadata)
	} else {
		log.Printf("INFO: %s", msg)
	}
}

func (l *MessagingCompatibleLogger) Error(msg string, keysAndValues ...interface{}) {
	metadata := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			metadata[key] = keysAndValues[i+1]
		}
	}

	if loggerClient != nil {
		loggerClient.Error(msg, metadata)
	} else {
		log.Printf("ERROR: %s", msg)
	}
}

func (l *MessagingCompatibleLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.Error(msg, keysAndValues...)
	os.Exit(1)
}

func (l *MessagingCompatibleLogger) Sugar() messaging.SugarInterface {
	return &MessagingSugarLogger{logger: l}
}

// MessagingSugarLogger - обертка для Sugar методов messaging пакета
type MessagingSugarLogger struct {
	logger messaging.LoggerInterface
}

func (s *MessagingSugarLogger) Infow(msg string, keysAndValues ...interface{}) {
	s.logger.Info(msg, keysAndValues...)
}

func (s *MessagingSugarLogger) Errorw(msg string, keysAndValues ...interface{}) {
	s.logger.Error(msg, keysAndValues...)
}

func (s *MessagingSugarLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	s.logger.Fatal(msg, keysAndValues...)
}

func (s *MessagingSugarLogger) Warnw(msg string, keysAndValues ...interface{}) {
	// Обрабатываем как info для упрощения
	s.logger.Info("WARN: "+msg, keysAndValues...)
}
