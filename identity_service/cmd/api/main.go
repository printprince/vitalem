package main

import (
	"context"
	"fmt"
	"identity_service/internal/config"
	"identity_service/internal/handlers"
	"identity_service/internal/models"
	"identity_service/internal/repository"
	"identity_service/internal/service"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/utils/middleware"
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

	// Инициализируем logger клиент для отправки логов в logger_service
	// Сначала проверяем настройки в конфигурации
	loggerServiceURL := ""
	if cfg.Logging != nil && cfg.Logging.ServiceURL != "" {
		loggerServiceURL = cfg.Logging.ServiceURL
	} else {
		// Если нет в конфигурации, проверяем переменную окружения
		loggerServiceURL = os.Getenv("LOGGER_SERVICE_URL")
	}

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
	e.Validator = middleware.NewValidator()
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(middleware.CORSMiddleware())

	// Настройка защищенных маршрутов
	protected := e.Group("/api/v1")
	protected.Use(middleware.JWTMiddleware(cfg.JWT.Secret))

	// Инициализируем наши 3 слоя - репозиторий, сервис и обработчики
	userRepository := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepository, cfg.JWT.Secret, cfg.JWT.Expire)

	// Подключаем RabbitMQ если настроен
	if cfg.RabbitMQ.Host != "" {
		rabbitMQURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
			cfg.RabbitMQ.User,
			cfg.RabbitMQ.Password,
			cfg.RabbitMQ.Host,
			cfg.RabbitMQ.Port,
		)

		// Инициализируем сервис сообщений
		messageService, err := service.NewMessageService(
			rabbitMQURL,
			cfg.RabbitMQ.Exchange,
			cfg.RabbitMQ.UserQueue,
			cfg.RabbitMQ.RoutingKey,
			appLogger,
		)
		if err != nil {
			logError("Ошибка подключения к RabbitMQ", map[string]interface{}{
				"error": err.Error(),
				"url":   rabbitMQURL,
			})
			log.Printf("Ошибка подключения к RabbitMQ: %v. Работаем без событий.", err)
		} else {
			// Настраиваем корректное завершение работы сервиса сообщений при выходе
			setupMessageServiceShutdown(messageService)

			// Устанавливаем сервис сообщений в authService
			authService.SetMessageService(messageService)
			logInfo("RabbitMQ подключен успешно", nil)
		}
	} else {
		logInfo("RabbitMQ не настроен, работаем без событий", nil)
	}

	// Роуты прописаны в обработчиках
	handlers.RegisterRoutes(e, authService, appLogger)

	logInfo("Сервер запущен", map[string]interface{}{
		"host": cfg.Server.Host,
		"port": cfg.Server.Port,
	})

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := e.Start(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)); err != nil {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидаем сигнала для корректного завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Создаем контекст с таймаутом для завершения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Корректно останавливаем сервер
	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при остановке сервера: %v", err)
	}
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

// setupMessageServiceShutdown настраивает корректное завершение работы сервиса сообщений при выходе
func setupMessageServiceShutdown(messageService service.MessageService) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Завершение работы, закрытие сервиса сообщений...")
		if messageService != nil {
			messageService.Close()
		}
	}()
}

// logInfo отправляет информационный лог
func logInfo(message string, metadata map[string]interface{}) {
	// Отправляем в logger_service
	if appLogger != nil {
		if err := appLogger.Info(message, metadata); err != nil {
			log.Printf("Ошибка отправки лога: %v", err)
		}
	}

	// Проверяем, нужно ли выводить в консоль
	cfg := config.GetConfig()
	if cfg != nil && cfg.Logging != nil && cfg.Logging.ConsoleLevel != "" {
		consoleLevel := strings.ToLower(cfg.Logging.ConsoleLevel)
		if consoleLevel == "debug" || consoleLevel == "info" {
			log.Println(message)
		}
	} else {
		// Если нет конфигурации, выводим по умолчанию
		log.Println(message)
	}
}

// logError отправляет лог об ошибке
func logError(message string, metadata map[string]interface{}) {
	// Отправляем в logger_service
	if appLogger != nil {
		if err := appLogger.Error(message, metadata); err != nil {
			log.Printf("Ошибка отправки лога: %v", err)
		}
	}

	// Ошибки всегда выводим в консоль, если уровень не задан или включает ошибки
	cfg := config.GetConfig()
	if cfg == nil || cfg.Logging == nil || cfg.Logging.ConsoleLevel == "" ||
		strings.ToLower(cfg.Logging.ConsoleLevel) != "none" {
		log.Printf("ОШИБКА: %s", message)
	}
}
