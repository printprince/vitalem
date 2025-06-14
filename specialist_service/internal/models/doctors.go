// Модель Doctor - врача
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Список специализаций врачей
const (
	RoleNotSpecified          = "Не указана"
	RoleRheumatologist        = "Ревматолог"
	RoleCardiologist          = "Кардиолог"
	RoleTherapist             = "Терапевт"
	RoleGastroenterologist    = "Гастроэнтеролог"
	RoleNeurosurgeon          = "Нейрохирург"
	RoleGynecologist          = "Гинеколог"
	RoleOphthalmologist       = "Офтальмолог"
	RoleOtolaryngologist      = "Отоларинголог"
	RoleSurgeon               = "Хирург"
	RoleUrologist             = "Уролог"
	RolePulmonologist         = "Пульмонолог"
	RolePediatrician          = "Педиатр"
	RoleOncologist            = "Онколог"
	RoleInfectiousDiseases    = "Инфекционист"
	RoleImmunologist          = "Иммунолог"
	RoleAllergist             = "Аллерголог"
	RoleHepatologist          = "Гепатолог"
	RoleHematologist          = "Гематолог"
	RolePsychiatrist          = "Психиатр"
	RoleCardiovascularSurgeon = "Сердечно-сосудистый хирург"
	RoleDermatologist         = "Дерматолог"
	RoleNephrologist          = "Нефролог"
	RoleNeurologist           = "Невролог"
	RoleEndocrinologist       = "Эндокринолог"
	RoleOrthopedist           = "Ортопед-травматолог"
)

// Doctor модель врача
type Doctor struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key"`
	UserID       uuid.UUID      `gorm:"type:uuid;index"`
	FirstName    string         `gorm:"type:varchar(100)"`
	MiddleName   string         `gorm:"type:varchar(100)"`
	LastName     string         `gorm:"type:varchar(100)"`
	Description  string         `gorm:"type:text"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex"`
	Phone        string         `gorm:"type:varchar(20)"`
	AvatarURL    string         `gorm:"type:varchar(255)" json:"avatar_url"`
	Roles        pq.StringArray `gorm:"type:varchar(100)[]"`
	Price        float64        `gorm:"type:decimal(10,2)"`  // Цена за прием
	Education    pq.StringArray `gorm:"type:varchar(255)[]"` // Массив образований
	Certificates pq.StringArray `gorm:"type:varchar(255)[]"` // Массив сертификатов
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// BeforeCreate - хук для генерации UUID перед созданием
func (d *Doctor) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// FullName - возвращает полное имя врача
func (d *Doctor) FullName() string {
	if d.MiddleName != "" {
		return d.LastName + " " + d.FirstName + " " + d.MiddleName
	}

	return d.LastName + " " + d.FirstName
}
