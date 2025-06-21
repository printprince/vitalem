# Интеграция FileServerService с Vitalem

## Обзор

FileServerService был успешно интегрирован в экосистему Vitalem и теперь является частью общей микросервисной архитектуры.

## Выполненные изменения

### 1. Унификация зависимостей

- **Go модуль**: Изменен с `github.com/vitalem/fileserver` на `github.com/printprince/vitalem/FileServerService`
- **JWT библиотека**: Переход с `github.com/golang-jwt/jwt/v5` на `github.com/dgrijalva/jwt-go` для совместимости
- **HTTP framework**: Переход с Chi на Echo для унификации с остальными сервисами
- **Middleware**: Интеграция с общим JWT middleware из `utils/middleware`

### 2. Обновление модели данных

Модель `File` обновлена для соответствия схеме БД:
```go
type File struct {
    ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
    CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
    
    Name         string `json:"name"`          // Отображаемое имя
    OriginalName string `json:"original_name"` // Оригинальное имя файла
    MimeType     string `json:"mime_type"`     // MIME тип
    Size         int64  `json:"size"`          // Размер в байтах
    Bucket       string `json:"bucket"`        // MinIO bucket
    Path         string `json:"path"`          // Путь в хранилище
    
    UserID       uuid.UUID `json:"user_id"`     // ID владельца
    IsPublic     bool      `json:"is_public"`   // Публичный доступ
}
```

### 3. API Endpoints

#### Защищенные маршруты (требуют JWT):
- `POST /files` - Загрузка файла
- `GET /files` - Список файлов пользователя
- `GET /files/:id` - Метаинформация о файле
- `GET /files/:id/download` - Скачивание файла
- `GET /files/:id/preview` - Предпросмотр файла
- `DELETE /files/:id` - Удаление файла
- `PATCH /files/:id/visibility` - Изменение публичности

#### Публичные маршруты:
- `GET /health` - Статус сервиса
- `GET /public/:id` - Публичное скачивание файла

### 4. Инфраструктура

#### Docker Configuration
```yaml
fileserver_service:
  build:
    context: .
    dockerfile: ./FileServerService/Dockerfile
  container_name: vitalem_fileserver
  environment:
    - SERVER_PORT=8803
    - DB_HOST=postgres
    - DB_NAME=vitalem_db
    - MINIO_ENDPOINT=minio:9000
    - JWT_SECRET=Hsb762HnbHGUAD
  ports:
    - "8803:8803"
  depends_on:
    - postgres
    - minio
    - logger_service
```

#### MinIO Configuration
- **Bucket**: `vitalem-files`
- **Endpoint**: `minio:9000` (внутри Docker сети)
- **Public endpoint**: `localhost:9000` (для разработки)
- **Management UI**: `localhost:9001` (admin: minioadmin/minioadmin)

### 5. База данных

#### Миграция
```sql
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    bucket VARCHAR(100) NOT NULL,
    path VARCHAR(255) NOT NULL,
    
    user_id UUID NOT NULL,
    is_public BOOLEAN DEFAULT FALSE
);
```

#### Индексы
- `idx_files_user_id` - для быстрого поиска файлов пользователя
- `idx_files_bucket` - для операций с bucket
- `idx_files_created_at` - для сортировки по дате
- `idx_files_deleted_at` - для soft delete
- `idx_files_is_public` - для публичных файлов

## Интеграция с другими сервисами

### Identity Service
- **JWT токены**: Использует общий секрет для валидации токенов
- **User ID**: Извлекается из JWT payload в формате UUID
- **Авторизация**: Проверка владельца файла перед операциями

### Logger Service
- **Логирование**: Интеграция с централизованной системой логирования
- **Elasticsearch**: Логи отправляются в общий индекс `vitalem-logs`
- **Структурированные логи**: Используется Zap logger

### Notification Service
В будущем может интегрироваться для:
- Уведомления о загрузке файлов
- Уведомления о превышении квот
- Уведомления о доступе к файлам

## Безопасность

### Аутентификация
- JWT токены с проверкой подписи
- Валидация срока действия токенов
- Извлечение user_id из токена

### Авторизация
- Проверка владельца файла
- Публичные файлы доступны без авторизации
- Приватные файлы доступны только владельцу

### Файловая безопасность
- Валидация размера файлов (до 50MB)
- Проверка MIME типов
- Безопасные имена файлов в хранилище

## Мониторинг

### Health Checks
- `GET /health` - проверка состояния сервиса
- Docker healthcheck каждые 30 секунд
- Проверка подключения к БД и MinIO

### Метрики
- Размер загруженных файлов
- Количество операций с файлами
- Время отклика API
- Использование дискового пространства

## Запуск и развертывание

### Разработка
```bash
# Запуск всей системы
docker-compose up -d

# Только FileServer и зависимости
docker-compose up -d postgres minio fileserver_service
```

### Тестирование API
```bash
# Получение JWT токена от Identity Service
curl -X POST http://localhost:8801/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Загрузка файла
curl -X POST http://localhost:8803/files \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@example.pdf"

# Список файлов
curl -X GET http://localhost:8803/files \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Будущие улучшения

1. **Квоты пользователей** - ограничение дискового пространства
2. **Thumbnail генерация** - автоматическое создание превью для изображений
3. **Версионирование файлов** - сохранение истории изменений
4. **Файловые теги** - система тегов и категорий
5. **Поиск по содержимому** - интеграция с Elasticsearch
6. **CDN интеграция** - для быстрой доставки файлов
7. **Антивирусная проверка** - сканирование загружаемых файлов

## Конфигурация Environment

```bash
# FileServer
APP_ENV=development
SERVER_PORT=8803
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=61324
DB_NAME=vitalem_db

# MinIO
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false

# Security
JWT_SECRET=Hsb762HnbHGUAD
```

## Troubleshooting

### Частые проблемы

1. **Ошибка подключения к MinIO**
   - Проверьте доступность minio:9000
   - Проверьте credentials

2. **JWT токен недействителен**
   - Убедитесь что JWT_SECRET одинаковый во всех сервисах
   - Проверьте формат токена

3. **Файл не найден**
   - Проверьте существование файла в MinIO
   - Проверьте правильность path в БД

4. **Превышен размер файла**
   - Максимальный размер: 50MB
   - Настраивается в коде handler'а 