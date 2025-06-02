package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Event — структура события в календаре.
// Используется для хранения информации о приеме специалиста и записи пациента.
type Event struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	Title           string     `gorm:"type:varchar(255);not null" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	StartTime       time.Time  `gorm:"not null" json:"start_time"`
	EndTime         time.Time  `gorm:"not null" json:"end_time"`
	SpecialistID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"specialist_id"`
	PatientID       *uuid.UUID `gorm:"type:uuid;index" json:"patient_id"`
	Status          string     `gorm:"type:varchar(20);not null;default:'available';check:status IN ('available', 'booked', 'canceled')" json:"status"`
	AppointmentType string     `gorm:"type:varchar(10);not null;default:'offline'" json:"appointment_type"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName указывает GORM использовать имя таблицы "events"
func (Event) TableName() string {
	return "events"
}

// BeforeCreate - хук для генерации UUID перед созданием
func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
