package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"CalendarService/internal/domain/models"

	"github.com/google/uuid"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Структура запроса на бронирование
type BookingNotification struct {
	EventID    string `json:"event_id"`
	PatientID  string `json:"patient_id"`
	Specialist string `json:"specialist_id"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
}

func (c *Client) SendBookingNotifications(
	ctx context.Context,
	event *models.Event,
	patientEmail string, // из токена
	doctorEmail string, // из таблицы
	patientID uuid.UUID, // из токена
	doctorID uuid.UUID, // из таблицы doctors
) error {
	// Сообщение для клиента (пациента)
	patientMsg := fmt.Sprintf(
		"Здравствуйте! Вы успешно записались на приём к доктору.\nДоктор: %s\nВремя: %s - %s\nФормат: %s\nID записи: %s.",
		doctorEmail, // можно имя доктора тоже добавить
		event.StartTime.Format("02.01.2006 15:04"),
		event.EndTime.Format("02.01.2006 15:04"),
		event.AppointmentType,
		event.ID.String(),
	)

	// Сообщение для доктора
	doctorMsg := fmt.Sprintf(
		"У вас новая запись! Пациент: %s (id: %s)\nВремя: %s - %s\nФормат: %s\nID записи: %s.",
		patientEmail, patientID.String(),
		event.StartTime.Format("02.01.2006 15:04"),
		event.EndTime.Format("02.01.2006 15:04"),
		event.AppointmentType,
		event.ID.String(),
	)

	// Для клиента
	clientPayload := map[string]interface{}{
		"type":        "appointment.confirmed",
		"channel":     "email",
		"recipientId": patientID.String(),
		"recipient":   patientEmail,
		"message":     patientMsg,
	}

	// Для доктора
	doctorPayload := map[string]interface{}{
		"type":        "appointment.new",
		"channel":     "email",
		"recipientId": doctorID.String(),
		"recipient":   doctorEmail,
		"message":     doctorMsg,
	}

	// Шлём оба уведомления (по очереди)
	if err := c.post(ctx, "/notifications", clientPayload); err != nil {
		return fmt.Errorf("failed to send notification to patient: %w", err)
	}
	if err := c.post(ctx, "/notifications", doctorPayload); err != nil {
		return fmt.Errorf("failed to send notification to doctor: %w", err)
	}
	return nil
}

// Отправляет уведомление об отмене бронирования
func (c *Client) SendCancelNotification(ctx context.Context, event *models.Event) error {
	var patientIDStr string
	if event.PatientID != nil {
		patientIDStr = event.PatientID.String()
	}

	payload := map[string]interface{}{
		"event_id":    event.ID.String(),
		"specialist":  event.SpecialistID.String(),
		"start_time":  event.StartTime.Format(time.RFC3339),
		"end_time":    event.EndTime.Format(time.RFC3339),
		"patient_id":  patientIDStr,
		"canceled_at": time.Now().Format(time.RFC3339),
	}
	return c.post(ctx, "/notifications/cancel", payload)
}

// Вспомогательная функция POST с JSON
func (c *Client) post(ctx context.Context, path string, data interface{}) error {
	url := c.baseURL + path

	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("notification service returned status: %s", resp.Status)
	}

	return nil
}
