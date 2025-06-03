package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Константы для уровней физической активности
// В будущем можно вынести в отдельную таблицу,
// но пока хватит и этого - упростит работу фронта
const (
	ActivityInactive      = "Неактивный"
	ActivityLowActive     = "Малоактивный"
	ActivityActive        = "Активный"
	ActivityVeryActive    = "Очень активный"
	ActivityExtremeActive = "Экстремально активный"
)

// Patient - основная модель пациента для хранения в БД
// Тут целый комбайн данных, чтобы не плодить таблицы.
// TODO: В будущем при увеличении базы вынести диагнозы, аллергены и диеты
// в отдельные таблицы с many-to-many связями для нормализации БД.
type Patient struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key"`
	UserID              uuid.UUID      `gorm:"type:uuid;index"` // Связь с юзером из identity_service
	Name                string         `gorm:"type:varchar(100)"`
	IIN                 string         `gorm:"type:varchar(12);uniqueIndex:idx_patients_iin,where:iin IS NOT NULL"` // ИИН должен быть уникальным, но может быть NULL
	Surname             string         `gorm:"type:varchar(100)"`
	DateOfBirth         time.Time      `gorm:"type:date"`
	Gender              string         `gorm:"type:varchar(20)"`
	Email               string         `gorm:"type:varchar(255);uniqueIndex"` // Уникальный для рассылок
	Phone               string         `gorm:"type:varchar(20)"`
	Height              float64        `gorm:"type:decimal(5,2)"`   // в сантиметрах - можем хранить до 999.99
	Weight              float64        `gorm:"type:decimal(5,2)"`   // в килограммах - аналогично до 999.99
	PhysActivity        string         `gorm:"type:varchar(50)"`    // Из констант выше
	Diagnoses           pq.StringArray `gorm:"type:varchar(255)[]"` // Основные диагнозы из справочника
	AdditionalDiagnoses pq.StringArray `gorm:"type:varchar(255)[]"` // Произвольные диагнозы
	Allergens           pq.StringArray `gorm:"type:varchar(255)[]"` // Аллергены из справочника
	AdditionalAllergens pq.StringArray `gorm:"type:varchar(255)[]"` // Произвольные аллергены
	Diet                pq.StringArray `gorm:"type:varchar(255)[]"` // Диеты из справочника
	AdditionalDiets     pq.StringArray `gorm:"type:varchar(255)[]"` // Произвольные диеты
	CreatedAt           time.Time      // Стандартные GORM метки времени
	UpdatedAt           time.Time      // Автоматически обновляется при изменении
}

// BeforeCreate - хук GORM для автогенерации UUID перед вставкой в базу
// Если ID не задан (пустой или nil), то генерим новый UUID v4
// Это фулпруф от ошибок в коде и битых ID в базе
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
