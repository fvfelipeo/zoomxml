package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOProvider implements StorageProvider for MinIO S3
type MinIOProvider struct {
	client      *minio.Client
	bucketName  string
	pathBuilder *PathBuilder
}

// NewMinIOProvider creates a new MinIO storage provider
func NewMinIOProvider(config StorageConfig) (*MinIOProvider, error) {
	// Initialize MinIO client
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	provider := &MinIOProvider{
		client:      client,
		bucketName:  config.BucketName,
		pathBuilder: NewPathBuilder(""),
	}

	// Ensure bucket exists
	ctx := context.Background()
	err = provider.ensureBucket(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %v", err)
	}

	return provider, nil
}

// ensureBucket ensures the bucket exists, creates it if not
func (m *MinIOProvider) ensureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}
		log.Printf("✅ Created MinIO bucket: %s", m.bucketName)
	}

	return nil
}

// Upload uploads a file to MinIO
func (m *MinIOProvider) Upload(ctx context.Context, path string, data io.Reader, size int64, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	options := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := m.client.PutObject(ctx, m.bucketName, path, data, size, options)
	if err != nil {
		return fmt.Errorf("failed to upload file to MinIO: %v", err)
	}

	return nil
}

// Download downloads a file from MinIO
func (m *MinIOProvider) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, m.bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from MinIO: %v", err)
	}

	return object, nil
}

// Delete deletes a file from MinIO
func (m *MinIOProvider) Delete(ctx context.Context, path string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %v", err)
	}

	return nil
}

// List lists files in MinIO with a prefix
func (m *MinIOProvider) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	var files []FileInfo

	objectCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %v", object.Err)
		}

		files = append(files, FileInfo{
			Path:         object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			ETag:         object.ETag,
		})
	}

	return files, nil
}

// Exists checks if a file exists in MinIO
func (m *MinIOProvider) Exists(ctx context.Context, path string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		// Check if error is "object not found"
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %v", err)
	}

	return true, nil
}

// GetURL gets a presigned URL for a file
func (m *MinIOProvider) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucketName, path, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return url.String(), nil
}

// Copy copies a file within MinIO
func (m *MinIOProvider) Copy(ctx context.Context, srcPath, dstPath string) error {
	src := minio.CopySrcOptions{
		Bucket: m.bucketName,
		Object: srcPath,
	}

	dst := minio.CopyDestOptions{
		Bucket: m.bucketName,
		Object: dstPath,
	}

	_, err := m.client.CopyObject(ctx, dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	return nil
}

// Move moves a file within MinIO (copy + delete)
func (m *MinIOProvider) Move(ctx context.Context, srcPath, dstPath string) error {
	// First copy the file
	err := m.Copy(ctx, srcPath, dstPath)
	if err != nil {
		return fmt.Errorf("failed to copy file during move: %v", err)
	}

	// Then delete the source
	err = m.Delete(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("failed to delete source file during move: %v", err)
	}

	return nil
}

// GetFileInfo gets information about a file
func (m *MinIOProvider) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	stat, err := m.client.StatObject(ctx, m.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	return &FileInfo{
		Path:         path,
		Size:         stat.Size,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified,
		ETag:         stat.ETag,
		Metadata:     stat.UserMetadata,
	}, nil
}

// NFSeMinIOManager implements NFSeStorageManager for MinIO
type NFSeMinIOManager struct {
	provider    *MinIOProvider
	pathBuilder *PathBuilder
}

// NewNFSeMinIOManager creates a new NFS-e MinIO manager
func NewNFSeMinIOManager(provider *MinIOProvider) *NFSeMinIOManager {
	return &NFSeMinIOManager{
		provider:    provider,
		pathBuilder: NewPathBuilder(""),
	}
}

// StoreXML stores an XML file with proper organization
func (nm *NFSeMinIOManager) StoreXML(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string, xmlData []byte) (string, error) {
	// Generate current date for filename
	dataEmissao := time.Now().Format("2006-01-02")
	path := nm.pathBuilder.BuildXMLPath(empresaCNPJ, competencia, numeroNFSe, dataEmissao)

	reader := bytes.NewReader(xmlData)
	err := nm.provider.Upload(ctx, path, reader, int64(len(xmlData)), "application/xml")
	if err != nil {
		return "", fmt.Errorf("failed to store XML: %v", err)
	}

	return path, nil
}

// StoreZIP stores a ZIP file with proper organization
func (nm *NFSeMinIOManager) StoreZIP(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string, zipData []byte) (string, error) {
	dataEmissao := time.Now().Format("2006-01-02")
	path := nm.pathBuilder.BuildZIPPath(empresaCNPJ, competencia, numeroNFSe, dataEmissao)

	reader := bytes.NewReader(zipData)
	err := nm.provider.Upload(ctx, path, reader, int64(len(zipData)), "application/zip")
	if err != nil {
		return "", fmt.Errorf("failed to store ZIP: %v", err)
	}

	return path, nil
}

// StoreReport stores a processing report
func (nm *NFSeMinIOManager) StoreReport(ctx context.Context, empresaCNPJ, batchID string, reportData []byte) (string, error) {
	path := nm.pathBuilder.BuildReportPath(empresaCNPJ, batchID)

	reader := bytes.NewReader(reportData)
	err := nm.provider.Upload(ctx, path, reader, int64(len(reportData)), "text/plain")
	if err != nil {
		return "", fmt.Errorf("failed to store report: %v", err)
	}

	return path, nil
}

// GetXML retrieves an XML file
func (nm *NFSeMinIOManager) GetXML(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string) ([]byte, error) {
	// We need to find the file since we don't know the exact date
	prefix := nm.pathBuilder.BuildCompetenciaPrefix(empresaCNPJ, competencia) + "xml/"
	files, err := nm.provider.List(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list XML files: %v", err)
	}

	// Find the file with matching NFS-e number
	var targetPath string
	for _, file := range files {
		if strings.Contains(file.Path, fmt.Sprintf("nfse_%s_", numeroNFSe)) {
			targetPath = file.Path
			break
		}
	}

	if targetPath == "" {
		return nil, fmt.Errorf("XML file not found for NFS-e %s", numeroNFSe)
	}

	reader, err := nm.provider.Download(ctx, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download XML: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read XML data: %v", err)
	}

	return data, nil
}

// GetZIP retrieves a ZIP file
func (nm *NFSeMinIOManager) GetZIP(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string) ([]byte, error) {
	prefix := nm.pathBuilder.BuildCompetenciaPrefix(empresaCNPJ, competencia) + "zip/"
	files, err := nm.provider.List(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list ZIP files: %v", err)
	}

	var targetPath string
	for _, file := range files {
		if strings.Contains(file.Path, fmt.Sprintf("nfse_%s_", numeroNFSe)) {
			targetPath = file.Path
			break
		}
	}

	if targetPath == "" {
		return nil, fmt.Errorf("ZIP file not found for NFS-e %s", numeroNFSe)
	}

	reader, err := nm.provider.Download(ctx, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download ZIP: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read ZIP data: %v", err)
	}

	return data, nil
}

// ListXMLs lists XML files for a company and competencia
func (nm *NFSeMinIOManager) ListXMLs(ctx context.Context, empresaCNPJ, competencia string) ([]FileInfo, error) {
	prefix := nm.pathBuilder.BuildCompetenciaPrefix(empresaCNPJ, competencia) + "xml/"
	return nm.provider.List(ctx, prefix)
}

// DeleteXML deletes an XML file
func (nm *NFSeMinIOManager) DeleteXML(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string) error {
	// Find the file first
	prefix := nm.pathBuilder.BuildCompetenciaPrefix(empresaCNPJ, competencia) + "xml/"
	files, err := nm.provider.List(ctx, prefix)
	if err != nil {
		return fmt.Errorf("failed to list XML files: %v", err)
	}

	var targetPath string
	for _, file := range files {
		if strings.Contains(file.Path, fmt.Sprintf("nfse_%s_", numeroNFSe)) {
			targetPath = file.Path
			break
		}
	}

	if targetPath == "" {
		return fmt.Errorf("XML file not found for NFS-e %s", numeroNFSe)
	}

	return nm.provider.Delete(ctx, targetPath)
}

// GetXMLPath generates the storage path for an XML file
func (nm *NFSeMinIOManager) GetXMLPath(empresaCNPJ, competencia, numeroNFSe string) string {
	dataEmissao := time.Now().Format("2006-01-02")
	return nm.pathBuilder.BuildXMLPath(empresaCNPJ, competencia, numeroNFSe, dataEmissao)
}

// GetZIPPath generates the storage path for a ZIP file
func (nm *NFSeMinIOManager) GetZIPPath(empresaCNPJ, competencia, numeroNFSe string) string {
	dataEmissao := time.Now().Format("2006-01-02")
	return nm.pathBuilder.BuildZIPPath(empresaCNPJ, competencia, numeroNFSe, dataEmissao)
}

// GetReportPath generates the storage path for a report
func (nm *NFSeMinIOManager) GetReportPath(empresaCNPJ, batchID string) string {
	return nm.pathBuilder.BuildReportPath(empresaCNPJ, batchID)
}
