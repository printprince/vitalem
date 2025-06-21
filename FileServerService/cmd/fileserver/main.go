package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fileserver/internal/config"
	"fileserver/internal/http/logger"
	"fileserver/internal/http/router"
	"fileserver/internal/repository"
	"fileserver/internal/service"
	"fileserver/internal/storage"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация логгера
	logger.InitLogger(cfg.Env == "production")
	defer logger.Sync()

	// Инициализация базы данных
	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{})
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}

	// Инициализация MinIO
	minioClient, err := storage.NewMinioClient(cfg.MinIO)
	if err != nil {
		logger.Fatalf("failed to initialize MinIO: %v", err)
	}

	// Репозиторий, сервис, хендлеры
	fileRepo := repository.NewFileRepository(db)
	fileService := service.NewFileService(fileRepo, minioClient)
	r := router.NewRouter(fileService, cfg.JWTSecret)

	// HTTP сервер
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Грейсфул шатдаун
	go func() {
		logger.Infof("🚀 Starting server on port %s...", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	// Ждём сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown error: %v", err)
	}
	logger.Info("Server exited")
}
