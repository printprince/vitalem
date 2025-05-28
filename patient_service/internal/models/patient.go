package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Константы для физической активности
const (
	ActivityInactive      = "Неактивный"
	ActivityLowActive     = "Малоактивный"
	ActivityActive        = "Активный"
	ActivityVeryActive    = "Очень активный"
	ActivityExtremeActive = "Экстремально активный"
)

// Patient модель пациента
type Patient struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key"`
	UserID              uuid.UUID      `gorm:"type:uuid;index"`
	Name                string         `gorm:"type:varchar(100)"`
	Surname             string         `gorm:"type:varchar(100)"`
	DateOfBirth         time.Time      `gorm:"type:date"`
	Gender              string         `gorm:"type:varchar(20)"`
	Email               string         `gorm:"type:varchar(255);uniqueIndex"`
	Phone               string         `gorm:"type:varchar(20)"`
	Height              float64        `gorm:"type:decimal(5,2)"` // в сантиметрах
	Weight              float64        `gorm:"type:decimal(5,2)"` // в килограммах
	PhysActivity        string         `gorm:"type:varchar(50)"`
	Diagnoses           pq.StringArray `gorm:"type:varchar(255)[]"`
	AdditionalDiagnoses pq.StringArray `gorm:"type:varchar(255)[]"`
	Allergens           pq.StringArray `gorm:"type:varchar(255)[]"`
	AdditionalAllergens pq.StringArray `gorm:"type:varchar(255)[]"`
	Diet                pq.StringArray `gorm:"type:varchar(255)[]"`
	AdditionalDiets     pq.StringArray `gorm:"type:varchar(255)[]"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// BeforeCreate - хук для генерации UUID перед созданием
func (p *Patient) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// FullName - возвращает полное имя пациента
func (p *Patient) FullName() string {
	return p.Surname + " " + p.Name
}
