package repository

import (
	"encoding/json"
	"time"

	"github.com/printprince/vitalem/logger_service/internal/models"

	"gorm.io/gorm"
)

// LogRepository - репо для работы с логами в базе
// Тут простая обёртка над GORM для CRUD операций с логами
// Если надумаем усложнять - можно добавить кеширование или батчинг
type LogRepository struct {
	db *gorm.DB
}

// NewLogRepository - фабрика для создания инстанса репозитория
// Просто прокидываем коннекшн к БД внутрь структуры
func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db: db}
}

// CreateLog - сохраняет лог в базу данных
// Превращает метаданные из мапы в JSON и записывает их в БД
// Не очень эффективно из-за сериализации, но зато гибко
// TODO: Добавить батчинг для массовой вставки логов
func (r *LogRepository) CreateLog(service string, level models.LogLevel, message string, metadata map[string]interface{}) error {
	// Сериализуем метаданные в JSON строку
	// Если сериализация фейлится - вся операция отменяется
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// Создаем объект лога со всеми полями
	logEntry := models.LogEntry{
		Timestamp: time.Now(),           // Текущее время с миллисекундами
		Service:   service,              // Имя сервиса откуда пришёл лог
		Level:     level,                // Уровень логирования (info/error/etc)
		Message:   message,              // Основной текст лога
		Metadata:  string(metadataJSON), // Сериализованные дополнительные данные
		CreatedAt: time.Now(),           // Метка создания для GORM
	}

	// Пишем в базу через GORM
	// Тут намеренно нет транзакции, т.к. операция атомарная
	// и производительность важнее консистентности в логировании
	return r.db.Create(&logEntry).Error
}

// GetLogs - получает логи из базы с применением фильтров
// Принимает параметры для фильтрации и пагинации
// Результаты сортируются от новых к старым (DESC)
// Поддерживает фильтрацию по сервису, уровню, временному диапазону
func (r *LogRepository) GetLogs(service string, level models.LogLevel, startTime, endTime time.Time, limit, offset int) ([]models.LogEntry, error) {
	var logs []models.LogEntry
	query := r.db.Model(&models.LogEntry{})

	// Билдим динамический запрос с нужными фильтрами
	// Если параметр не задан (пустой), то фильтр не применяется
	if service != "" {
		query = query.Where("service = ?", service)
	}

	if level != "" {
		query = query.Where("level = ?", level)
	}

	if !startTime.IsZero() {
		query = query.Where("timestamp >= ?", startTime)
	}

	if !endTime.IsZero() {
		query = query.Where("timestamp <= ?", endTime)
	}

	// Добавляем пагинацию и сортировку по времени (свежие в начале)
	// Жёстко задаём DESC сортировку, т.к. обычно нужны самые свежие логи
	query = query.Limit(limit).Offset(offset).Order("timestamp DESC")

	// Выполняем запрос и собираем результаты
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}
