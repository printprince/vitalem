package service

import (
	"context"
	"encoding/json"

	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/patient_service/internal/models"
	"github.com/streadway/amqp"
)

type MessageService interface {
	PublishPatientCreated(ctx context.Context, event *models.PatientCreatedEvent) error
	ConsumeUserCreated(ctx context.Context) (<-chan models.UserCreatedEvent, error)
	Close() error
}

type messageService struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	exchange      string
	userQueueName string
	routingKey    string
	logger        *logger.Client
}

func NewMessageService(
	rabbitMQURL string,
	exchange string,
	patientQueueName string, // Оставляем для совместимости, но не используем
	userQueueName string,
	routingKey string,
	logger *logger.Client,
) (MessageService, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Объявляем exchange
	err = channel.ExchangeDeclare(
		exchange, // имя
		"topic",  // тип
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // аргументы
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Объявляем только очередь для получения событий пользователей
	_, err = channel.QueueDeclare(
		userQueueName, // имя
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // аргументы
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Привязываем очередь к exchange
	err = channel.QueueBind(
		userQueueName, // имя очереди
		routingKey,    // ключ маршрутизации из конфигурации
		exchange,      // имя exchange
		false,         // no-wait
		nil,           // аргументы
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &messageService{
		conn:          conn,
		channel:       channel,
		exchange:      exchange,
		userQueueName: userQueueName,
		routingKey:    routingKey,
		logger:        logger,
	}, nil
}

func (s *messageService) PublishPatientCreated(ctx context.Context, event *models.PatientCreatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal patient created event", map[string]interface{}{
			"error":     err.Error(),
			"patientID": event.PatientID,
		})
		return err
	}

	// Публикуем события о создании пациента с отдельным routing key
	err = s.channel.Publish(
		s.exchange,
		"patient.created", // используем специальный routing key для событий пациентов
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		s.logger.Error("Failed to publish patient created event", map[string]interface{}{
			"error":     err.Error(),
			"patientID": event.PatientID,
		})
		return err
	}

	s.logger.Info("Published patient created event", map[string]interface{}{
		"patientID": event.PatientID,
	})
	return nil
}

func (s *messageService) ConsumeUserCreated(ctx context.Context) (<-chan models.UserCreatedEvent, error) {
	msgs, err := s.channel.Consume(
		s.userQueueName, // queue
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		s.logger.Error("Failed to register a consumer", map[string]interface{}{
			"error": err.Error(),
			"queue": s.userQueueName,
		})
		return nil, err
	}

	s.logger.Info("Started consuming user created events", map[string]interface{}{
		"queue":      s.userQueueName,
		"routingKey": s.routingKey,
	})

	events := make(chan models.UserCreatedEvent)

	go func() {
		defer close(events)

		for d := range msgs {
			var event models.UserCreatedEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				s.logger.Error("Failed to unmarshal user created event", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			s.logger.Info("Received user created event", map[string]interface{}{
				"userID": event.UserID,
				"email":  event.Email,
				"role":   event.Role,
			})

			select {
			case events <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return events, nil
}

func (s *messageService) Close() error {
	if err := s.channel.Close(); err != nil {
		return err
	}

	return s.conn.Close()
}
