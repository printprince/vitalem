package models

// UserCreatedEvent представляет событие создания пользователя,
// которое будет отправлено в RabbitMQ
type UserCreatedEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
