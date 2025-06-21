package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"NotificationService/internal/config"
)

type Sender interface {
	Send(message string) error
	SendMarkdown(message string) error
}

type TelegramSender struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegramSender(cfg *config.TelegramConfig) Sender {
	return &TelegramSender{
		botToken: cfg.BotToken,
		chatID:   cfg.ChatID,
		client:   &http.Client{},
	}
}

func (t *TelegramSender) Send(message string) error {
	return t.sendMessage(message, "")
}

func (t *TelegramSender) SendMarkdown(message string) error {
	return t.sendMessage(message, "MarkdownV2")
}

func (t *TelegramSender) sendMessage(message, parseMode string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	payload := map[string]interface{}{
		"chat_id": t.chatID,
		"text":    message,
	}

	if parseMode != "" {
		payload["parse_mode"] = parseMode
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}
