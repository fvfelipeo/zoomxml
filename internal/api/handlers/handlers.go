package handlers

import (
	"fmt"

	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/services"
	"github.com/zoomxml/internal/storage"
)

// HandlerContainer holds all application handlers
type HandlerContainer struct {
	// Authentication handlers
	Auth *AuthHandler

	// Business entity handlers
	Empresa *EmpresaHandler

	// NFS-e specific handlers
	NFSe *NFSeHandler

	// Job management handlers
	Job *JobHandler

	// Statistics handlers
	Stats *StatsHandler
}

// HandlerDependencies holds all dependencies needed to create handlers
type HandlerDependencies struct {
	// Services
	AuthService *services.AuthService

	// Repositories
	EmpresaRepo *database.EmpresaRepository
	AuthRepo    *database.AuthRepository
	JobRepo     *database.JobRepository

	// Storage
	StorageProvider *storage.MinIOProvider
	NFSeManager     *storage.NFSeMinIOManager
}

// NewHandlerContainer creates a new container with all handlers
func NewHandlerContainer(deps HandlerDependencies) *HandlerContainer {
	return &HandlerContainer{
		Auth:    NewAuthHandler(deps.AuthService),
		Empresa: NewEmpresaHandler(deps.EmpresaRepo),
		NFSe: NewNFSeHandler(
			deps.JobRepo,
			deps.EmpresaRepo,
			deps.StorageProvider,
			deps.NFSeManager,
		),
		Job: NewJobHandler(deps.JobRepo),
		Stats: NewStatsHandler(
			deps.EmpresaRepo,
			deps.JobRepo,
			deps.StorageProvider,
		),
	}
}

// HandlerConfig provides configuration for handlers
type HandlerConfig struct {
	// Rate limiting
	EnableRateLimit bool
	RateLimitRPM    int

	// Timeouts
	RequestTimeout  int // seconds
	UploadTimeout   int // seconds
	DownloadTimeout int // seconds

	// File limits
	MaxFileSize  int64 // bytes
	MaxBatchSize int   // number of files
	AllowedTypes []string

	// Security
	EnableCSRF     bool
	EnableCORS     bool
	TrustedOrigins []string

	// Logging
	EnableRequestLogging  bool
	EnableResponseLogging bool
	LogLevel              string
}

// GetDefaultHandlerConfig returns default configuration
func GetDefaultHandlerConfig() HandlerConfig {
	return HandlerConfig{
		EnableRateLimit: true,
		RateLimitRPM:    1000,

		RequestTimeout:  30,
		UploadTimeout:   300,
		DownloadTimeout: 60,

		MaxFileSize:  10 * 1024 * 1024, // 10MB
		MaxBatchSize: 100,
		AllowedTypes: []string{"application/xml", "application/zip", "text/xml"},

		EnableCSRF:     true,
		EnableCORS:     true,
		TrustedOrigins: []string{"http://localhost:3000", "https://app.zoomxml.com"},

		EnableRequestLogging:  true,
		EnableResponseLogging: false,
		LogLevel:              "info",
	}
}

// HandlerMetrics provides metrics for monitoring
type HandlerMetrics struct {
	// Request metrics
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64

	// Response time metrics
	AvgResponseTime float64
	MinResponseTime float64
	MaxResponseTime float64

	// Error metrics
	ErrorRate     float64
	LastError     string
	LastErrorTime string

	// Handler-specific metrics
	AuthRequests    int64
	EmpresaRequests int64
	NFSeRequests    int64
	JobRequests     int64
	StatsRequests   int64
}

// GetHandlerMetrics returns current handler metrics
func (hc *HandlerContainer) GetHandlerMetrics() HandlerMetrics {
	// TODO: Implement actual metrics collection
	return HandlerMetrics{
		TotalRequests:      0,
		SuccessfulRequests: 0,
		FailedRequests:     0,
		AvgResponseTime:    0.0,
		MinResponseTime:    0.0,
		MaxResponseTime:    0.0,
		ErrorRate:          0.0,
		LastError:          "",
		LastErrorTime:      "",
		AuthRequests:       0,
		EmpresaRequests:    0,
		NFSeRequests:       0,
		JobRequests:        0,
		StatsRequests:      0,
	}
}

// ValidateHandlerDependencies validates that all required dependencies are provided
func ValidateHandlerDependencies(deps HandlerDependencies) error {
	if deps.AuthService == nil {
		return fmt.Errorf("AuthService is required")
	}
	if deps.EmpresaRepo == nil {
		return fmt.Errorf("EmpresaRepo is required")
	}
	if deps.AuthRepo == nil {
		return fmt.Errorf("AuthRepo is required")
	}
	if deps.JobRepo == nil {
		return fmt.Errorf("JobRepo is required")
	}
	if deps.StorageProvider == nil {
		return fmt.Errorf("StorageProvider is required")
	}
	if deps.NFSeManager == nil {
		return fmt.Errorf("NFSeManager is required")
	}
	return nil
}

// HandlerStatus represents the status of a handler
type HandlerStatus struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	LastRequest string `json:"last_request"`
	ErrorCount  int64  `json:"error_count"`
	Uptime      string `json:"uptime"`
}

// GetHandlerStatus returns status of all handlers
func (hc *HandlerContainer) GetHandlerStatus() []HandlerStatus {
	// TODO: Implement actual status tracking
	return []HandlerStatus{
		{
			Name:        "AuthHandler",
			Status:      "healthy",
			LastRequest: "2025-01-17T10:30:00Z",
			ErrorCount:  0,
			Uptime:      "99.9%",
		},
		{
			Name:        "EmpresaHandler",
			Status:      "healthy",
			LastRequest: "2025-01-17T10:29:45Z",
			ErrorCount:  0,
			Uptime:      "99.9%",
		},
		{
			Name:        "NFSeHandler",
			Status:      "healthy",
			LastRequest: "2025-01-17T10:29:30Z",
			ErrorCount:  0,
			Uptime:      "99.9%",
		},
		{
			Name:        "JobHandler",
			Status:      "healthy",
			LastRequest: "2025-01-17T10:29:15Z",
			ErrorCount:  0,
			Uptime:      "99.9%",
		},
		{
			Name:        "StatsHandler",
			Status:      "healthy",
			LastRequest: "2025-01-17T10:29:00Z",
			ErrorCount:  0,
			Uptime:      "99.9%",
		},
	}
}
