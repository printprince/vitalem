package main

import (
	"log"
	"net/http"

	"github.com/printprince/vitalem/appointment_service/internal/config"
	"github.com/printprince/vitalem/appointment_service/internal/database"
	"github.com/printprince/vitalem/appointment_service/internal/handlers"
	"github.com/printprince/vitalem/appointment_service/internal/repository"
	"github.com/printprince/vitalem/appointment_service/internal/router"
	"github.com/printprince/vitalem/appointment_service/internal/service"

	"github.com/labstack/echo/v4"
)

func main() {
	// Загрузка конфигурации
	log.Println("🚀 Starting Appointment Service...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	log.Printf("📝 Loaded config: %s v%s (%s)", cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Подключение к базе данных
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Выполнение миграций
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// Создание индексов
	if err := database.CreateIndexes(db); err != nil {
		log.Printf("⚠️ Failed to create indexes: %v", err)
	}

	// Инициализация слоев
	repo := repository.NewAppointmentRepository(db)
	svc := service.NewAppointmentService(repo)
	handler := handlers.NewAppointmentHandler(svc)

	// Настройка Echo сервера
	e := echo.New()
	e.HideBanner = true

	// Настройка маршрутов
	router.SetupRoutes(e, handler, cfg.Auth.JWTSecret)

	// Запуск сервера
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("🎯 Server starting on %s", serverAddr)
	log.Printf("📋 Health check: http://%s/health", serverAddr)
	log.Printf("📚 API documentation: http://%s/api", serverAddr)

	if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
