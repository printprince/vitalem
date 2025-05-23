package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ESClient struct {
	baseURL string
	index   string
	client  *http.Client
}

func NewESClient(baseURL, index string) *ESClient {
	return &ESClient{
		baseURL: baseURL,
		index:   index,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *ESClient) IndexLog(logData map[string]interface{}) error {
	// Добавляем timestamp, если его нет
	if _, ok := logData["@timestamp"]; !ok {
		logData["@timestamp"] = time.Now().Format(time.RFC3339)
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		return fmt.Errorf("failed to marshal log data: %w", err)
	}

	// Формируем URL для индекса с датой
	indexName := fmt.Sprintf("%s-%s", c.index, time.Now().Format("2006.01.02"))
	url := fmt.Sprintf("%s/%s/_doc", c.baseURL, indexName)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("elasticsearch returned error status: %d", resp.StatusCode)
	}

	return nil
}

// Ping - метод для проверки подключения к es
func (c *ESClient) Ping() error {
	req, err := http.NewRequest("GET", c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("elasticsearch returned error status: %d", resp.StatusCode)
	}

	return nil
}
