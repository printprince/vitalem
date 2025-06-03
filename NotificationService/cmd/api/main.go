package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"NotificationService/internal/config"
	"NotificationService/internal/delivery/http/router"
	"NotificationService/internal/domain/models"
	"NotificationService/internal/domain/repository"
	"NotificationService/internal/infrastructure/codegen"
	"NotificationService/internal/infrastructure/email"
	"NotificationService/internal/infrastructure/telegram"
	"NotificationService/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/printprince/vitalem/logger_service/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Инициализируем логгер
	logg := logger.NewClient(
		cfg.Logger.ServiceURL,
		cfg.Logger.ServiceName,
		"",
		logger.WithAsync(3),
		logger.WithTimeout(3*time.Second),
	)

	// Создаем DSN для подключения к PostgreSQL
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password,
		cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	// Подключаемся к базе данных через GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logg.Error("Failed to connect to database", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("failed to connect to database: %v", err)
	}

	logg.Info("Successfully connected to database", nil)

	// Выполняем автомиграции для создания таблиц
	err = db.AutoMigrate(&models.Notification{})
	if err != nil {
		logg.Error("Failed to run migrations", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("failed to run migrations: %v", err)
	}

	logg.Info("Database migrations completed successfully", nil)

	// Создаем репозитории с GORM
	notifRepo := repository.NewGormNotificationRepository(db)

	// Инициализируем инфраструктурные сервисы
	emailSender := email.NewSMTPEmailSender(&cfg.SMTP)
	telegramSender := telegram.NewTelegramSender(&cfg.Telegram)
	codeGenerator := codegen.NewCodeGenerator()

	// Создаем сервис уведомлений
	notifService := service.NewNotificationService(notifRepo, emailSender, telegramSender, codeGenerator, logg)

	// Создаем Echo и настраиваем middleware
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Настраиваем маршруты
	router.SetupRoutes(e, notifService)

	// Запускаем сервер в отдельной горутине
	serverErrCh := make(chan error)
	go func() {
		logg.Info("Starting API", map[string]interface{}{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
		})
		if err := e.Start(cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			serverErrCh <- err
		}
	}()

	// Ожидаем сигналы прерывания для корректного завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logg.Info("Shutdown signal received", map[string]interface{}{
			"signal": sig.String(),
		})
	case err := <-serverErrCh:
		logg.Error("Server error", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Создаем контекст с таймаутом для graceful shutdown
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctxShutDown); err != nil {
		logg.Error("Server shutdown error", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logg.Info("Server shutdown completed", nil)
	}
}
