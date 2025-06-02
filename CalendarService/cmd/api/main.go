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

	"CalendarService/internal/config"
	"CalendarService/internal/delivery/http/router"
	"CalendarService/internal/domain/models"
	"CalendarService/internal/domain/repository"
	"CalendarService/internal/infrastructure/notification"
	"CalendarService/internal/service"
	"CalendarService/pkg/logger"

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
	logg := logger.NewLogger(cfg.Logger.Level)

	// Создаем DSN для подключения к PostgreSQL
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password,
		cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	// Подключаемся к базе данных через GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logg.Errorf("failed to connect to database: %v", err)
		log.Fatalf("failed to connect to database: %v", err)
	}

	logg.Info("Successfully connected to database")

	// Выполняем автомиграции для создания таблиц
	err = db.AutoMigrate(&models.Event{}, &models.Doctor{})
	if err != nil {
		logg.Errorf("failed to run migrations: %v", err)
		log.Fatalf("failed to run migrations: %v", err)
	}

	logg.Info("Database migrations completed successfully")

	// Создаем репозитории с GORM
	eventRepo := repository.NewGormEventRepository(db)
	doctorRepo := repository.NewGormDoctorRepository(db)

	// Создаем клиент уведомлений
	notifClient := notification.NewClient(cfg.Notification.URL)

	// Создаем сервис календаря с репозиторием и клиентом уведомлений
	calService := service.NewCalendarService(eventRepo, doctorRepo, notifClient, logg)

	// Создаем роутер и передаем зависимости
	r := router.NewRouter(calService, logg)

	// Запускаем сервер в отдельной горутине
	serverErrCh := make(chan error)
	go func() {
		logg.Infof("Starting api at %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := r.Start(cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			serverErrCh <- err
		}
	}()

	// Ожидаем сигналы прерывания для корректного завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logg.Infof("Shutdown signal received: %v", sig)
	case err := <-serverErrCh:
		logg.Errorf("Server error: %v", err)
	}

	// Создаем контекст с таймаутом для graceful shutdown
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := r.Shutdown(ctxShutDown); err != nil {
		logg.Errorf("Server shutdown error: %v", err)
	} else {
		logg.Info("Server shutdown completed")
	}
}
