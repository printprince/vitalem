// internal/service/file_service.go
package service

import (
	"FileServerService/internal/storage"
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	"FileServerService/internal/model"
	"FileServerService/internal/repository"
)

// FileService описывает бизнес-логику работы с файлами.
type FileService interface {
	Upload(ctx context.Context, file *model.File, fileReader io.Reader, fileSize int64) error
	Get(ctx context.Context, id string) (*model.File, error)
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string) ([]model.File, error)

	DownloadFile(ctx context.Context, id string) (io.ReadCloser, string, string, error) // content, filename, contentType
	DownloadByPath(ctx context.Context, path string) (io.ReadCloser, error)
	Update(ctx context.Context, file *model.File) error
	TogglePublic(ctx context.Context, id string, userID uuid.UUID) error
}

// fileService — конкретная реализация FileService.
type fileService struct {
	repo        repository.FileRepository
	minioClient *storage.MinioClient
}

// NewFileService создаёт новый экземпляр fileService.
func NewFileService(repo repository.FileRepository, minioClient *storage.MinioClient) FileService {
	return &fileService{repo: repo, minioClient: minioClient}
}

// Upload сохраняет информацию о файле и присваивает ID и время загрузки.
// Добавляем параметр fileBytes для загрузки в MinIO
func (s *fileService) Upload(ctx context.Context, file *model.File, fileReader io.Reader, fileSize int64) error {
	// Генерация ID и установка времени загрузки
	file.ID = uuid.New()
	file.UploadedAt = time.Now()

	// Генерация имени объекта для хранения в MinIO
	objectName := file.ID.String() + "_" + file.Name

	// Загружаем файл в MinIO
	err := s.minioClient.UploadFile(ctx, objectName, fileReader, fileSize, file.ContentType)
	if err != nil {
		return err
	}

	// Сохраняем путь к объекту (только имя, bucket известен в MinioClient)
	file.Path = objectName

	// Сохраняем метаданные в БД
	return s.repo.Save(ctx, file)
}

// Get возвращает файл по его ID.
func (s *fileService) Get(ctx context.Context, id string) (*model.File, error) {
	return s.repo.GetByID(ctx, id)
}

// Delete удаляет файл по его ID.
func (s *fileService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ListByUser возвращает список файлов, принадлежащих пользователю.
func (s *fileService) ListByUser(ctx context.Context, userID string) ([]model.File, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *fileService) DownloadFile(ctx context.Context, id string) (io.ReadCloser, string, string, error) {
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", "", err
	}

	reader, err := s.minioClient.DownloadFile(ctx, file.Path)
	if err != nil {
		return nil, "", "", err
	}

	return reader, file.Name, file.ContentType, nil
}
func (s *fileService) DownloadByPath(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.minioClient.DownloadFile(ctx, path)
}

func (s *fileService) Update(ctx context.Context, file *model.File) error {
	file.UpdatedAt = time.Now()
	return s.repo.Update(ctx, file)
}

func (s *fileService) TogglePublic(ctx context.Context, id string, userID uuid.UUID) error {
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if file.UserID != userID {
		return ErrForbidden // ты можешь определить эту ошибку как кастомную
	}

	file.IsPublic = !file.IsPublic
	file.UpdatedAt = time.Now()

	return s.repo.Update(ctx, file)
}
