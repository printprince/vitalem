package service

import (
	"context"
	"encoding/json"

	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/specialist_service/internal/models"
	"github.com/streadway/amqp"
)

type MessageService interface {
	PublishDoctorCreated(ctx context.Context, event *models.DoctorCreatedEvent) error
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
	doctorQueueName string, // Оставляем для совместимости, но не используем
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

	// Exchange "vitalem" уже существует, не пытаемся его переобъявить
	// Это предотвращает ошибку ACCESS_REFUSED

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

	// НЕ создаем QueueBind - binding уже создан в identity_service
	// Это предотвращает конфликты при множественных consumer'ах

	return &messageService{
		conn:          conn,
		channel:       channel,
		exchange:      exchange,
		userQueueName: userQueueName,
		routingKey:    routingKey,
		logger:        logger,
	}, nil
}

func (s *messageService) PublishDoctorCreated(ctx context.Context, event *models.DoctorCreatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal doctor created event", map[string]interface{}{
			"error":    err.Error(),
			"doctorID": event.DoctorID,
		})
		return err
	}

	// Публикуем события о создании доктора с отдельным routing key
	err = s.channel.Publish(
		s.exchange,
		"doctor.created", // используем специальный routing key для событий докторов
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		s.logger.Error("Failed to publish doctor created event", map[string]interface{}{
			"error":    err.Error(),
			"doctorID": event.DoctorID,
		})
		return err
	}

	s.logger.Info("Published doctor created event", map[string]interface{}{
		"doctorID": event.DoctorID,
	})
	return nil
}

func (s *messageService) ConsumeUserCreated(ctx context.Context) (<-chan models.UserCreatedEvent, error) {
	s.logger.Info("=== STARTING ConsumeUserCreated ===", map[string]interface{}{
		"queue":      s.userQueueName,
		"routingKey": s.routingKey,
		"exchange":   s.exchange,
	})

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
		s.logger.Info("=== CONSUMER GOROUTINE STARTED ===", nil)

		for d := range msgs {
			s.logger.Info("=== RAW MESSAGE RECEIVED ===", map[string]interface{}{
				"body":         string(d.Body),
				"routing_key":  d.RoutingKey,
				"exchange":     d.Exchange,
				"content_type": d.ContentType,
				"headers":      d.Headers,
			})

			var event models.UserCreatedEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				s.logger.Error("Failed to unmarshal user created event", map[string]interface{}{
					"error":    err.Error(),
					"raw_body": string(d.Body),
				})
				continue
			}

			s.logger.Info("Successfully unmarshaled event", map[string]interface{}{
				"userID": event.UserID,
				"email":  event.Email,
				"role":   event.Role,
			})

			s.logger.Info("Received user created event", map[string]interface{}{
				"userID": event.UserID,
				"email":  event.Email,
				"role":   event.Role,
			})

			s.logger.Info("Sending event to channel", map[string]interface{}{
				"userID": event.UserID,
			})

			select {
			case events <- event:
				s.logger.Info("Event sent to channel successfully", map[string]interface{}{
					"userID": event.UserID,
				})
			case <-ctx.Done():
				s.logger.Info("Context cancelled, stopping consumer", nil)
				return
			}
		}
		s.logger.Info("=== CONSUMER GOROUTINE ENDED ===", nil)
	}()

	return events, nil
}

func (s *messageService) Close() error {
	if err := s.channel.Close(); err != nil {
		return err
	}

	return s.conn.Close()
}
