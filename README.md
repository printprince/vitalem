# Vitalem Microservices

Микросервисная архитектура для медицинской платформы Vitalem.

## Сервисы

### Identity Service

Сервис аутентификации и авторизации пользователей. Отвечает за регистрацию, вход и управление JWT-токенами.

**Основные функции:**
- Регистрация пользователей (пациентов и врачей)
- Аутентификация и выдача JWT-токенов
- Валидация токенов
- Управление пользователями

**Технологии:**
- Go
- Echo Framework
- PostgreSQL
- JWT

### Logger Service

Централизованный сервис логирования для всех микросервисов с интеграцией с ELK-стеком.

**Основные функции:**
- Сбор логов от всех микросервисов
- Отправка логов в Elasticsearch
- Асинхронная обработка логов
- Различные уровни логирования (debug, info, warn, error)

**Технологии:**
- Go
- Echo Framework
- Elasticsearch
- Kibana
- Filebeat

## Запуск проекта

### Предварительные требования

- Docker и Docker Compose
- Go 1.21+

### Запуск с помощью Docker Compose

```bash
docker-compose up -d
```

### Доступные сервисы

- Identity Service: http://localhost:8801
- Logger Service: http://localhost:8802
- Kibana: http://localhost:5601
- Elasticsearch: http://localhost:9200

## API

### Identity Service

#### Регистрация пользователя

```
POST /register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "role": "patient"
}
```

#### Вход пользователя

```
POST /login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

#### Валидация токена

```
GET /validate-token
Authorization: Bearer {token}
```

### Logger Service

#### Отправка лога

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

## Архитектура

Проект построен на микросервисной архитектуре, где каждый сервис отвечает за свою область ответственности и может быть разработан, развернут и масштабирован независимо от других.

Коммуникация между сервисами осуществляется через HTTP API, а в будущем планируется внедрение RabbitMQ для асинхронного взаимодействия.

## Разработка

### Структура проекта

```
vitalem-microservices/
├── docker-compose.yml
├── identity_service/
│   ├── cmd/
│   ├── internal/
│   ├── pkg/
│   ├── Dockerfile
│   └── config.yaml
├── logger_service/
│   ├── cmd/
│   ├── internal/
│   ├── pkg/
│   ├── Dockerfile
│   └── config.yaml
└── filebeat/
    └── filebeat.yml
```

### Добавление нового сервиса

1. Создайте новую директорию для сервиса
2. Реализуйте необходимую логику
3. Добавьте Dockerfile
4. Обновите docker-compose.yml
5. Интегрируйте с Logger Service для логирования 