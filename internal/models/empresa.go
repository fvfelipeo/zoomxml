package models

import (
	"encoding/json"
	"time"
)

// Empresa represents a company in the multi-tenant system
type Empresa struct {
	ID                int                    `json:"id" db:"id"`
	UUID              string                 `json:"uuid" db:"uuid"`
	CNPJ              string                 `json:"cnpj" db:"cnpj"`
	RazaoSocial       string                 `json:"razao_social" db:"razao_social"`
	NomeFantasia      string                 `json:"nome_fantasia" db:"nome_fantasia"`
	Municipio         string                 `json:"municipio" db:"municipio"`
	SecurityKey       string                 `json:"security_key" db:"security_key"`
	APIEndpoint       string                 `json:"api_endpoint" db:"api_endpoint"`
	Status            string                 `json:"status" db:"status"`
	Configuracoes     map[string]interface{} `json:"configuracoes" db:"configuracoes"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	LastSync          *time.Time             `json:"last_sync" db:"last_sync"`
	SyncIntervalHours int                    `json:"sync_interval_hours" db:"sync_interval_hours"`
	AutoSyncEnabled   bool                   `json:"auto_sync_enabled" db:"auto_sync_enabled"`
}

// EmpresaCreateRequest represents the request to create a new empresa
type EmpresaCreateRequest struct {
	CNPJ              string                 `json:"cnpj" validate:"required,len=14"`
	RazaoSocial       string                 `json:"razao_social" validate:"required,max=255"`
	NomeFantasia      string                 `json:"nome_fantasia" validate:"max=255"`
	Municipio         string                 `json:"municipio" validate:"required,max=100"`
	SecurityKey       string                 `json:"security_key" validate:"required"`
	APIEndpoint       string                 `json:"api_endpoint"`
	SyncIntervalHours int                    `json:"sync_interval_hours" validate:"min=1,max=168"`
	AutoSyncEnabled   bool                   `json:"auto_sync_enabled"`
	Configuracoes     map[string]interface{} `json:"configuracoes"`
}

// EmpresaUpdateRequest represents the request to update an empresa
type EmpresaUpdateRequest struct {
	RazaoSocial       *string                `json:"razao_social,omitempty" validate:"omitempty,max=255"`
	NomeFantasia      *string                `json:"nome_fantasia,omitempty" validate:"omitempty,max=255"`
	Municipio         *string                `json:"municipio,omitempty" validate:"omitempty,max=100"`
	SecurityKey       *string                `json:"security_key,omitempty"`
	APIEndpoint       *string                `json:"api_endpoint,omitempty"`
	Status            *string                `json:"status,omitempty" validate:"omitempty,oneof=active inactive suspended"`
	SyncIntervalHours *int                   `json:"sync_interval_hours,omitempty" validate:"omitempty,min=1,max=168"`
	AutoSyncEnabled   *bool                  `json:"auto_sync_enabled,omitempty"`
	Configuracoes     map[string]interface{} `json:"configuracoes,omitempty"`
}

// AuthToken represents a JWT authentication token
type AuthToken struct {
	ID        int        `json:"id" db:"id"`
	UUID      string     `json:"uuid" db:"uuid"`
	EmpresaID int        `json:"empresa_id" db:"empresa_id"`
	TokenHash string     `json:"-" db:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	LastUsed  *time.Time `json:"last_used" db:"last_used"`
	IsActive  bool       `json:"is_active" db:"is_active"`
}

// ProcessingJob represents a background processing job
type ProcessingJob struct {
	ID           int                    `json:"id" db:"id"`
	UUID         string                 `json:"uuid" db:"uuid"`
	EmpresaID    int                    `json:"empresa_id" db:"empresa_id"`
	JobType      string                 `json:"job_type" db:"job_type"`
	Status       string                 `json:"status" db:"status"`
	Priority     int                    `json:"priority" db:"priority"`
	ScheduledAt  time.Time              `json:"scheduled_at" db:"scheduled_at"`
	StartedAt    *time.Time             `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at" db:"completed_at"`
	Parameters   map[string]interface{} `json:"parameters" db:"parameters"`
	Result       map[string]interface{} `json:"result" db:"result"`
	ErrorMessage string                 `json:"error_message" db:"error_message"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	MaxRetries   int                    `json:"max_retries" db:"max_retries"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// SystemConfig represents system configuration
type SystemConfig struct {
	ID          int       `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	Value       string    `json:"value" db:"value"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// EmpresaStats represents statistics for a company
type EmpresaStats struct {
	TotalNFSe          int64     `json:"total_nfse"`
	TotalValorServicos float64   `json:"total_valor_servicos"`
	TotalValorISS      float64   `json:"total_valor_iss"`
	LastProcessed      time.Time `json:"last_processed"`
	TotalPrestadores   int64     `json:"total_prestadores"`
	TotalCompetencias  int64     `json:"total_competencias"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	CNPJ     string `json:"cnpj" validate:"required,len=14"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Empresa   Empresa   `json:"empresa"`
}

// JobCreateRequest represents a request to create a new job
type JobCreateRequest struct {
	EmpresaUUID string                 `json:"empresa_uuid" validate:"required,uuid"`
	JobType     string                 `json:"job_type" validate:"required,oneof=sync_nfse process_xml generate_report"`
	Priority    int                    `json:"priority" validate:"min=1,max=10"`
	ScheduledAt *time.Time             `json:"scheduled_at"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *APIMeta    `json:"meta,omitempty"`
}

// APIMeta represents metadata for paginated responses
type APIMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page    int `json:"page" query:"page" validate:"min=1"`
	PerPage int `json:"per_page" query:"per_page" validate:"min=1,max=100"`
}

// GetDefaultPagination returns default pagination values
func GetDefaultPagination() PaginationRequest {
	return PaginationRequest{
		Page:    1,
		PerPage: 20,
	}
}

// CalculateOffset calculates the offset for database queries
func (p PaginationRequest) CalculateOffset() int {
	return (p.Page - 1) * p.PerPage
}

// CalculateTotalPages calculates total pages based on total records
func (p PaginationRequest) CalculateTotalPages(total int) int {
	if total == 0 {
		return 0
	}
	return (total + p.PerPage - 1) / p.PerPage
}

// IsActive checks if empresa is active
func (e *Empresa) IsActive() bool {
	return e.Status == EmpresaStatusActive
}

// ShouldSync checks if empresa should be synced based on interval
func (e *Empresa) ShouldSync() bool {
	if !e.AutoSyncEnabled || !e.IsActive() {
		return false
	}

	if e.LastSync == nil {
		return true
	}

	nextSync := e.LastSync.Add(time.Duration(e.SyncIntervalHours) * time.Hour)
	return time.Now().After(nextSync)
}

// GetAPIEndpoint returns the API endpoint for the empresa
func (e *Empresa) GetAPIEndpoint() string {
	if e.APIEndpoint != "" {
		return e.APIEndpoint
	}
	// Default endpoint pattern
	return "https://api-nfse-" + e.Municipio + ".prefeituramoderna.com.br/ws/services"
}

// MarshalConfiguracao marshals configuracoes to JSON
func (e *Empresa) MarshalConfiguracao() ([]byte, error) {
	return json.Marshal(e.Configuracoes)
}

// UnmarshalConfiguracao unmarshals configuracoes from JSON
func (e *Empresa) UnmarshalConfiguracao(data []byte) error {
	return json.Unmarshal(data, &e.Configuracoes)
}

// IsExpired checks if auth token is expired
func (t *AuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// CanRetry checks if job can be retried
func (j *ProcessingJob) CanRetry() bool {
	return j.RetryCount < j.MaxRetries && j.Status == JobStatusFailed
}

// IsCompleted checks if job is completed
func (j *ProcessingJob) IsCompleted() bool {
	return j.Status == JobStatusCompleted || j.Status == JobStatusFailed
}

// Duration returns the job execution duration
func (j *ProcessingJob) Duration() time.Duration {
	if j.StartedAt == nil || j.CompletedAt == nil {
		return 0
	}
	return j.CompletedAt.Sub(*j.StartedAt)
}

// JobTypes constants
const (
	JobTypeSyncNFSe        = "sync_nfse"
	JobTypeProcessXML      = "process_xml"
	JobTypeGenerateReport  = "generate_report"
	JobTypeCleanup         = "cleanup"
	JobTypeXMLConsultation = "xml_consultation"
)

// Job statuses constants
const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
	JobStatusCancelled = "cancelled"
)

// Empresa statuses constants
const (
	EmpresaStatusActive    = "active"
	EmpresaStatusInactive  = "inactive"
	EmpresaStatusSuspended = "suspended"
)
