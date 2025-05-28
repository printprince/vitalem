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

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/patient_service/internal/config"
	"github.com/printprince/vitalem/patient_service/internal/handlers"
	jwtmiddleware "github.com/printprince/vitalem/patient_service/internal/middleware"
	"github.com/printprince/vitalem/patient_service/internal/models"
	"github.com/printprince/vitalem/patient_service/internal/repository"
	"github.com/printprince/vitalem/patient_service/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CustomValidator содержит валидатор запросов
type CustomValidator struct {
	validator *validator.Validate
}

// Validate проверяет данные на соответствие правилам валидации
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

var loggerClient *logger.Client

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация логгера
	if cfg.Logging != nil && cfg.Logging.ServiceURL != "" {
		loggerClient = logger.NewClient(cfg.Logging.ServiceURL, "patient_service", "", logger.WithAsync(3))
		if loggerClient != nil {
			defer loggerClient.Close()
			//Тестовый лог
			loggerClient.Info("Patient service started", map[string]interface{}{
				"config_loaded": true,
			})
		}
	}

	// Подключаемся к базе данных
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		if loggerClient != nil {
			loggerClient.Error("Ошибка подключения к базе данных", map[string]interface{}{
				"error": err.Error(),
			})
		}
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	if loggerClient != nil {
		loggerClient.Info("Успешное подключение к базе данных", nil)
	}

	// Миграция моделей
	err = db.AutoMigrate(&models.Patient{})
	if err != nil {
		if loggerClient != nil {
			loggerClient.Error("Ошибка миграции моделей", map[string]interface{}{
				"error": err.Error(),
			})
		}
		log.Fatalf("Ошибка миграции моделей: %v", err)
	}

	if loggerClient != nil {
		loggerClient.Info("Миграция моделей успешно выполнена", nil)
	}

	// Настройка RabbitMQ URL
	rabbitMQURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)

	// Инициализируем слои приложения
	// 1. Репозиторий
	patientRepo := repository.NewPatientRepository(db, loggerClient)

	// 2. Сервисы
	patientService := service.NewPatientService(patientRepo, loggerClient)
	eventService := service.NewEventService(patientService, loggerClient)

	messageService, err := service.NewMessageService(
		rabbitMQURL,
		cfg.RabbitMQ.Exchange,
		cfg.RabbitMQ.PatientQueue,
		cfg.RabbitMQ.UserQueue,
		cfg.RabbitMQ.RoutingKey,
		loggerClient,
	)
	if err != nil {
		if loggerClient != nil {
			loggerClient.Error("Ошибка инициализации сервиса сообщений", map[string]interface{}{
				"error": err.Error(),
			})
		}
		log.Printf("Ошибка инициализации сервиса сообщений: %v", err)
	} else {
		defer messageService.Close()
	}

	// Запуск потребителя сообщений о создании пользователей
	if messageService != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		userEvents, err := messageService.ConsumeUserCreated(ctx)
		if err != nil {
			if loggerClient != nil {
				loggerClient.Error("Ошибка подписки на события создания пользователей", map[string]interface{}{
					"error": err.Error(),
				})
			}
		} else {
			// Обработка сообщений о создании пользователей через eventService
			go func() {
				for event := range userEvents {
					if err := eventService.ProcessUserCreatedEvent(context.Background(), event); err != nil {
						if loggerClient != nil {
							loggerClient.Error("Ошибка обработки события", map[string]interface{}{
								"error":  err.Error(),
								"userID": event.UserID,
							})
						}
					}
				}
			}()
		}
	}

	// 3. HTTP обработчики
	patientHandlers := handlers.NewPatientHandlers(patientService, loggerClient)

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{echo.HeaderContentLength},
		AllowCredentials: true,
		MaxAge:           86400, // 24 часа
	}))

	// Регистрируем публичные маршруты
	publicRoutes := e.Group("")
	patientHandlers.RegisterPublicRoutes(publicRoutes)

	// Регистрируем защищенные маршруты с JWT middleware
	protectedRoutes := e.Group("/api")
	protectedRoutes.Use(jwtmiddleware.JWTMiddleware(cfg.JWT.Secret))
	patientHandlers.RegisterProtectedRoutes(protectedRoutes)

	// Создаем канал для отслеживания сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		if loggerClient != nil {
			loggerClient.Info("Сервер запущен", map[string]interface{}{
				"port": cfg.Server.Port,
			})
		}

		serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			if loggerClient != nil {
				loggerClient.Error("Ошибка запуска сервера", map[string]interface{}{
					"error": err.Error(),
				})
			}
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	<-quit
	if loggerClient != nil {
		loggerClient.Info("Завершение работы сервера...", nil)
	}

	// Создаем контекст с таймаутом для graceful shutdown
	shutdownTimeout, err := time.ParseDuration(cfg.Server.ShutdownTimeout)
	if err != nil {
		shutdownTimeout = 5 * time.Second // Значение по умолчанию
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Останавливаем сервер
	if err := e.Shutdown(ctx); err != nil {
		if loggerClient != nil {
			loggerClient.Error("Ошибка остановки сервера", map[string]interface{}{
				"error": err.Error(),
			})
		}
		log.Fatalf("Ошибка остановки сервера: %v", err)
	}

	if loggerClient != nil {
		loggerClient.Info("Сервер остановлен", nil)
	}
}
