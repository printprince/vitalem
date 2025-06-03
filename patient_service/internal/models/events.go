package models

import (
	"github.com/google/uuid"
)

// PatientCreatedEvent событие создания пациента в системе
type PatientCreatedEvent struct {
	UserID    uuid.UUID `json:"user_id"`
	PatientID uuid.UUID `json:"patient_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
}

// UserCreatedEvent событие создания пользователя в identity_service
type UserCreatedEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
