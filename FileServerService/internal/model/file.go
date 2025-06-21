// internal/model/file.go
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name         string `gorm:"type:varchar(255);not null" json:"name"`
	OriginalName string `gorm:"type:varchar(255);not null" json:"original_name"`
	MimeType     string `gorm:"type:varchar(100);not null" json:"mime_type"`
	Size         int64  `gorm:"not null" json:"size"`
	Bucket       string `gorm:"type:varchar(100);not null" json:"bucket"`
	Path         string `gorm:"type:varchar(255);not null" json:"path"`

	UserID   uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	IsPublic bool      `gorm:"default:false" json:"is_public"`
}

// Для обратной совместимости с текущим кодом
func (f *File) ContentType() string {
	return f.MimeType
}

func (f *File) SetContentType(contentType string) {
	f.MimeType = contentType
}

func (f *File) UploadedAt() time.Time {
	return f.CreatedAt
}

func (f *File) SetUploadedAt(t time.Time) {
	f.CreatedAt = t
}
