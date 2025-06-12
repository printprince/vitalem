# File Server Service

Современный файловый сервис, построенный на Go, предоставляющий безопасное хранение и управление файлами. Сервис использует PostgreSQL для хранения метаданных и MinIO для хранения объектов, реализуя чистую архитектуру.

## Возможности

- Безопасная загрузка и скачивание файлов
- Аутентификация на основе JWT
- Интеграция с PostgreSQL через GORM
- Интеграция с MinIO для хранения объектов
- Структурированное логирование с использованием Zap
- Реализация чистой архитектуры
- Маршрутизация HTTP с использованием Chi
- Конфигурация через переменные окружения
- Поддержка миграций базы данных

## Технологический стек

- Go 1.24
- PostgreSQL (с драйвером pgx)
- MinIO для хранения объектов
- GORM для работы с базой данных
- Chi для маршрутизации HTTP
- JWT для аутентификации
- Zap для логирования
- Godotenv для конфигурации

## Структура проекта

```
.
├── cmd/
│   └── fileserver/    # Точка входа приложения
├── internal/
│   ├── config/        # Управление конфигурацией
│   ├── http/          # HTTP слой
│   │   ├── handler/   # HTTP обработчики
│   │   ├── middleware/# HTTP промежуточное ПО
│   │   ├── router/    # Определение маршрутов
│   │   └── logger/    # Конфигурация логирования
│   ├── model/         # Доменные модели
│   ├── repository/    # Слой доступа к данным
│   ├── service/       # Бизнес-логика
│   └── storage/       # Реализации хранилищ
├── migrations/        # Миграции базы данных
├── .env              # Переменные окружения (опционально)
└── README.md         # Этот файл
```

## Требования

- Go 1.24 или выше
- PostgreSQL 15 или выше
- MinIO Server
- Docker (опционально)

## Переменные окружения

Сервис можно настроить с помощью переменных окружения или файла `.env`:

### Конфигурация сервера
- `APP_ENV` - Окружение приложения (по умолчанию: "development")
- `SERVER_PORT` - Порт сервера (по умолчанию: "8080")
- `JWT_SECRET` - Секретный ключ для генерации JWT токенов

### Конфигурация базы данных
- `DB_HOST` - Хост PostgreSQL (по умолчанию: "localhost")
- `DB_PORT` - Порт PostgreSQL (по умолчанию: "5432")
- `DB_USER` - Пользователь базы данных (по умолчанию: "postgres")
- `DB_PASSWORD` - Пароль базы данных (по умолчанию: "12345")
- `DB_NAME` - Имя базы данных (по умолчанию: "fileserver_db")

### Конфигурация MinIO
- `MINIO_ENDPOINT` - Endpoint MinIO (по умолчанию: "localhost:9000")
- `MINIO_ACCESS_KEY` - Access Key MinIO (по умолчанию: "minioadmin")
- `MINIO_SECRET_KEY` - Secret Key MinIO (по умолчанию: "minioadmin")
- `MINIO_USE_SSL` - Использовать SSL для MinIO (по умолчанию: false)

## Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/yourusername/FileServerService.git
cd FileServerService
```

2. Установите зависимости:
```bash
go mod download
```

3. Создайте файл `.env` в корневой директории:
```env
APP_ENV=development
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=fileserver_db
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=your_access_key
MINIO_SECRET_KEY=your_secret_key
JWT_SECRET=your_jwt_secret
```

4. Запустите миграции базы данных:
```bash
# Добавьте команды миграции здесь
```

5. Запустите сервис:
```bash
go run cmd/fileserver/main.go
```

## Разработка

### Запуск тестов
```bash
go test ./...
```

### Стиль кода
Проект следует стандартным рекомендациям по стилю кода Go. Используйте `gofmt` для форматирования кода:
```bash
gofmt -w .
```

### Миграции базы данных
Миграции базы данных хранятся в директории `migrations`. Используйте соответствующий инструмент для управления изменениями схемы базы данных.

## Документация API

### Эндпоинты

#### Публичные маршруты

- `GET /health` - Проверка состояния сервиса
- `GET /public/{id}` - Публичный доступ к файлам (без авторизации)

#### Защищенные маршруты (требуют JWT авторизации)

Все защищенные маршруты требуют JWT токен в заголовке `Authorization: Bearer <token>`

##### Управление файлами

- `POST /files` - Загрузка нового файла
  - Content-Type: multipart/form-data
  - Параметры:
    - file: файл для загрузки
    - name: (опционально) имя файла
    - description: (опционально) описание файла

- `GET /files` - Получение списка всех файлов пользователя
  - Query параметры:
    - page: номер страницы (по умолчанию 1)
    - limit: количество файлов на странице (по умолчанию 10)

- `GET /files/{id}` - Получение метаинформации о файле
  - Параметры:
    - id: идентификатор файла

- `DELETE /files/{id}` - Удаление файла
  - Параметры:
    - id: идентификатор файла

- `GET /files/{id}/download` - Скачивание файла
  - Параметры:
    - id: идентификатор файла

- `GET /files/{id}/preview` - Предпросмотр файла
  - Параметры:
    - id: идентификатор файла

- `PATCH /files/{id}/visibility` - Изменение публичности файла
  - Параметры:
    - id: идентификатор файла
  - Body:
    ```json
    {
      "is_public": "true/false"
    }
    ```

### Ответы API

#### Успешные ответы

- `200 OK` - Успешное выполнение запроса
- `201 Created` - Успешное создание ресурса
- `204 No Content` - Успешное выполнение запроса без возвращаемых данных

#### Ошибки

- `400 Bad Request` - Неверный формат запроса
- `401 Unauthorized` - Отсутствует или неверный токен авторизации
- `403 Forbidden` - Нет прав доступа к ресурсу
- `404 Not Found` - Ресурс не найден
- `500 Internal Server Error` - Внутренняя ошибка сервера

### Примеры запросов

#### Загрузка файла
```bash
curl -X POST http://localhost:8080/files \
  -H "Authorization: Bearer <your-jwt-token>" \
  -F "file=@/path/to/file.txt" \
  -F "name=My File" \
  -F "description=My file description"
```

#### Получение списка файлов
```bash
curl -X GET http://localhost:8080/files \
  -H "Authorization: Bearer <your-jwt-token>"
```

#### Скачивание файла
```bash
curl -X GET http://localhost:8080/files/{id}/download \
  -H "Authorization: Bearer <your-jwt-token>" \
  --output downloaded_file.txt
```

## Участие в разработке

1. Форкните репозиторий
2. Создайте ветку для вашей функции (`git checkout -b feature/amazing-feature`)
3. Зафиксируйте ваши изменения (`git commit -m 'Добавлена новая функция'`)
4. Отправьте изменения в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## Лицензия

Этот проект распространяется под лицензией MIT - подробности в файле LICENSE.

## Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/yourusername/FileServerService.git
cd FileServerService
```

2. Запустите сервисы с помощью Docker Compose:
```bash
docker-compose up -d
```

Сервис будет доступен по адресу: http://localhost:8080

## API Endpoints

### Файлы

- `POST /api/v1/files` - Загрузка файла
- `GET /api/v1/files/{fileID}` - Скачивание файла
- `DELETE /api/v1/files/{fileID}` - Удаление файла
- `GET /api/v1/files` - Список файлов

### Здоровье

- `GET /api/v1/health` - Проверка состояния сервиса

## Разработка

1. Установите зависимости:
```bash
go mod download
```

2. Запустите тесты:
```bash
go test ./...
```

3. Запустите сервис локально:
```bash
go run cmd/fileserver/main.go
```

# Docker инструкции

## Требования

- Docker
- Docker Compose

## Структура

```
build/
├── Dockerfile        # Конфигурация сборки приложения
├── docker-compose.yml # Конфигурация всех сервисов
└── README.md         # Этот файл
```

## Запуск

1. Сборка и запуск всех сервисов:
```bash
docker-compose -f build/docker-compose.yml up -d
```

2. Просмотр логов:
```bash
# Все сервисы
docker-compose -f build/docker-compose.yml logs -f

# Только приложение
docker-compose -f build/docker-compose.yml logs -f app
```

3. Остановка всех сервисов:
```bash
docker-compose -f build/docker-compose.yml down
```

## Доступ к сервисам

- Приложение: http://localhost:8080
- MinIO Console: http://localhost:9001
- MinIO API: http://localhost:9000
- PostgreSQL: localhost:5432

## Переменные окружения

Все переменные окружения можно изменить в файле `docker-compose.yml`. Основные настройки:

### Приложение
- `APP_ENV` - Окружение приложения
- `SERVER_PORT` - Порт сервера
- `JWT_SECRET` - Секретный ключ для JWT

### База данных
- `POSTGRES_USER` - Пользователь PostgreSQL
- `POSTGRES_PASSWORD` - Пароль PostgreSQL
- `POSTGRES_DB` - Имя базы данных

### MinIO
- `MINIO_ROOT_USER` - Пользователь MinIO
- `MINIO_ROOT_PASSWORD` - Пароль MinIO

## Тома данных

- `postgres_data` - Данные PostgreSQL
- `minio_data` - Данные MinIO

## Сеть

Все сервисы подключены к сети `fileserver_network`, что обеспечивает их взаимодействие между собой.

## Безопасность

1. В продакшене обязательно измените все пароли и секретные ключи
2. Настройте SSL для MinIO
3. Ограничьте доступ к портам в продакшене
4. Используйте секреты Docker для хранения чувствительных данных

## Миграции

Миграции базы данных автоматически применяются при первом запуске контейнера PostgreSQL из директории `migrations`. 