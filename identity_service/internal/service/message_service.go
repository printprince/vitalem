package service

import (
	"context"
	"encoding/json"
	"fmt"
	"identity_service/internal/models"
	"log"

	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/streadway/amqp"
)

// MessageService представляет сервис для работы с сообщениями
type MessageService interface {
	// PublishUserCreated публикует событие создания пользователя
	PublishUserCreated(ctx context.Context, event *models.UserCreatedEvent) error
	// Close закрывает соединение с RabbitMQ
	Close() error
}

// messageService реализация сервиса сообщений
type messageService struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	queue      string
	routingKey string
	logger     *logger.Client
}

// NewMessageService создает новый сервис сообщений
func NewMessageService(
	rabbitMQURL string,
	exchange string,
	queue string,
	routingKey string,
	logger *logger.Client,
) (MessageService, error) {
	// Подключаемся к RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to connect to RabbitMQ", map[string]interface{}{
				"error": err.Error(),
				"url":   rabbitMQURL,
			})
		}
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Создаем канал
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		if logger != nil {
			logger.Error("Failed to open RabbitMQ channel", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return nil, fmt.Errorf("failed to open channel: %w", err)
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
		if logger != nil {
			logger.Error("Failed to declare exchange", map[string]interface{}{
				"error":    err.Error(),
				"exchange": exchange,
			})
		}
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Объявляем очередь
	_, err = channel.QueueDeclare(
		queue, // имя
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // аргументы
	)
	if err != nil {
		channel.Close()
		conn.Close()
		if logger != nil {
			logger.Error("Failed to declare queue", map[string]interface{}{
				"error": err.Error(),
				"queue": queue,
			})
		}
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Привязываем очередь к exchange
	err = channel.QueueBind(
		queue,      // имя очереди
		routingKey, // ключ маршрутизации
		exchange,   // имя exchange
		false,      // no-wait
		nil,        // аргументы
	)
	if err != nil {
		channel.Close()
		conn.Close()
		if logger != nil {
			logger.Error("Failed to bind queue", map[string]interface{}{
				"error":      err.Error(),
				"queue":      queue,
				"exchange":   exchange,
				"routingKey": routingKey,
			})
		}
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	if logger != nil {
		logger.Info("RabbitMQ connection established", map[string]interface{}{
			"exchange":   exchange,
			"queue":      queue,
			"routingKey": routingKey,
		})
	} else {
		log.Printf("RabbitMQ connection established: %s, %s, %s", exchange, queue, routingKey)
	}

	return &messageService{
		conn:       conn,
		channel:    channel,
		exchange:   exchange,
		queue:      queue,
		routingKey: routingKey,
		logger:     logger,
	}, nil
}

// PublishUserCreated публикует событие создания пользователя
func (s *messageService) PublishUserCreated(ctx context.Context, event *models.UserCreatedEvent) error {
	// Кодируем событие в JSON
	body, err := json.Marshal(event)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to marshal user created event", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Публикуем сообщение
	err = s.channel.Publish(
		s.exchange,   // exchange
		s.routingKey, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // сообщение будет сохранено при перезапуске RabbitMQ
		})
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to publish user created event", map[string]interface{}{
				"error":  err.Error(),
				"userID": event.UserID,
			})
		}
		return fmt.Errorf("failed to publish message: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("User created event published", map[string]interface{}{
			"userID": event.UserID,
			"email":  event.Email,
			"role":   event.Role,
		})
	} else {
		log.Printf("User created event published: %s, %s, %s", event.UserID, event.Email, event.Role)
	}

	return nil
}

// Close закрывает соединение с RabbitMQ
func (s *messageService) Close() error {
	if s.channel != nil {
		if err := s.channel.Close(); err != nil {
			if s.logger != nil {
				s.logger.Error("Failed to close RabbitMQ channel", map[string]interface{}{
					"error": err.Error(),
				})
			}
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			if s.logger != nil {
				s.logger.Error("Failed to close RabbitMQ connection", map[string]interface{}{
					"error": err.Error(),
				})
			}
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	if s.logger != nil {
		s.logger.Info("RabbitMQ connection closed", nil)
	}

	return nil
}
