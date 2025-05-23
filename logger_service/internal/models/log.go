package models

import "time"

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug" // Debug level for detailed information
	LogLevelInfo  LogLevel = "info"  // Info level for general information
	LogLevelWarn  LogLevel = "warn"  // Warn level for potential issues
	LogLevelError LogLevel = "error" // Error level for errors
)

// Представляет запись в логе
type LogEntry struct {
	ID        uint   `gorm:"primary_key" json:"id"` // Идентификатор записи
	Timestamp time.Time `json:"timestamp"` // Время создания записи
	Service   string    `json:"service"` // Имя сервиса, который записал лог
	Level     LogLevel   `json:"level"` // Уровень логирования
	Message   string    `json:"message"` // Сообщение лога
	Metadata  string    `json:"metadata"` // Дополнительные данные
	CreatedAt time.Time `json:"created_at"` // Время создания записи
}