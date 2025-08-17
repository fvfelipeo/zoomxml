package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zoomxml/internal/models"
)

// EmpresaRepository handles database operations for empresas
type EmpresaRepository struct {
	db *sql.DB
}

// NewEmpresaRepository creates a new empresa repository
func NewEmpresaRepository(db *sql.DB) *EmpresaRepository {
	return &EmpresaRepository{db: db}
}

// Create creates a new empresa
func (r *EmpresaRepository) Create(empresa *models.EmpresaCreateRequest) (*models.Empresa, error) {
	// Marshal configuracoes to JSON
	configJSON, err := json.Marshal(empresa.Configuracoes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal configuracoes: %v", err)
	}

	query := `
		INSERT INTO nfse.empresas (
			cnpj, razao_social, nome_fantasia, municipio, security_key,
			api_endpoint, sync_interval_hours, auto_sync_enabled, configuracoes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, uuid, created_at, updated_at`

	var result models.Empresa
	err = r.db.QueryRow(
		query,
		empresa.CNPJ,
		empresa.RazaoSocial,
		empresa.NomeFantasia,
		empresa.Municipio,
		empresa.SecurityKey,
		empresa.APIEndpoint,
		empresa.SyncIntervalHours,
		empresa.AutoSyncEnabled,
		configJSON,
	).Scan(&result.ID, &result.UUID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create empresa: %v", err)
	}

	// Fill in the rest of the fields
	result.CNPJ = empresa.CNPJ
	result.RazaoSocial = empresa.RazaoSocial
	result.NomeFantasia = empresa.NomeFantasia
	result.Municipio = empresa.Municipio
	result.SecurityKey = empresa.SecurityKey
	result.APIEndpoint = empresa.APIEndpoint
	result.SyncIntervalHours = empresa.SyncIntervalHours
	result.AutoSyncEnabled = empresa.AutoSyncEnabled
	result.Configuracoes = empresa.Configuracoes
	result.Status = models.EmpresaStatusActive

	return &result, nil
}

// GetByID gets an empresa by ID
func (r *EmpresaRepository) GetByID(id int) (*models.Empresa, error) {
	query := `
		SELECT id, uuid, cnpj, razao_social, nome_fantasia, municipio, 
			   security_key, api_endpoint, status, configuracoes, 
			   created_at, updated_at, last_sync, sync_interval_hours, auto_sync_enabled
		FROM nfse.empresas WHERE id = $1`

	return r.scanEmpresa(r.db.QueryRow(query, id))
}

// GetByUUID gets an empresa by UUID
func (r *EmpresaRepository) GetByUUID(uuid string) (*models.Empresa, error) {
	query := `
		SELECT id, uuid, cnpj, razao_social, nome_fantasia, municipio, 
			   security_key, api_endpoint, status, configuracoes, 
			   created_at, updated_at, last_sync, sync_interval_hours, auto_sync_enabled
		FROM nfse.empresas WHERE uuid = $1`

	return r.scanEmpresa(r.db.QueryRow(query, uuid))
}

// GetByCNPJ gets an empresa by CNPJ
func (r *EmpresaRepository) GetByCNPJ(cnpj string) (*models.Empresa, error) {
	query := `
		SELECT id, uuid, cnpj, razao_social, nome_fantasia, municipio, 
			   security_key, api_endpoint, status, configuracoes, 
			   created_at, updated_at, last_sync, sync_interval_hours, auto_sync_enabled
		FROM nfse.empresas WHERE cnpj = $1`

	return r.scanEmpresa(r.db.QueryRow(query, cnpj))
}

// List lists empresas with pagination
func (r *EmpresaRepository) List(pagination models.PaginationRequest, status string) ([]models.Empresa, int, error) {
	// Count total
	countQuery := "SELECT COUNT(*) FROM nfse.empresas"
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		countQuery += " WHERE status = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count empresas: %v", err)
	}

	// Get paginated results
	query := `
		SELECT id, uuid, cnpj, razao_social, nome_fantasia, municipio, 
			   security_key, api_endpoint, status, configuracoes, 
			   created_at, updated_at, last_sync, sync_interval_hours, auto_sync_enabled
		FROM nfse.empresas`

	if status != "" {
		query += " WHERE status = $1"
		query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
		args = []interface{}{status, pagination.PerPage, pagination.CalculateOffset()}
	} else {
		query += " ORDER BY created_at DESC LIMIT $1 OFFSET $2"
		args = []interface{}{pagination.PerPage, pagination.CalculateOffset()}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list empresas: %v", err)
	}
	defer rows.Close()

	var empresas []models.Empresa
	for rows.Next() {
		empresa, err := r.scanEmpresa(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan empresa: %v", err)
		}
		empresas = append(empresas, *empresa)
	}

	return empresas, total, nil
}

// Update updates an empresa
func (r *EmpresaRepository) Update(id int, updates *models.EmpresaUpdateRequest) (*models.Empresa, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.RazaoSocial != nil {
		setParts = append(setParts, fmt.Sprintf("razao_social = $%d", argIndex))
		args = append(args, *updates.RazaoSocial)
		argIndex++
	}

	if updates.NomeFantasia != nil {
		setParts = append(setParts, fmt.Sprintf("nome_fantasia = $%d", argIndex))
		args = append(args, *updates.NomeFantasia)
		argIndex++
	}

	if updates.Municipio != nil {
		setParts = append(setParts, fmt.Sprintf("municipio = $%d", argIndex))
		args = append(args, *updates.Municipio)
		argIndex++
	}

	if updates.SecurityKey != nil {
		setParts = append(setParts, fmt.Sprintf("security_key = $%d", argIndex))
		args = append(args, *updates.SecurityKey)
		argIndex++
	}

	if updates.APIEndpoint != nil {
		setParts = append(setParts, fmt.Sprintf("api_endpoint = $%d", argIndex))
		args = append(args, *updates.APIEndpoint)
		argIndex++
	}

	if updates.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *updates.Status)
		argIndex++
	}

	if updates.SyncIntervalHours != nil {
		setParts = append(setParts, fmt.Sprintf("sync_interval_hours = $%d", argIndex))
		args = append(args, *updates.SyncIntervalHours)
		argIndex++
	}

	if updates.AutoSyncEnabled != nil {
		setParts = append(setParts, fmt.Sprintf("auto_sync_enabled = $%d", argIndex))
		args = append(args, *updates.AutoSyncEnabled)
		argIndex++
	}

	if updates.Configuracoes != nil {
		configJSON, err := json.Marshal(updates.Configuracoes)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal configuracoes: %v", err)
		}
		setParts = append(setParts, fmt.Sprintf("configuracoes = $%d", argIndex))
		args = append(args, configJSON)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByID(id) // No updates, return current
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause
	args = append(args, id)

	query := fmt.Sprintf("UPDATE nfse.empresas SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update empresa: %v", err)
	}

	return r.GetByID(id)
}

// Delete deletes an empresa (soft delete by setting status to inactive)
func (r *EmpresaRepository) Delete(id int) error {
	query := "UPDATE nfse.empresas SET status = $1, updated_at = $2 WHERE id = $3"
	_, err := r.db.Exec(query, models.EmpresaStatusInactive, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete empresa: %v", err)
	}
	return nil
}

// UpdateLastSync updates the last sync time for an empresa
func (r *EmpresaRepository) UpdateLastSync(id int, lastSync time.Time) error {
	query := "UPDATE nfse.empresas SET last_sync = $1, updated_at = $2 WHERE id = $3"
	_, err := r.db.Exec(query, lastSync, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last sync: %v", err)
	}
	return nil
}

// GetEmpresasForSync gets empresas that should be synced
func (r *EmpresaRepository) GetEmpresasForSync() ([]models.Empresa, error) {
	query := `
		SELECT id, uuid, cnpj, razao_social, nome_fantasia, municipio, 
			   security_key, api_endpoint, status, configuracoes, 
			   created_at, updated_at, last_sync, sync_interval_hours, auto_sync_enabled
		FROM nfse.empresas
		WHERE status = $1 AND auto_sync_enabled = true
		AND (last_sync IS NULL OR last_sync + (sync_interval_hours * INTERVAL '1 hour') <= NOW())`

	rows, err := r.db.Query(query, models.EmpresaStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get empresas for sync: %v", err)
	}
	defer rows.Close()

	var empresas []models.Empresa
	for rows.Next() {
		empresa, err := r.scanEmpresa(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan empresa: %v", err)
		}
		empresas = append(empresas, *empresa)
	}

	return empresas, nil
}

// scanEmpresa scans a row into an Empresa struct
func (r *EmpresaRepository) scanEmpresa(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Empresa, error) {
	var empresa models.Empresa
	var configJSON []byte
	var lastSync sql.NullTime
	var apiEndpoint sql.NullString
	var nomeFantasia sql.NullString

	err := scanner.Scan(
		&empresa.ID,
		&empresa.UUID,
		&empresa.CNPJ,
		&empresa.RazaoSocial,
		&nomeFantasia,
		&empresa.Municipio,
		&empresa.SecurityKey,
		&apiEndpoint,
		&empresa.Status,
		&configJSON,
		&empresa.CreatedAt,
		&empresa.UpdatedAt,
		&lastSync,
		&empresa.SyncIntervalHours,
		&empresa.AutoSyncEnabled,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan empresa: %v", err)
	}

	// Handle nullable fields
	if lastSync.Valid {
		empresa.LastSync = &lastSync.Time
	}
	if apiEndpoint.Valid {
		empresa.APIEndpoint = apiEndpoint.String
	}
	if nomeFantasia.Valid {
		empresa.NomeFantasia = nomeFantasia.String
	}

	// Unmarshal configuracoes
	if len(configJSON) > 0 {
		err = json.Unmarshal(configJSON, &empresa.Configuracoes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal configuracoes: %v", err)
		}
	}

	return &empresa, nil
}
