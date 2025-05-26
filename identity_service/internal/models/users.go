package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleAdmin   = "admin"
	RoleDoctor  = "doctor"
	RolePatient = "patient"
)

type Users struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email          string    `gorm:"type:varchar(255);uniqueIndex:idx_email"`
	HashedPassword string    `gorm:"size:255"`
	Role           string
	CreatedAt      time.Time
}

// TableName указывает GORM использовать имя таблицы "users"
func (Users) TableName() string {
	return "users"
}

// BeforeCreate - хук для генерации UUID перед созданием
func (u *Users) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
