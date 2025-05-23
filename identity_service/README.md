# Identity Service

Сервис аутентификации и авторизации пользователей для медицинской платформы Vitalem.

## Обзор

Identity Service отвечает за регистрацию пользователей, аутентификацию, выдачу и проверку JWT-токенов. Сервис интегрируется с Logger Service для централизованного логирования.

## Основные функции

- Регистрация пользователей (пациентов и врачей)
- Аутентификация и выдача JWT-токенов
- Валидация токенов
- Управление пользователями

## Технологии

- Go
- Echo Framework
- PostgreSQL
- JWT

## API

### Регистрация пользователя

```
POST /register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "role": "patient"
}

Ответ:
{
  "message": "User created successfully"
}
```

### Вход пользователя

```
POST /login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

Ответ:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Валидация токена

```
GET /validate-token
Authorization: Bearer {token}

Ответ:
{
  "valid": true,
  "user_id": 1,
  "email": "user@example.com",
  "role": "patient",
  "expire": 1634567890
}
```

### Получение данных пользователя (защищенный маршрут)

```
GET /protected/user
Authorization: Bearer {token}

Ответ:
{
  "user_id": 1,
  "email": "user@example.com",
  "role": "patient"
}
```

## Интеграция с другими сервисами

Сервис интегрируется с Logger Service для централизованного логирования. Подробная информация о системе логирования доступна в [основной документации проекта](../README.md).

## Конфигурация

Конфигурация сервиса осуществляется через файл `config.yaml` и переменные окружения:

| Параметр | Переменная окружения | Описание |
|----------|----------------------|----------|
| Database Host | DB_HOST | Хост базы данных |
| Database Port | DB_PORT | Порт базы данных |
| Database User | DB_USER | Пользователь базы данных |
| Database Password | DB_PASS | Пароль базы данных |
| Database Name | DB_NAME | Имя базы данных |
| Database SSL Mode | DB_SSL_MODE | Режим SSL для базы данных |
| JWT Secret | JWT_SECRET | Секретный ключ для JWT |
| JWT Expire | JWT_EXPIRE | Время жизни JWT в часах |
| Logger Service URL | LOGGER_SERVICE_URL | URL для подключения к Logger Service |

## Разработка

### Запуск локально

```bash
go run cmd/api/main.go
```

### Сборка

```bash
go build -o identity_service ./cmd/api
```

### Docker

```bash
docker build -t identity_service .
docker run -p 8801:8801 \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASS=61324 \
  -e DB_NAME=vitalem_db \
  -e DB_SSL_MODE=disable \
  -e LOGGER_SERVICE_URL=http://logger_service:8802 \
  identity_service
```

## Структура проекта

```
identity_service/
├── cmd/
│   └── api/            # Точка входа в приложение
│       └── main.go
├── internal/
│   ├── config/         # Конфигурация
│   ├── handlers/       # Обработчики HTTP запросов
│   ├── middleware/     # Middleware
│   ├── models/         # Модели данных
│   ├── repository/     # Слой доступа к данным
│   └── service/        # Бизнес-логика
├── Dockerfile          # Инструкции для сборки Docker образа
└── config.yaml         # Файл конфигурации
```