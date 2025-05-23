# Vitalem Microservices

Микросервисная архитектура для медицинской платформы Vitalem.

## Обзор проекта

Vitalem - это медицинская платформа, построенная на микросервисной архитектуре. Проект разделен на независимые сервисы, каждый из которых отвечает за свою область функциональности.

### Сервисы

- **Identity Service** - Аутентификация и авторизация пользователей
- **Logger Service** - Централизованное логирование для всех микросервисов

## Архитектура

Проект построен на микросервисной архитектуре с использованием следующих принципов:

- **Независимость сервисов** - Каждый сервис может быть разработан, развернут и масштабирован независимо
- **Единая точка логирования** - Все сервисы отправляют логи в централизованный Logger Service
- **Контейнеризация** - Все сервисы запускаются в Docker-контейнерах
- **Интеграция с ELK** - Для анализа и визуализации логов используется стек Elasticsearch, Kibana, Filebeat

### Взаимодействие между сервисами

```
┌─────────────────┐     HTTP     ┌─────────────────┐
│  Identity       │───────────-->│  Logger         │
│  Service        │   (логи)     │  Service        │
└─────────────────┘              └─────────────────┘
        │                                │
        │ HTTP                           │ HTTP
        │ (аутентификация)               │ (логи)
        ▼                                ▼
┌─────────────────┐              ┌─────────────────┐
│  Клиенты        │              │  Elasticsearch  │
│  (фронтенд)     │              │  (хранение)     │
└─────────────────┘              └─────────────────┘
                                         │
                                         │ HTTP
                                         ▼
                                  ┌─────────────────┐
                                  │  Kibana         │
                                  │  (визуализация) │
                                  └─────────────────┘
```

## Технологический стек

- **Языки программирования**: Go
- **Фреймворки**: Echo
- **Базы данных**: PostgreSQL
- **Логирование**: ELK Stack (Elasticsearch, Kibana, Filebeat)
- **Контейнеризация**: Docker, Docker Compose

## Система логирования

Проект использует централизованную систему логирования на базе ELK-стека (Elasticsearch, Kibana, Filebeat) и собственного Logger Service. Логи всех сервисов отправляются в единое хранилище и доступны для просмотра через Kibana.

Подробная информация о настройке и использовании системы логирования доступна в [документации Logger Service](./logger_service/README.md).

## Запуск проекта

### Предварительные требования

- Docker и Docker Compose
- Go 1.21+

### Запуск с помощью Docker Compose

```bash
docker-compose up -d
```

## Разработка

Каждый сервис содержит свою собственную документацию в соответствующем README.md файле с подробными инструкциями по разработке и использованию.

## Лицензия

Vitalem

