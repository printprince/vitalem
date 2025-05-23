package repository

import (
	"encoding/json"
	"logger_service/internal/models"
	"time"

	"gorm.io/gorm"
)

type LogRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db: db}
}

// CreateLog сохраняет запись лога в базу данных
func (r *LogRepository) CreateLog(service string, level models.LogLevel, message string, metadata map[string]interface{}) error {
	// Преобразуем метаданные в JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// Создаем запись лога
	logEntry := models.LogEntry{
		Timestamp: time.Now(),
		Service:   service,
		Level:     level,
		Message:   message,
		Metadata:  string(metadataJSON),
		CreatedAt: time.Now(),
	}

	// Сохраняем в базу данных
	return r.db.Create(&logEntry).Error
}

func (r *LogRepository) GetLogs(service string, level models.LogLevel, startTime, endTime time.Time, limit, offset int) ([]models.LogEntry, error) {
	var logs []models.LogEntry
	query := r.db.Model(&models.LogEntry{})

	// применяем фильтры, если они заданы
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

	// Пагинация, с конца читаем
	query = query.Limit(limit).Offset(offset).Order("timestamp DESC")

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}
