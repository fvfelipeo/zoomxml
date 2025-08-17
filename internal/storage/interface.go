package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// StorageProvider defines the interface for file storage operations
type StorageProvider interface {
	// Upload uploads a file to storage
	Upload(ctx context.Context, path string, data io.Reader, size int64, contentType string) error

	// Download downloads a file from storage
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete deletes a file from storage
	Delete(ctx context.Context, path string) error

	// List lists files in a directory
	List(ctx context.Context, prefix string) ([]FileInfo, error)

	// Exists checks if a file exists
	Exists(ctx context.Context, path string) (bool, error)

	// GetURL gets a presigned URL for a file
	GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// Copy copies a file within storage
	Copy(ctx context.Context, srcPath, dstPath string) error

	// Move moves a file within storage
	Move(ctx context.Context, srcPath, dstPath string) error

	// GetFileInfo gets information about a file
	GetFileInfo(ctx context.Context, path string) (*FileInfo, error)
}

// FileInfo represents information about a stored file
type FileInfo struct {
	Path         string            `json:"path"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	ETag         string            `json:"etag"`
	Metadata     map[string]string `json:"metadata"`
}

// UploadOptions provides options for file uploads
type UploadOptions struct {
	ContentType          string
	Metadata             map[string]string
	ServerSideEncryption bool
	StorageClass         string
}

// ListOptions provides options for listing files
type ListOptions struct {
	Prefix    string
	Delimiter string
	MaxKeys   int
	Recursive bool
}

// StorageConfig holds configuration for storage providers
type StorageConfig struct {
	Provider   string `json:"provider"`
	Endpoint   string `json:"endpoint"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	BucketName string `json:"bucket_name"`
	Region     string `json:"region"`
	UseSSL     bool   `json:"use_ssl"`
	PathStyle  bool   `json:"path_style"`
}

// NFSeStorageManager manages NFS-e file organization
type NFSeStorageManager interface {
	// StoreXML stores an XML file with proper organization
	StoreXML(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string, xmlData []byte) (string, error)

	// StoreZIP stores a ZIP file with proper organization
	StoreZIP(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string, zipData []byte) (string, error)

	// StoreReport stores a processing report
	StoreReport(ctx context.Context, empresaCNPJ, batchID string, reportData []byte) (string, error)

	// GetXML retrieves an XML file
	GetXML(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string) ([]byte, error)

	// GetZIP retrieves a ZIP file
	GetZIP(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string) ([]byte, error)

	// ListXMLs lists XML files for a company and competencia
	ListXMLs(ctx context.Context, empresaCNPJ, competencia string) ([]FileInfo, error)

	// DeleteXML deletes an XML file
	DeleteXML(ctx context.Context, empresaCNPJ, competencia, numeroNFSe string) error

	// GetXMLPath generates the storage path for an XML file
	GetXMLPath(empresaCNPJ, competencia, numeroNFSe string) string

	// GetZIPPath generates the storage path for a ZIP file
	GetZIPPath(empresaCNPJ, competencia, numeroNFSe string) string

	// GetReportPath generates the storage path for a report
	GetReportPath(empresaCNPJ, batchID string) string
}

// PathBuilder helps build storage paths following the organization structure
type PathBuilder struct {
	BasePath string
}

// NewPathBuilder creates a new path builder
func NewPathBuilder(basePath string) *PathBuilder {
	return &PathBuilder{BasePath: basePath}
}

// BuildXMLPath builds path for XML files: {cnpj}/{ano}/{mes}/xml/nfse_{numero}_{data}.xml
func (pb *PathBuilder) BuildXMLPath(cnpj, competencia, numeroNFSe, dataEmissao string) string {
	// competencia format: "2025-08"
	// dataEmissao format: "2025-08-08"
	ano := competencia[:4]
	mes := competencia[5:7]
	dataFormatted := strings.ReplaceAll(dataEmissao, "-", "")

	return fmt.Sprintf("%s/%s/%s/%s/xml/nfse_%s_%s.xml",
		pb.BasePath, cnpj, ano, mes, numeroNFSe, dataFormatted)
}

// BuildZIPPath builds path for ZIP files: {cnpj}/{ano}/{mes}/zip/nfse_{numero}_{data}.zip
func (pb *PathBuilder) BuildZIPPath(cnpj, competencia, numeroNFSe, dataEmissao string) string {
	ano := competencia[:4]
	mes := competencia[5:7]
	dataFormatted := strings.ReplaceAll(dataEmissao, "-", "")

	return fmt.Sprintf("%s/%s/%s/%s/zip/nfse_%s_%s.zip",
		pb.BasePath, cnpj, ano, mes, numeroNFSe, dataFormatted)
}

// BuildReportPath builds path for reports: {cnpj}/relatorios/processing_report_{batch_id}.txt
func (pb *PathBuilder) BuildReportPath(cnpj, batchID string) string {
	return fmt.Sprintf("%s/%s/relatorios/processing_report_%s.txt",
		pb.BasePath, cnpj, batchID)
}

// BuildCompetenciaPrefix builds prefix for listing files by competencia
func (pb *PathBuilder) BuildCompetenciaPrefix(cnpj, competencia string) string {
	ano := competencia[:4]
	mes := competencia[5:7]

	return fmt.Sprintf("%s/%s/%s/%s/", pb.BasePath, cnpj, ano, mes)
}

// BuildEmpresaPrefix builds prefix for listing all files of a company
func (pb *PathBuilder) BuildEmpresaPrefix(cnpj string) string {
	return fmt.Sprintf("%s/%s/", pb.BasePath, cnpj)
}
