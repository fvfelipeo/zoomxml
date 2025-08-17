package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/zoomxml/internal/models"
)

// AuthRepository handles database operations for authentication
type AuthRepository struct {
	db *sql.DB
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// CreateToken creates a new auth token
func (r *AuthRepository) CreateToken(empresaID int, tokenHash string, expiresAt time.Time) (*models.AuthToken, error) {
	query := `
		INSERT INTO nfse.auth_tokens (empresa_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, uuid, created_at`

	var token models.AuthToken
	err := r.db.QueryRow(query, empresaID, tokenHash, expiresAt).Scan(
		&token.ID, &token.UUID, &token.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth token: %v", err)
	}

	token.EmpresaID = empresaID
	token.TokenHash = tokenHash
	token.ExpiresAt = expiresAt
	token.IsActive = true

	return &token, nil
}

// GetTokenByHash gets a token by its hash
func (r *AuthRepository) GetTokenByHash(tokenHash string) (*models.AuthToken, error) {
	query := `
		SELECT id, uuid, empresa_id, token_hash, expires_at, created_at, last_used, is_active
		FROM nfse.auth_tokens
		WHERE token_hash = $1 AND is_active = true`

	return r.scanToken(r.db.QueryRow(query, tokenHash))
}

// GetTokensByEmpresa gets all active tokens for an empresa
func (r *AuthRepository) GetTokensByEmpresa(empresaID int) ([]models.AuthToken, error) {
	query := `
		SELECT id, uuid, empresa_id, token_hash, expires_at, created_at, last_used, is_active
		FROM nfse.auth_tokens
		WHERE empresa_id = $1 AND is_active = true
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, empresaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens by empresa: %v", err)
	}
	defer rows.Close()

	var tokens []models.AuthToken
	for rows.Next() {
		token, err := r.scanToken(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token: %v", err)
		}
		tokens = append(tokens, *token)
	}

	return tokens, nil
}

// UpdateLastUsed updates the last used time for a token
func (r *AuthRepository) UpdateLastUsed(tokenHash string) error {
	query := "UPDATE nfse.auth_tokens SET last_used = $1 WHERE token_hash = $2"
	_, err := r.db.Exec(query, time.Now(), tokenHash)
	if err != nil {
		return fmt.Errorf("failed to update last used: %v", err)
	}
	return nil
}

// DeactivateToken deactivates a token
func (r *AuthRepository) DeactivateToken(tokenHash string) error {
	query := "UPDATE nfse.auth_tokens SET is_active = false WHERE token_hash = $1"
	_, err := r.db.Exec(query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to deactivate token: %v", err)
	}
	return nil
}

// DeactivateAllTokensForEmpresa deactivates all tokens for an empresa
func (r *AuthRepository) DeactivateAllTokensForEmpresa(empresaID int) error {
	query := "UPDATE nfse.auth_tokens SET is_active = false WHERE empresa_id = $1"
	_, err := r.db.Exec(query, empresaID)
	if err != nil {
		return fmt.Errorf("failed to deactivate all tokens: %v", err)
	}
	return nil
}

// CleanupExpiredTokens removes expired tokens
func (r *AuthRepository) CleanupExpiredTokens() error {
	query := "DELETE FROM nfse.auth_tokens WHERE expires_at < $1"
	_, err := r.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %v", err)
	}
	return nil
}

// ValidateToken validates a token and returns the associated empresa
func (r *AuthRepository) ValidateToken(tokenHash string) (*models.AuthToken, error) {
	token, err := r.GetTokenByHash(tokenHash)
	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, fmt.Errorf("token not found")
	}

	if token.IsExpired() {
		return nil, fmt.Errorf("token expired")
	}

	if !token.IsActive {
		return nil, fmt.Errorf("token inactive")
	}

	// Update last used
	err = r.UpdateLastUsed(tokenHash)
	if err != nil {
		// Log error but don't fail validation
		fmt.Printf("Warning: failed to update last used for token: %v\n", err)
	}

	return token, nil
}

// GetActiveTokenCount gets the count of active tokens for an empresa
func (r *AuthRepository) GetActiveTokenCount(empresaID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM nfse.auth_tokens
		WHERE empresa_id = $1 AND is_active = true AND expires_at > $2`

	var count int
	err := r.db.QueryRow(query, empresaID, time.Now()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get active token count: %v", err)
	}

	return count, nil
}

// scanToken scans a row into an AuthToken struct
func (r *AuthRepository) scanToken(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.AuthToken, error) {
	var token models.AuthToken
	var lastUsed sql.NullTime

	err := scanner.Scan(
		&token.ID,
		&token.UUID,
		&token.EmpresaID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&lastUsed,
		&token.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan token: %v", err)
	}

	// Handle nullable last_used
	if lastUsed.Valid {
		token.LastUsed = &lastUsed.Time
	}

	return &token, nil
}
