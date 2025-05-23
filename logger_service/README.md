# Logger Service

Централизованный сервис логирования для всех микросервисов с интеграцией с ELK-стеком.

## Обзор

Logger Service предоставляет единый интерфейс для сбора, хранения и анализа логов со всех микросервисов платформы Vitalem. Сервис интегрируется с Elasticsearch для хранения логов и Kibana для их визуализации.

## Основные функции

- Сбор логов от всех микросервисов
- Отправка логов в Elasticsearch
- Асинхронная обработка логов
- Различные уровни логирования (debug, info, warn, error)
- API для отправки логов
- Клиентская библиотека для интеграции с другими сервисами

## Технологии

- Go
- Echo Framework
- Elasticsearch
- Kibana
- Filebeat

## API

### Отправка лога

```
POST /logs
Content-Type: application/json

{
  "service": "service_name",
  "level": "info",
  "message": "Log message",
  "metadata": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

### Защищенные маршруты

```
GET /protected/logs/stats
Authorization: Bearer {jwt_token}

Ответ:
{
  "status": "Authenticated access"
}
```

## Использование Logger Service как библиотеки

Logger Service можно использовать как библиотеку для централизованного логирования в любом Go-проекте.

### Установка

```bash
go get github.com/printprince/vitalem/logger_service
```

### Инициализация клиента логгера

```go
import (
    "github.com/printprince/vitalem/logger_service/pkg/logger"
    "time"
    "os"
    "os/signal"
    "syscall"
    "fmt"
)

// Создание клиента с синхронной отправкой логов
loggerClient := logger.NewClient(
    "http://logger_service:8802", // URL сервиса логирования
    "your_service_name",          // Имя вашего сервиса
    "",                           // API ключ (опционально)
)

// Создание клиента с асинхронной отправкой логов (рекомендуется)
loggerClient := logger.NewClient(
    "http://logger_service:8802",
    "your_service_name",
    "",
    logger.WithAsync(3),             // Запуск 3 воркеров для асинхронной отправки
    logger.WithTimeout(3*time.Second), // Таймаут HTTP запросов
)
```

### Отправка логов разных уровней

```go
// Отправка информационного лога
loggerClient.Info("Сервер запущен", map[string]interface{}{
    "host": "localhost",
    "port": 8801,
})

// Отправка отладочного лога
loggerClient.Debug("Детальная информация", map[string]interface{}{
    "request_id": "abc123",
    "user_id": 42,
})

// Отправка предупреждения
loggerClient.Warn("Внимание", map[string]interface{}{
    "resource": "database",
    "latency_ms": 500,
})

// Отправка ошибки
loggerClient.Error("Ошибка подключения", map[string]interface{}{
    "error": err.Error(),
    "component": "database",
})
```

### Корректное завершение работы

Для асинхронного клиента важно корректно завершить работу, чтобы все логи были отправлены:

```go
// Настройка корректного завершения работы
func setupGracefulShutdown(logger *logger.Client) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        fmt.Println("Завершение работы, закрытие логгера...")
        if logger != nil {
            logger.Close() // Дожидается отправки всех логов
        }
        os.Exit(0)
    }()
}
```

### Рекомендации по использованию

1. **Используйте асинхронную отправку** - это не блокирует основной поток выполнения.
2. **Добавляйте информативные метаданные** - это поможет при поиске и фильтрации логов.
3. **Используйте правильные уровни логирования**:
   - `Debug` - для отладочной информации (большой объем)
   - `Info` - для информационных сообщений о работе приложения
   - `Warn` - для предупреждений, которые не являются ошибками
   - `Error` - для ошибок, требующих внимания

4. **Всегда закрывайте клиент** - используйте `defer loggerClient.Close()` или настройте корректное завершение работы.
5. **Обрабатывайте ошибки отправки логов** - методы клиента возвращают ошибку, если лог не удалось отправить.

### Пример полной интеграции

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/printprince/vitalem/logger_service/pkg/logger"
)

var loggerClient *logger.Client

func main() {
    // Инициализация клиента логгера
    loggerClient = logger.NewClient(
        "http://logger_service:8802",
        "example_service",
        "",
        logger.WithAsync(3),
        logger.WithTimeout(3*time.Second),
    )
    
    // Настройка корректного завершения работы
    setupGracefulShutdown(loggerClient)
    
    // Отправка тестового лога
    if err := loggerClient.Info("Сервис запущен", map[string]interface{}{
        "version": "1.0.0",
        "environment": "development",
    }); err != nil {
        fmt.Printf("Ошибка отправки лога: %v\n", err)
    }
    
    // Основная логика приложения
    // ...
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

// Пример функции с логированием
func processRequest(requestID string, userID int) error {
    // Логирование начала обработки
    loggerClient.Debug("Начало обработки запроса", map[string]interface{}{
        "request_id": requestID,
        "user_id": userID,
    })
    
    // Какая-то логика
    // ...
    
    // Логирование успешного завершения
    loggerClient.Info("Запрос обработан успешно", map[string]interface{}{
        "request_id": requestID,
        "user_id": userID,
        "duration_ms": 150,
    })
    
    return nil
}

// Пример функции с логированием ошибки
func connectToDatabase() error {
    // Попытка подключения
    err := someDbConnection()
    
    if err != nil {
        // Логирование ошибки
        loggerClient.Error("Ошибка подключения к базе данных", map[string]interface{}{
            "error": err.Error(),
            "retry_count": 3,
        })
        return err
    }
    
    return nil
}
```

## Конфигурация

Конфигурация сервиса осуществляется через файл `config.yaml` и переменные окружения:

| Параметр | Переменная окружения | Описание |
|----------|----------------------|----------|
| Elasticsearch URL | ELASTICSEARCH_URL | URL для подключения к Elasticsearch |
| JWT Secret | - | Секретный ключ для проверки JWT токенов |
| Server Port | - | Порт, на котором запускается сервис |
| Logging Level | - | Уровень логирования (debug, info, warn, error) |

## Разработка

### Запуск локально

```bash
go run cmd/api/main.go
```

### Сборка

```bash
go build -o logger_service ./cmd/api
```

### Docker

```bash
docker build -t logger_service .
docker run -p 8802:8802 -e ELASTICSEARCH_URL=http://elasticsearch:9200 logger_service
```

## Структура проекта

```
logger_service/
├── cmd/
│   └── api/                  # Точка входа в приложение
│       └── main.go
├── internal/
│   ├── config/               # Конфигурация
│   ├── handlers/             # Обработчики HTTP запросов
│   ├── middleware/           # Middleware
│   ├── models/               # Модели данных
│   ├── repository/           # Слой доступа к данным (устаревший)
│   └── service/              # Бизнес-логика
├── pkg/
│   ├── elasticsearch/        # Клиент для работы с Elasticsearch
│   └── logger/               # Клиентская библиотека для других сервисов
├── Dockerfile                # Инструкции для сборки Docker образа
└── config.yaml               # Файл конфигурации
```

## Интеграция с ELK

Logger Service интегрируется со стеком ELK (Elasticsearch, Logstash, Kibana) для хранения и анализа логов:

1. **Elasticsearch** - хранит все логи в индексах с датой
2. **Kibana** - предоставляет веб-интерфейс для поиска и визуализации логов
3. **Filebeat** - собирает логи из контейнеров Docker

### Просмотр логов в Kibana

1. Откройте Kibana по адресу http://localhost:5601
2. Перейдите в раздел "Discover"
3. Создайте индекс-паттерн "vitalem-logs-*"
4. Используйте поиск для фильтрации логов по сервису, уровню и другим полям
```