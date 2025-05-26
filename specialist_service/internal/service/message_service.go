package service

import (
	"context"
	"encoding/json"

	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/specialist_service/internal/models"
	"github.com/streadway/amqp" // Добавьте эту зависимость в go.mod
)

type MessageService interface {
	PublishDoctorCreated(ctx context.Context, event *models.DoctorCreatedEvent) error
	ConsumeUserCreated(ctx context.Context) (<-chan models.UserCreatedEvent, error)
	Close() error
}

type messageService struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	exchange        string
	doctorQueueName string
	userQueueName   string
	routingKey      string
	logger          *logger.Client
}

func NewMessageService(
	rabbitMQURL string,
	exchange string,
	doctorQueueName string,
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

	// Обьявляем очередь для докторов
	_, err = channel.QueueDeclare(
		doctorQueueName, // имя
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // аргументы
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Привязываем очередь к exchange
	err = channel.QueueBind(
		doctorQueueName, // имя очереди
		routingKey,      // ключ маршрутизации
		exchange,        // имя exchange
		false,           // no-wait
		nil,             // аргументы
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Объявляем очередь для пользователей
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
		userQueueName,  // имя очереди
		"user.created", // ключ маршрутизации
		exchange,       // имя exchange
		false,          // no-wait
		nil,            // аргументы
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &messageService{
		conn:            conn,
		channel:         channel,
		exchange:        exchange,
		doctorQueueName: doctorQueueName,
		userQueueName:   userQueueName,
		routingKey:      routingKey,
		logger:          logger,
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

	err = s.channel.Publish(
		s.exchange,
		s.routingKey,
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
