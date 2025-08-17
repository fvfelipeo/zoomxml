package storage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zoomxml/internal/logger"

	"github.com/zoomxml/config"
)

// StorageService interface para operações de storage
type StorageService interface {
	Initialize() error
	UploadFile(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error
	DownloadFile(ctx context.Context, bucketName, objectName string) ([]byte, error)
	DeleteFile(ctx context.Context, bucketName, objectName string) error
	FileExists(ctx context.Context, bucketName, objectName string) (bool, error)
}

// MinIOService implementa StorageService usando MinIO
type MinIOService struct {
	client *minio.Client
	config *config.StorageConfig
}

// NewMinIOService cria uma nova instância do serviço MinIO
func NewMinIOService() *MinIOService {
	cfg := config.Get()
	return &MinIOService{
		config: &cfg.Storage,
	}
}

// Initialize inicializa o cliente MinIO e cria o bucket se necessário
func (s *MinIOService) Initialize() error {
	logger.Printf("Initializing MinIO storage service...")
	logger.Printf("Endpoint: %s", s.config.Endpoint)
	logger.Printf("Bucket: %s", s.config.Bucket)

	// Inicializar cliente MinIO
	client, err := minio.New(s.config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.config.AccessKey, s.config.SecretKey, ""),
		Secure: s.config.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %v", err)
	}
	s.client = client

	// Verificar se o bucket existe, criar se necessário
	ctx := context.Background()
	exists, err := s.client.BucketExists(ctx, s.config.Bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.config.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}
		logger.Printf("Created MinIO bucket '%s'", s.config.Bucket)
	}

	logger.Printf("MinIO bucket '%s' ready", s.config.Bucket)
	logger.Println("MinIO storage service initialized successfully")
	return nil
}

// UploadFile faz upload de um arquivo
func (s *MinIOService) UploadFile(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error {
	logger.Printf("Uploading file: %s/%s (%d bytes)", bucketName, objectName, len(data))

	// Upload do arquivo para o MinIO
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		logger.Printf("Failed to upload file to MinIO: %v", err)
		return err
	}

	logger.Printf("Successfully uploaded file: %s/%s", bucketName, objectName)
	return nil
}

// DownloadFile faz download de um arquivo
func (s *MinIOService) DownloadFile(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	// TODO: Implementar download real
	logger.Printf("Downloading file: %s/%s", bucketName, objectName)
	return nil, fmt.Errorf("not implemented yet")
}

// DeleteFile remove um arquivo
func (s *MinIOService) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	// TODO: Implementar remoção real
	logger.Printf("Deleting file: %s/%s", bucketName, objectName)
	return nil
}

// FileExists verifica se um arquivo existe
func (s *MinIOService) FileExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	// TODO: Implementar verificação real
	logger.Printf("Checking if file exists: %s/%s", bucketName, objectName)
	return false, nil
}

// Global storage service instance
var Storage StorageService

// InitializeStorage inicializa o serviço de storage global
func InitializeStorage() error {
	Storage = NewMinIOService()
	return Storage.Initialize()
}
