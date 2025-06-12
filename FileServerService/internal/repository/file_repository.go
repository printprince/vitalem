// internal/repository/file_repository.go
package repository

import (
	"context"

	"github.com/vitalem/fileserver/internal/model"
	"gorm.io/gorm"
)

type FileRepository interface {
	Save(ctx context.Context, file *model.File) error
	GetByID(ctx context.Context, id string) (*model.File, error)
	Delete(ctx context.Context, id string) error
	ListByUserID(ctx context.Context, userID string) ([]model.File, error)

	Update(ctx context.Context, file *model.File) error // <-- добавлено
}

type fileRepo struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepo{db: db}
}

func (r *fileRepo) Save(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *fileRepo) GetByID(ctx context.Context, id string) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.File{}, "id = ?", id).Error
}

func (r *fileRepo) ListByUserID(ctx context.Context, userID string) ([]model.File, error) {
	var files []model.File
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&files).Error
	return files, err
}
func (r *fileRepo) Update(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Model(&model.File{}).
		Where("id = ?", file.ID).
		Updates(file).Error
}
