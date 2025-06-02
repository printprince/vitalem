package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Doctor - модель врача в календарном сервисе
type Doctor struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName указывает GORM использовать имя таблицы "doctors"
func (Doctor) TableName() string {
	return "doctors"
}

// BeforeCreate - хук для генерации UUID перед созданием
func (d *Doctor) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
