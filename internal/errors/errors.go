package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Credential-related errors
var (
	ErrCredentialNotFound = &AppError{
		Code:       "CREDENTIAL_NOT_FOUND",
		Message:    "Credencial não encontrada",
		HTTPStatus: http.StatusNotFound,
	}

	ErrCredentialInvalidType = &AppError{
		Code:       "CREDENTIAL_INVALID_TYPE",
		Message:    "Tipo de credencial inválido",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrCredentialEncryptionFailed = &AppError{
		Code:       "CREDENTIAL_ENCRYPTION_FAILED",
		Message:    "Falha ao criptografar dados da credencial",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrCredentialDecryptionFailed = &AppError{
		Code:       "CREDENTIAL_DECRYPTION_FAILED",
		Message:    "Falha ao descriptografar dados da credencial",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrCredentialInvalidData = &AppError{
		Code:       "CREDENTIAL_INVALID_DATA",
		Message:    "Dados da credencial inválidos",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrCredentialAlreadyExists = &AppError{
		Code:       "CREDENTIAL_ALREADY_EXISTS",
		Message:    "Credencial já existe para esta empresa",
		HTTPStatus: http.StatusConflict,
	}
)

// Company-related errors
var (
	ErrCompanyNotFound = &AppError{
		Code:       "COMPANY_NOT_FOUND",
		Message:    "Empresa não encontrada",
		HTTPStatus: http.StatusNotFound,
	}

	ErrCompanyAccessDenied = &AppError{
		Code:       "COMPANY_ACCESS_DENIED",
		Message:    "Acesso negado a esta empresa",
		HTTPStatus: http.StatusForbidden,
	}

	ErrCompanyInactive = &AppError{
		Code:       "COMPANY_INACTIVE",
		Message:    "Empresa inativa",
		HTTPStatus: http.StatusForbidden,
	}
)

// Authentication and authorization errors
var (
	ErrAuthenticationRequired = &AppError{
		Code:       "AUTHENTICATION_REQUIRED",
		Message:    "Autenticação necessária",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "Token inválido",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrTokenExpired = &AppError{
		Code:       "TOKEN_EXPIRED",
		Message:    "Token expirado",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInsufficientPermissions = &AppError{
		Code:       "INSUFFICIENT_PERMISSIONS",
		Message:    "Permissões insuficientes",
		HTTPStatus: http.StatusForbidden,
	}

	ErrUserNotFound = &AppError{
		Code:       "USER_NOT_FOUND",
		Message:    "Usuário não encontrado",
		HTTPStatus: http.StatusNotFound,
	}

	ErrUserInactive = &AppError{
		Code:       "USER_INACTIVE",
		Message:    "Usuário inativo",
		HTTPStatus: http.StatusForbidden,
	}
)

// Validation errors
var (
	ErrValidationFailed = &AppError{
		Code:       "VALIDATION_FAILED",
		Message:    "Falha na validação dos dados",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidRequestBody = &AppError{
		Code:       "INVALID_REQUEST_BODY",
		Message:    "Corpo da requisição inválido",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidParameter = &AppError{
		Code:       "INVALID_PARAMETER",
		Message:    "Parâmetro inválido",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrMissingRequiredField = &AppError{
		Code:       "MISSING_REQUIRED_FIELD",
		Message:    "Campo obrigatório ausente",
		HTTPStatus: http.StatusBadRequest,
	}
)

// Database errors
var (
	ErrDatabaseConnection = &AppError{
		Code:       "DATABASE_CONNECTION_FAILED",
		Message:    "Falha na conexão com o banco de dados",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrDatabaseQuery = &AppError{
		Code:       "DATABASE_QUERY_FAILED",
		Message:    "Falha na consulta ao banco de dados",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrDatabaseTransaction = &AppError{
		Code:       "DATABASE_TRANSACTION_FAILED",
		Message:    "Falha na transação do banco de dados",
		HTTPStatus: http.StatusInternalServerError,
	}
)

// Generic errors
var (
	ErrInternalServer = &AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "Erro interno do servidor",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrServiceUnavailable = &AppError{
		Code:       "SERVICE_UNAVAILABLE",
		Message:    "Serviço indisponível",
		HTTPStatus: http.StatusServiceUnavailable,
	}

	ErrRateLimitExceeded = &AppError{
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    "Limite de requisições excedido",
		HTTPStatus: http.StatusTooManyRequests,
	}
)

// Helper functions to create custom errors with details

// NewCredentialError creates a new credential-related error with details
func NewCredentialError(baseError *AppError, details string) *AppError {
	return &AppError{
		Code:       baseError.Code,
		Message:    baseError.Message,
		Details:    details,
		HTTPStatus: baseError.HTTPStatus,
	}
}

// NewValidationError creates a new validation error with details
func NewValidationError(details string) *AppError {
	return &AppError{
		Code:       ErrValidationFailed.Code,
		Message:    ErrValidationFailed.Message,
		Details:    details,
		HTTPStatus: ErrValidationFailed.HTTPStatus,
	}
}

// NewDatabaseError creates a new database error with details
func NewDatabaseError(baseError *AppError, details string) *AppError {
	return &AppError{
		Code:       baseError.Code,
		Message:    baseError.Message,
		Details:    details,
		HTTPStatus: baseError.HTTPStatus,
	}
}

// NewPermissionError creates a new permission error with details
func NewPermissionError(details string) *AppError {
	return &AppError{
		Code:       ErrInsufficientPermissions.Code,
		Message:    ErrInsufficientPermissions.Message,
		Details:    details,
		HTTPStatus: ErrInsufficientPermissions.HTTPStatus,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// WrapError wraps a generic error into an AppError
func WrapError(err error, baseError *AppError) *AppError {
	return &AppError{
		Code:       baseError.Code,
		Message:    baseError.Message,
		Details:    err.Error(),
		HTTPStatus: baseError.HTTPStatus,
	}
}
