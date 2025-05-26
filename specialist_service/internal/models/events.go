package models

import (
	"github.com/google/uuid"
)

// DoctorCreatedEvent событие создания врача в системе
type DoctorCreatedEvent struct {
	UserID   uuid.UUID `json:"user_id"`
	DoctorID uuid.UUID `json:"doctor_id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	Roles    []string  `json:"roles"`
}

// UserCreatedEvent событие создания пользователя в identity_service
type UserCreatedEvent struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}
