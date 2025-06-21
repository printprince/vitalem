package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/printprince/vitalem/FileServerService/internal/config"
)

type MinioClient struct {
	client *minio.Client
	bucket string
}

func NewMinioClient(cfg config.MinIOConfig) (*MinioClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	// Проверка существования бакета и создание, если надо
	ctx := context.Background()
	bucketName := "vitalem-files"
	exists, errBucketExists := client.BucketExists(ctx, bucketName)
	if errBucketExists != nil {
		return nil, errBucketExists
	}
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &MinioClient{
		client: client,
		bucket: bucketName,
	}, nil
}

// UploadFile загружает файл в MinIO из io.Reader с известным размером и типом контента
func (m *MinioClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.client.PutObject(
		ctx,
		m.bucket,
		objectName,
		reader,
		objectSize,
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}

// DownloadFile загружает файл из MinIO и возвращает поток (io.ReadCloser)
func (m *MinioClient) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	// Пробуем прочитать метаинформацию чтобы убедиться, что объект существует
	_, err = object.Stat()
	if err != nil {
		return nil, err
	}
	return object, nil
}

// DeleteFile удаляет файл из MinIO
func (m *MinioClient) DeleteFile(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{})
	return err
}
