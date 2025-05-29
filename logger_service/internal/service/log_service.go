package service

import (
	"fmt"
	"time"

	"github.com/printprince/vitalem/logger_service/internal/models"
)

// ElasticsearchClient - интерфейс клиента для Elasticsearch
// Абстракция для возможности подмены реализации (например, для тестирования)
type ElasticsearchClient interface {
	IndexLog(logData map[string]interface{}) error
}

// LogService - основной сервис для работы с логами
// Содержит бизнес-логику обработки и отправки логов в Elasticsearch
type LogService struct {
	esClient ElasticsearchClient
}

// NewLogService - фабрика для создания экземпляра сервиса логирования
// По умолчанию esClient не инициализирован, нужно вызвать SetElasticsearchClient
func NewLogService() *LogService {
	return &LogService{}
}

// SetElasticsearchClient - сеттер для Elasticsearch клиента
// Позволяет отложенную инициализацию клиента после создания сервиса
func (s *LogService) SetElasticsearchClient(client ElasticsearchClient) {
	s.esClient = client
}

// Log - основной метод для записи лога в Elasticsearch
// Принимает сервис-источник, уровень, сообщение и доп. метаданные
// Транформирует все это в формат для индексации в Elasticsearch
func (s *LogService) Log(service string, level models.LogLevel, message string, metadata map[string]interface{}) error {
	// Проверка наличия ES клиента - фейлим сразу, если не настроен
	// Без этой проверки можно словить NPE при высокой нагрузке
	if s.esClient == nil {
		return fmt.Errorf("elasticsearch client is not configured")
	}

	// Формируем структуру лога для Elasticsearch
	// Объединяем базовые поля и все переданные метаданные
	logData := map[string]interface{}{
		"service":    service,                         // Имя сервиса-отправителя
		"level":      string(level),                   // Уровень важности (info/warn/error/debug)
		"message":    message,                         // Основной текст сообщения
		"@timestamp": time.Now().Format(time.RFC3339), // Стандартный формат времени для ELK
	}

	// Мержим метаданные, если они есть
	// Каждое поле метаданных становится отдельным полем в документе ES
	if metadata != nil {
		for k, v := range metadata {
			logData[k] = v
		}
	}

	// Отправляем лог в Elasticsearch через клиент
	// При ошибке пробрасываем её вызывающему коду
	return s.esClient.IndexLog(logData)
}

// GetLogs - заглушка для получения логов
// В реальности мы используем Kibana для поиска по Elasticsearch
// Метод оставлен для совместимости с интерфейсом, но не реализован
func (s *LogService) GetLogs(service string, level models.LogLevel, startTime, endTime time.Time, limit, offset int) ([]models.LogEntry, error) {
	// TODO: Реализовать прямой поиск по ES, если потребуется API
	// Сейчас предпочтительнее использовать Kibana для анализа логов
	return nil, fmt.Errorf("method not supported: use Kibana for log retrieval")
}
