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
	"NotificationService/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Инициализация логгера
	logg := logger.NewLogger(cfg.Logger.Level)
	defer logg.Sync()

	// 3. Подключение к базе данных через GORM
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password,
		cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logg.Fatal("failed to connect to database", "error", err)
	}

	logg.Info("Successfully connected to database")

	// 4. Выполняем автомиграции
	err = db.AutoMigrate(&models.Notification{})
	if err != nil {
		logg.Fatal("failed to run migrations", "error", err)
	}

	logg.Info("Database migrations completed successfully")

	// 5. Инициализация инфраструктуры отправки
	emailSender := email.NewSMTPEmailSender(&cfg.SMTP)
	telegramSender := telegram.NewTelegramSender(&cfg.Telegram)
	codeGenerator := codegen.NewCodeGenerator()

	// 6. Инициализация репозитория и сервиса с GORM
	notifRepo := repository.NewGormNotificationRepository(db)
	notifService := service.NewNotificationService(notifRepo, emailSender, telegramSender, codeGenerator, logg)

	// 7. Инициализация Echo и роутера
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	router.SetupRoutes(e, notifService)

	// Initialize and start RabbitMQ consumer
	consumer, err := messaging.NewConsumer(cfg.RabbitMQ.URL, notifService, logg)
	if err != nil {
		logg.Fatal("Failed to create RabbitMQ consumer", "error", err)
	}
	defer consumer.Close()

	ctx := context.Background()
	if err := consumer.StartConsumer(ctx); err != nil {
		logg.Fatal("Failed to start RabbitMQ consumer", "error", err)
	}

	// 8. Запуск HTTP сервера с graceful shutdown
	serverAddr := ":" + cfg.Server.Port
	go func() {
		logg.Info("starting api", "address", serverAddr)
		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			logg.Fatal("shutting down api due to error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logg.Info("shutting down api...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctxShutdown); err != nil {
		logg.Error("api shutdown failed", "error", err)
	} else {
		logg.Info("api stopped gracefully")
	}
}
