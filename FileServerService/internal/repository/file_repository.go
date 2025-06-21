// internal/repository/file_repository.go
package repository

import (
	"context"

	"github.com/printprince/vitalem/FileServerService/internal/model"
	"gorm.io/gorm"
)

type FileRepository interface {
	Save(ctx context.Context, file *model.File) error
	GetByID(ctx context.Context, id string) (*model.File, error)
	Update(ctx context.Context, file *model.File) error
	Delete(ctx context.Context, id string) error
	ListByUserID(ctx context.Context, userID string) ([]model.File, error)
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Save(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *fileRepository) GetByID(ctx context.Context, id string) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) Update(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *fileRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.File{}, "id = ?", id).Error
}

func (r *fileRepository) ListByUserID(ctx context.Context, userID string) ([]model.File, error) {
	var files []model.File
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&files).Error
	return files, err
}
