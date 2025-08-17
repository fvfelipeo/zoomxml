package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct {
	empresaRepo *database.EmpresaRepository
	authRepo    *database.AuthRepository
	jwtSecret   []byte
}

// NewAuthService creates a new auth service
func NewAuthService(empresaRepo *database.EmpresaRepository, authRepo *database.AuthRepository, jwtSecret string) *AuthService {
	return &AuthService{
		empresaRepo: empresaRepo,
		authRepo:    authRepo,
		jwtSecret:   []byte(jwtSecret),
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(cnpj, password string) (*models.LoginResponse, error) {
	// Get empresa by CNPJ
	empresa, err := s.empresaRepo.GetByCNPJ(cnpj)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if empresa == nil || !empresa.IsActive() {
		return nil, fmt.Errorf("invalid credentials")
	}

	// For simplicity, we'll use the security_key as password
	// Try bcrypt first, if it fails, try direct comparison
	err = bcrypt.CompareHashAndPassword([]byte(empresa.SecurityKey), []byte(password))
	if err != nil {
		// If bcrypt fails, try direct comparison (for plain text passwords)
		if empresa.SecurityKey != password {
			return nil, fmt.Errorf("invalid credentials")
		}
	}

	// Generate JWT token
	token, expiresAt, err := s.generateJWT(empresa.ID, empresa.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Store token hash in database
	tokenHash := s.hashToken(token)
	_, err = s.authRepo.CreateToken(empresa.ID, tokenHash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store token: %v", err)
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Empresa:   *empresa,
	}, nil
}

// ValidateToken validates a JWT token and returns the empresa
func (s *AuthService) ValidateToken(tokenString string) (*models.Empresa, error) {
	// Parse JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	empresaID, ok := claims["empresa_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid empresa_id in token")
	}

	// Validate token in database
	tokenHash := s.hashToken(tokenString)
	authToken, err := s.authRepo.ValidateToken(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %v", err)
	}

	if authToken.EmpresaID != int(empresaID) {
		return nil, fmt.Errorf("token empresa mismatch")
	}

	// Get empresa
	empresa, err := s.empresaRepo.GetByID(int(empresaID))
	if err != nil {
		return nil, fmt.Errorf("failed to get empresa: %v", err)
	}

	if empresa == nil || !empresa.IsActive() {
		return nil, fmt.Errorf("empresa not found or inactive")
	}

	return empresa, nil
}

// Logout invalidates a token
func (s *AuthService) Logout(tokenString string) error {
	tokenHash := s.hashToken(tokenString)
	return s.authRepo.DeactivateToken(tokenHash)
}

// RefreshToken generates a new token for an existing valid token
func (s *AuthService) RefreshToken(tokenString string) (*models.LoginResponse, error) {
	empresa, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Deactivate old token
	oldTokenHash := s.hashToken(tokenString)
	err = s.authRepo.DeactivateToken(oldTokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate old token: %v", err)
	}

	// Generate new token
	newToken, expiresAt, err := s.generateJWT(empresa.ID, empresa.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token: %v", err)
	}

	// Store new token
	newTokenHash := s.hashToken(newToken)
	_, err = s.authRepo.CreateToken(empresa.ID, newTokenHash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store new token: %v", err)
	}

	return &models.LoginResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
		Empresa:   *empresa,
	}, nil
}

// generateJWT generates a JWT token
func (s *AuthService) generateJWT(empresaID int, empresaUUID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hours

	claims := jwt.MapClaims{
		"empresa_id":   float64(empresaID), // Store as float64 for JSON compatibility
		"empresa_uuid": empresaUUID,
		"exp":          expiresAt.Unix(),
		"iat":          time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// hashToken creates a hash of the token for storage
func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CleanupExpiredTokens removes expired tokens
func (s *AuthService) CleanupExpiredTokens() error {
	return s.authRepo.CleanupExpiredTokens()
}
