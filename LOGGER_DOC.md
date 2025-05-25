# Инструкция по использованию Logger Service в Vitalem

Данная инструкция описывает, как интегрировать и использовать централизованный сервис логирования в микросервисах Vitalem.

## Общий принцип работы

Logger Service предоставляет единый интерфейс для логирования во всех микросервисах. Вместо того чтобы каждый сервис писал логи в свои файлы или разные системы, все логи отправляются в централизованный сервис, который затем сохраняет их в Elasticsearch для последующего анализа через Kibana.

```
┌─────────────────┐     
│  Ваш сервис     │     
│  (микросервис)  │     
└────────┬────────┘     
         │              
         │ HTTP API     
         ▼              
┌─────────────────┐     
│  Logger         │     
│  Service        │     
└────────┬────────┘     
         │              
         │              
         ▼              
┌─────────────────┐     
│  Elasticsearch  │     
└────────┬────────┘     
         │              
         │              
         ▼              
┌─────────────────┐
│  Kibana         │
│  (визуализация) │
└─────────────────┘
```

## Добавление зависимости в ваш сервис

### 1. Добавьте зависимость в ваш проект Go

```bash
go get github.com/printprince/vitalem/logger_service

или

go get github.com/printprince/vitalem
```

### 2. Импортируйте пакет логгера в ваш код

```go
import "github.com/printprince/vitalem/logger_service/pkg/logger"
```

## Инициализация клиента логгера

### Вариант 1: Синхронный клиент (простой, но может замедлять работу)

```go
// Создание клиента с синхронной отправкой логов
loggerClient := logger.NewClient(
    "http://logger_service:8802", // URL сервиса логирования
    "your_service_name",          // Имя вашего сервиса
    "",                           // API ключ (оставьте пустым)
)
```

### Вариант 2: Асинхронный клиент (рекомендуется)

```go
import (
    "github.com/printprince/vitalem/logger_service/pkg/logger"
    "time"
)

// Создание клиента с асинхронной отправкой логов
loggerClient := logger.NewClient(
    "http://logger_service:8802",  // URL сервиса логирования
    "your_service_name",           // Имя вашего сервиса (например, "patient_service")
    "",                            // API ключ (оставьте пустым)
    logger.WithAsync(3),           // Запуск 3 воркеров для асинхронной отправки
    logger.WithTimeout(3*time.Second), // Таймаут HTTP запросов
)

// Важно: закройте клиент при завершении работы
defer loggerClient.Close()
```

## Настройка URL логгера через переменные окружения

Рекомендуется использовать переменные окружения для настройки URL логгера:

```go
import (
    "os"
    "github.com/printprince/vitalem/logger_service/pkg/logger"
    "time"
)

func main() {
    loggerURL := os.Getenv("LOGGER_SERVICE_URL")
    if loggerURL == "" {
        loggerURL = "http://logger_service:8802" // Значение по умолчанию
    }
    
    loggerClient := logger.NewClient(
        loggerURL,
        "your_service_name",
        "",
        logger.WithAsync(3),
        logger.WithTimeout(3*time.Second),
    )
    defer loggerClient.Close()
    
    // Остальной код...
}
```

## Отправка логов разного уровня

Клиент логгера предоставляет методы для отправки логов разных уровней:

```go
// Информационный лог
loggerClient.Info("Сервер запущен", map[string]interface{}{
    "host": "localhost",
    "port": 8801,
})

// Отладочный лог (более детальный)
loggerClient.Debug("Детальная информация", map[string]interface{}{
    "request_id": "abc123",
    "user_id": 42,
})

// Предупреждение
loggerClient.Warn("Повышенное время ответа БД", map[string]interface{}{
    "query": "SELECT * FROM users",
    "latency_ms": 500,
})

// Ошибка
loggerClient.Error("Ошибка подключения к БД", map[string]interface{}{
    "error": err.Error(),
    "component": "database",
})
```

## Структура метаданных

Для единообразия и удобства поиска рекомендуется включать следующие метаданные в логи:

1. Для всех логов:
   - `request_id` - уникальный идентификатор запроса
   - `component` - название компонента (например, "database", "http", "auth")

2. Для логов HTTP запросов:
   - `method` - HTTP метод (GET, POST, etc.)
   - `path` - путь запроса
   - `status_code` - код ответа
   - `duration_ms` - длительность выполнения в миллисекундах

3. Для логов ошибок:
   - `error` - текст ошибки
   - `stack_trace` - стек вызовов (опционально)

## Пример интеграции с middleware Echo

```go
import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/printprince/vitalem/logger_service/pkg/logger"
)

// LoggerMiddleware создает middleware для логирования HTTP запросов
func LoggerMiddleware(loggerClient *logger.Client) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            req := c.Request()
            res := c.Response()
            
            start := time.Now()
            
            // Генерируем уникальный ID запроса
            requestID := req.Header.Get(echo.HeaderXRequestID)
            if requestID == "" {
                requestID = generateRequestID() // Ваша функция генерации ID
            }
            
            // Устанавливаем ID запроса в контекст
            c.Set("request_id", requestID)
            
            // Логируем начало запроса
            loggerClient.Info("HTTP Request", map[string]interface{}{
                "request_id": requestID,
                "method": req.Method,
                "path": req.URL.Path,
                "ip": c.RealIP(),
                "component": "http",
            })
            
            // Выполняем следующий обработчик
            err := next(c)
            
            // Логируем завершение запроса
            duration := time.Since(start)
            metadata := map[string]interface{}{
                "request_id": requestID,
                "method": req.Method,
                "path": req.URL.Path,
                "status": res.Status,
                "duration_ms": duration.Milliseconds(),
                "component": "http",
            }
            
            if err != nil {
                // Если произошла ошибка, логируем её
                metadata["error"] = err.Error()
                loggerClient.Error("HTTP Request Error", metadata)
            } else {
                loggerClient.Info("HTTP Response", metadata)
            }
            
            return err
        }
    }
}

// Регистрация middleware в Echo
func setupRoutes(e *echo.Echo, loggerClient *logger.Client) {
    // Применяем middleware логирования ко всем маршрутам
    e.Use(LoggerMiddleware(loggerClient))
    
    // Остальная настройка маршрутов...
}
```

## Корректное завершение работы

Для асинхронного клиента важно корректно завершить работу, чтобы все логи были отправлены:

```go
import (
    "os"
    "os/signal"
    "syscall"
)

func setupGracefulShutdown(loggerClient *logger.Client) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        fmt.Println("Завершение работы, отправка оставшихся логов...")
        if loggerClient != nil {
            loggerClient.Close() // Дожидается отправки всех логов
        }
        os.Exit(0)
    }()
}

func main() {
    // Инициализация логгера
    loggerClient := logger.NewClient(...)
    
    // Настройка корректного завершения работы
    setupGracefulShutdown(loggerClient)
    
    // Остальной код...
}
```

## Полный пример использования

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/labstack/echo/v4"
    "github.com/printprince/vitalem/logger_service/pkg/logger"
)

var loggerClient *logger.Client

func main() {
    // 1. Инициализация клиента логгера
    loggerURL := os.Getenv("LOGGER_SERVICE_URL")
    if loggerURL == "" {
        loggerURL = "http://logger_service:8802"
    }
    
    loggerClient = logger.NewClient(
        loggerURL,
        "patient_service", // Имя вашего сервиса
        "",
        logger.WithAsync(3),
        logger.WithTimeout(3*time.Second),
    )
    
    // 2. Настройка корректного завершения работы
    setupGracefulShutdown(loggerClient)
    
    // 3. Логирование запуска сервиса
    loggerClient.Info("Сервис запущен", map[string]interface{}{
        "version": "1.0.0",
        "environment": os.Getenv("ENV"),
    })
    
    // 4. Настройка Echo
    e := echo.New()
    e.Use(LoggerMiddleware(loggerClient))
    
    // 5. Настройка маршрутов
    e.GET("/patients", getPatients)
    e.POST("/patients", createPatient)
    
    // 6. Запуск сервера
    port := os.Getenv("PORT")
    if port == "" {
        port = "8803"
    }
    
    if err := e.Start(":" + port); err != nil {
        loggerClient.Error("Ошибка запуска сервера", map[string]interface{}{
            "error": err.Error(),
        })
    }
}

func getPatients(c echo.Context) error {
    requestID := c.Get("request_id").(string)
    
    // Логирование бизнес-операции
    loggerClient.Debug("Получение списка пациентов", map[string]interface{}{
        "request_id": requestID,
        "component": "patient_repository",
    })
    
    // Какая-то логика получения пациентов...
    
    return c.JSON(200, map[string]interface{}{
        "message": "Success",
    })
}

func createPatient(c echo.Context) error {
    requestID := c.Get("request_id").(string)
    
    // Логирование важной бизнес-операции
    loggerClient.Info("Создание нового пациента", map[string]interface{}{
        "request_id": requestID,
        "component": "patient_service",
    })
    
    // Логика создания пациента...
    err := someOperation()
    if err != nil {
        // Логирование ошибки
        loggerClient.Error("Ошибка создания пациента", map[string]interface{}{
            "request_id": requestID,
            "error": err.Error(),
            "component": "patient_service",
        })
        
        return c.JSON(400, map[string]interface{}{
            "error": "Не удалось создать пациента",
        })
    }
    
    return c.JSON(201, map[string]interface{}{
        "message": "Patient created",
    })
}

func setupGracefulShutdown(logger *logger.Client) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        fmt.Println("Завершение работы, закрытие логгера...")
        if logger != nil {
            logger.Close()
        }
        os.Exit(0)
    }()
}

// Дополнительные функции...
```

## Просмотр логов в Kibana

После настройки и запуска системы вы можете просматривать логи в Kibana:

1. Откройте Kibana по адресу http://[ip_address]:5601
2. Перейдите в раздел "Discover"
3. Создайте индекс-паттерн "vitalem-logs-*"
4. Используйте поиск для фильтрации логов

### Полезные фильтры для поиска в Kibana:

- `service: "patient_service"` - логи конкретного сервиса
- `level: "error"` - только ошибки
- `metadata.request_id: "abc123"` - логи конкретного запроса
- `metadata.component: "database"` - логи определенного компонента
- `message: "Ошибка подключения"` - поиск по тексту сообщения

## Рекомендации по логированию

1. **Выбирайте правильный уровень логирования**:
   - `Debug` - детальная информация для отладки (большой объем)
   - `Info` - важная информация о работе приложения
   - `Warn` - предупреждения, которые не являются ошибками
   - `Error` - ошибки, требующие внимания

2. **Логируйте бизнес-события**:
   - Создание/изменение важных данных
   - Вход пользователя
   - Операции с платежами
   - Действия администраторов

3. **Структурируйте метаданные**:
   - Используйте одинаковые имена полей во всех сервисах
   - Всегда добавляйте request_id для связывания логов
   - Добавляйте контекст (component, user_id и т.д.)

4. **Не логируйте конфиденциальные данные**:
   - Пароли
   - Токены доступа
   - Полные номера карт
   - Персональные медицинские данные

5. **Ограничивайте размер логов**:
   - Не логируйте большие объемы данных (например, тела запросов)
   - Обрезайте длинные сообщения
   - Выборочно логируйте поля объектов

## Troubleshooting

### Логи не появляются в Kibana

1. Проверьте, что Logger Service запущен:
   ```bash
   docker ps | grep logger_service
   ```

2. Проверьте соединение с Logger Service:
   ```bash
   curl -X POST http://[ip_address]:8802/logs \
     -H "Content-Type: application/json" \
     -d '{"service":"test","level":"info","message":"Test log"}'
   ```

3. Проверьте, что Elasticsearch запущен:
   ```bash
   curl http://[ip_address]:9200
   ```

4. Проверьте логи самого Logger Service:
   ```bash
   docker logs logger_service
   ```

### Ошибки при отправке логов

Если вы получаете ошибки при отправке логов:

1. Проверьте URL логгера и доступность сервиса
2. Убедитесь, что структура метаданных корректна
3. Используйте синхронный режим для отладки проблем с логированием:
   ```go
   loggerClient := logger.NewClient(
       loggerURL,
       "your_service_name",
       ""
   )
   ```

## Дополнительная информация

Более подробная информация о Logger Service доступна в документации:
- [README Logger Service](./logger_service/README.md)