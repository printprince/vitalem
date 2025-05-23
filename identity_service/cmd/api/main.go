package main

import (
	"fmt"
	"identity_service/internal/config"
	"identity_service/internal/handlers"
	"identity_service/internal/models"
	"identity_service/internal/repository"
	"identity_service/internal/service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Глобальная переменная для логгера
var appLogger *logger.Client

func main() {
	// Загружаем конфигурации
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Ошибка получения конфигураций: %v", err)
	}

	// Инициализируем logger клиент для отправки логов в logger_service, если указан URL
	loggerServiceURL := os.Getenv("LOGGER_SERVICE_URL")
	if loggerServiceURL != "" {
		// Создаем клиент логгера с асинхронной отправкой (3 воркера)
		appLogger = logger.NewClient(
			loggerServiceURL,
			"identity_service",
			"",
			logger.WithAsync(3),
			logger.WithTimeout(3*time.Second),
		)

		// Настраиваем корректное завершение работы логгера при выходе
		setupGracefulShutdown(appLogger)

		// Тестовый лог для проверки подключения
		err := appLogger.Info("Identity service started", map[string]interface{}{
			"config_loaded": true,
		})

		if err != nil {
			log.Printf("Ошибка отправки лога: %v", err)
		} else {
			log.Printf("Логирование через logger_service настроено: %s", loggerServiceURL)
		}
	} else {
		log.Println("LOGGER_SERVICE_URL не указан, логирование через logger_service отключено")
	}

	// Создаем переменную с конфигами для базы данных
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password,
		cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	// Подключаемся к базе данных
	// Postgres
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logError("Ошибка подключения к базе данных", map[string]interface{}{"error": err.Error()})
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	logInfo("Успешное подключение к базе данных", nil)

	// Миграция моделей
	err = db.AutoMigrate(&models.Users{})
	if err != nil {
		logError("Ошибка миграции", map[string]interface{}{"error": err.Error()})
		log.Fatalf("Ошибка миграции: %v", err)
	}

	logInfo("Миграция моделей успешно выполнена", nil)

	// Инициализируем echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{echo.HeaderContentLength},
		AllowCredentials: true,
		MaxAge:           86400, // 24 часа
	}))

	// Инициализируем наши 3 слоя - репозиторий, сервис и обработчики
	// Роуты прописаны в обработчиках
	userRepository := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepository, cfg.JWT.Secret, cfg.JWT.Expire)
	handlers.RegisterRoutes(e, authService, appLogger)

	logInfo("Сервер запущен", map[string]interface{}{
		"host": cfg.Server.Host,
		"port": cfg.Server.Port,
	})

	// Запускаем сервер
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)))
}

// setupGracefulShutdown настраивает корректное завершение работы логгера при выходе
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

// logInfo отправляет информационный лог
func logInfo(message string, metadata map[string]interface{}) {
	if appLogger != nil {
		if err := appLogger.Info(message, metadata); err != nil {
			log.Printf("Ошибка отправки лога: %v", err)
		}
	}
	log.Println(message)
}

// logError отправляет лог об ошибке
func logError(message string, metadata map[string]interface{}) {
	if appLogger != nil {
		if err := appLogger.Error(message, metadata); err != nil {
			log.Printf("Ошибка отправки лога: %v", err)
		}
	}
	log.Printf("ОШИБКА: %s", message)
}
