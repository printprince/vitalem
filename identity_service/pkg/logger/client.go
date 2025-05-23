package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Level определяет уровень логирования
type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
)

// logRequest представляет структуру запроса для отправки лога
type logRequest struct {
	Service  string                 `json:"service"`
	Level    string                 `json:"level"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Client представляет клиент для отправки логов в logger_service
type Client struct {
	baseURL     string
	serviceName string
	httpClient  *http.Client
	apiKey      string

	// Каналы для асинхронной отправки
	logChan  chan logRequest
	stopChan chan struct{}
	wg       sync.WaitGroup

	// Настройки
	async       bool
	workerCount int
}

// ClientOption функция для настройки клиента
type ClientOption func(*Client)

// WithAsync включает асинхронное логирование
func WithAsync(workerCount int) ClientOption {
	return func(c *Client) {
		c.async = true
		c.workerCount = workerCount
	}
}

// WithTimeout устанавливает таймаут для HTTP запросов
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient создает новый клиент для отправки логов
func NewClient(baseURL, serviceName, apiKey string, options ...ClientOption) *Client {
	c := &Client{
		baseURL:     baseURL,
		serviceName: serviceName,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		apiKey:      apiKey,
		logChan:     make(chan logRequest, 1000), // Буфер на 1000 логов
		stopChan:    make(chan struct{}),
		async:       false,
		workerCount: 1,
	}

	// Применяем опции
	for _, option := range options {
		option(c)
	}

	// Запускаем воркеры, если включен асинхронный режим
	if c.async {
		c.startWorkers()
	}

	return c
}

// startWorkers запускает горутины для обработки логов
func (c *Client) startWorkers() {
	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.worker()
	}
}

// worker обрабатывает логи из канала
func (c *Client) worker() {
	defer c.wg.Done()

	for {
		select {
		case req := <-c.logChan:
			// Отправляем лог синхронно внутри горутины
			if err := c.sendLog(req); err != nil {
				// В реальном приложении здесь можно добавить повторные попытки
				fmt.Printf("Error sending log: %v\n", err)
			}
		case <-c.stopChan:
			return
		}
	}
}

// sendLog отправляет лог на сервер
func (c *Client) sendLog(req logRequest) error {
	// Преобразуем запрос в JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal log request: %w", err)
	}

	// Создаем HTTP запрос
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/logs", c.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	// Отправляем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send log: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Log отправляет лог с указанным уровнем, сообщением и метаданными
func (c *Client) Log(level Level, message string, metadata map[string]interface{}) error {
	req := logRequest{
		Service:  c.serviceName,
		Level:    string(level),
		Message:  message,
		Metadata: metadata,
	}

	if c.async {
		// Асинхронная отправка
		select {
		case c.logChan <- req:
			// Лог добавлен в очередь
			return nil
		default:
			// Канал переполнен, отбрасываем лог
			return fmt.Errorf("log channel is full, log dropped")
		}
	} else {
		// Синхронная отправка
		return c.sendLog(req)
	}
}

// Debug отправляет лог с уровнем DEBUG
func (c *Client) Debug(message string, metadata map[string]interface{}) error {
	return c.Log(DebugLevel, message, metadata)
}

// Info отправляет лог с уровнем INFO
func (c *Client) Info(message string, metadata map[string]interface{}) error {
	return c.Log(InfoLevel, message, metadata)
}

// Warn отправляет лог с уровнем WARN
func (c *Client) Warn(message string, metadata map[string]interface{}) error {
	return c.Log(WarnLevel, message, metadata)
}

// Error отправляет лог с уровнем ERROR
func (c *Client) Error(message string, metadata map[string]interface{}) error {
	return c.Log(ErrorLevel, message, metadata)
}

// Close закрывает клиент и ожидает завершения всех горутин
func (c *Client) Close() {
	if c.async {
		close(c.stopChan)
		c.wg.Wait()
	}
}
