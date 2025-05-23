package service

import (
	"fmt"
	"time"

	"github.com/printprince/vitalem/logger_service/internal/models"
)

type ElasticsearchClient interface {
	IndexLog(logData map[string]interface{}) error
}

type LogService struct {
	esClient ElasticsearchClient
}

func NewLogService() *LogService {
	return &LogService{}
}

func (s *LogService) SetElasticsearchClient(client ElasticsearchClient) {
	s.esClient = client
}

// Log создает новую запись лога
func (s *LogService) Log(service string, level models.LogLevel, message string, metadata map[string]interface{}) error {
	// Проверяем что Elasticsearch клиент настроен
	if s.esClient == nil {
		return fmt.Errorf("elasticsearch client is not configured")
	}

	logData := map[string]interface{}{
		"service":    service,
		"level":      string(level),
		"message":    message,
		"@timestamp": time.Now().Format(time.RFC3339),
	}

	if metadata != nil {
		for k, v := range metadata {
			logData[k] = v
		}
	}

	return s.esClient.IndexLog(logData)
}

// GetLogs получает логи с фильтрацией
func (s *LogService) GetLogs(service string, level models.LogLevel, startTime, endTime time.Time, limit, offset int) ([]models.LogEntry, error) {
	return nil, fmt.Errorf("method not supported: use Kibana for log retry")
}
