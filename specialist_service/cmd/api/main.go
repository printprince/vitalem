package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/specialist_service/internal/config"
	"github.com/printprince/vitalem/specialist_service/internal/handlers"
	"github.com/printprince/vitalem/specialist_service/internal/models"
	"github.com/printprince/vitalem/specialist_service/internal/repository"
	"github.com/printprince/vitalem/specialist_service/internal/service"
	"github.com/printprince/vitalem/utils/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var loggerClient *logger.Client

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация логгера
	if cfg.Logging != nil && cfg.Logging.ServiceURL != "" {
		loggerClient = logger.NewClient(cfg.Logging.ServiceURL, "specialist_service", "", logger.WithAsync(3))
		if loggerClient != nil {
			defer loggerClient.Close()
			//Тестовый лог
			loggerClient.Info("Specialist service started", map[string]interface{}{
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
	err = db.AutoMigrate(&models.Doctor{})
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
	doctorRepo := repository.NewDoctorRepository(db, loggerClient)

	// 2. Сервисы
	doctorService := service.NewDoctorService(doctorRepo, loggerClient)
	eventService := service.NewEventService(doctorService, loggerClient)

	messageService, err := service.NewMessageService(
		rabbitMQURL,
		cfg.RabbitMQ.Exchange,
		cfg.RabbitMQ.DoctorQueueName,
		cfg.RabbitMQ.UserQueueName,
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
		if loggerClient != nil {
			loggerClient.Info("MessageService успешно инициализирован", map[string]interface{}{
				"exchange":      cfg.RabbitMQ.Exchange,
				"userQueueName": cfg.RabbitMQ.UserQueueName,
				"routingKey":    cfg.RabbitMQ.RoutingKey,
			})
		}
		log.Printf("MessageService успешно инициализирован")
	}

	// Запуск потребителя сообщений о создании пользователей
	if messageService != nil {
		if loggerClient != nil {
			loggerClient.Info("Запуск consumer для user created events", nil)
		}
		log.Printf("Запуск consumer для user created events")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		userEvents, err := messageService.ConsumeUserCreated(ctx)
		if err != nil {
			if loggerClient != nil {
				loggerClient.Error("Ошибка подписки на события создания пользователей", map[string]interface{}{
					"error": err.Error(),
				})
			}
			log.Printf("Ошибка подписки на события создания пользователей: %v", err)
		} else {
			if loggerClient != nil {
				loggerClient.Info("Consumer успешно запущен", nil)
			}
			log.Printf("Consumer успешно запущен")

			// Обработка сообщений о создании пользователей через eventService
			go func() {
				for event := range userEvents {
					if loggerClient != nil {
						loggerClient.Info("Получено событие создания пользователя", map[string]interface{}{
							"userID": event.UserID,
							"email":  event.Email,
							"role":   event.Role,
						})
					}
					log.Printf("Получено событие создания пользователя: %s, %s, %s", event.UserID, event.Email, event.Role)

					if err := eventService.ProcessUserCreatedEvent(context.Background(), event); err != nil {
						if loggerClient != nil {
							loggerClient.Error("Ошибка обработки события", map[string]interface{}{
								"error":  err.Error(),
								"userID": event.UserID,
							})
						}
						log.Printf("Ошибка обработки события: %v", err)
					}
				}
			}()
		}
	} else {
		if loggerClient != nil {
			loggerClient.Error("MessageService равен nil, consumer не будет запущен", nil)
		}
		log.Printf("MessageService равен nil, consumer не будет запущен")
	}

	// 3. HTTP обработчики
	doctorHandlers := handlers.NewDoctorHandlers(doctorService, loggerClient)

	e := echo.New()
	e.Validator = middleware.NewValidator()
	e.Use(middleware.LoggerMiddleware()) // Кастомное middleware без health check логирования
	e.Use(echomiddleware.Recover())
	e.Use(middleware.CORSMiddleware())

	// Добавляем эндпоинт проверки здоровья
	e.HEAD("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// Регистрируем маршруты
	doctorHandlers.RegisterRoutes(e)

	// Защищенные маршруты (требуют JWT аутентификации)
	protectedGroup := e.Group("")

	// Debug: выводим информацию о JWT секрете
	if loggerClient != nil {
		loggerClient.Info("JWT configuration loaded", map[string]interface{}{
			"jwt_secret_length":  len(cfg.JWT.Secret),
			"jwt_secret_preview": cfg.JWT.Secret[:8] + "...",
			"jwt_expire":         cfg.JWT.Expire,
		})
	}

	protectedGroup.Use(middleware.JWTMiddleware(cfg.JWT.Secret))
	doctorHandlers.RegisterProtectedRoutes(protectedGroup)

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
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
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
