// internal/model/file.go
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Size        int64  `gorm:"not null" json:"size"`
	ContentType string `gorm:"type:varchar(100);not null" json:"content_type"`
	Path        string `gorm:"type:text;not null" json:"path"`

	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	UploadedAt time.Time `gorm:"not null" json:"uploaded_at"`
	IsPublic   bool      `gorm:"default:false" json:"is_public"`
}
