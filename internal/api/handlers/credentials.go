package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/api/middleware"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
	"github.com/zoomxml/internal/permissions"
)

// CredentialHandler gerencia as operações de credenciais
type CredentialHandler struct{}

// NewCredentialHandler cria uma nova instância do handler de credenciais
func NewCredentialHandler() *CredentialHandler {
	return &CredentialHandler{}
}

// CreateCredentialRequest representa a requisição para criar credencial
type CreateCredentialRequest struct {
	Type        string `json:"type" validate:"required,oneof=prefeitura_user_pass prefeitura_token prefeitura_mixed"`
	Name        string `json:"name" validate:"required,min=2,max=255"`
	Description string `json:"description,omitempty"`                                                           // Descrição opcional da credencial
	Login       string `json:"login,omitempty"`                                                                 // Para user/pass e mixed
	Password    string `json:"password,omitempty"`                                                              // Para user/pass e mixed
	Token       string `json:"token,omitempty"`                                                                 // Para token e mixed
	Environment string `json:"environment,omitempty" validate:"omitempty,oneof=production staging development"` // Ambiente
}

// UpdateCredentialRequest representa a requisição para atualizar credencial
type UpdateCredentialRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description *string `json:"description,omitempty"`
	Login       *string `json:"login,omitempty"`
	Password    *string `json:"password,omitempty"`
	Token       *string `json:"token,omitempty"`
	Environment *string `json:"environment,omitempty" validate:"omitempty,oneof=production staging development"`
	Active      *bool   `json:"active,omitempty"`
}

// CreateCredential cria uma nova credencial para uma empresa
// @Summary Criar credencial
// @Description Cria uma nova credencial para uma empresa (requer autenticação)
// @Tags credentials
// @Accept json
// @Produce json
// @Param company_id path int true "ID da empresa"
// @Param credential body CreateCredentialRequest true "Dados da credencial"
// @Success 201 {object} models.CompanyCredential
// @Failure 400 {object} SwaggerValidationError "Erro de validação"
// @Failure 401 {object} SwaggerError "Autenticação necessária"
// @Failure 403 {object} SwaggerError "Sem permissão para esta empresa"
// @Failure 404 {object} SwaggerError "Empresa não encontrada"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Security UserToken
// @Router /companies/{company_id}/credentials [post]
func (h *CredentialHandler) CreateCredential(c *fiber.Ctx) error {
	// Obter ID da empresa
	companyIDStr := c.Params("company_id")
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	// Obter usuário do contexto
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Verificar permissões do usuário para esta empresa
	err = permissions.CanCreateCredentials(c.Context(), user, companyID)
	if err != nil {
		if err == permissions.ErrCompanyNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Company not found",
			})
		}
		if err == permissions.ErrAccessDenied {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied to this company",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to validate permissions",
		})
	}

	// Parse do request
	var req CreateCredentialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validar request
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": validateStruct(req),
		})
	}

	// Criar credencial
	credential := &models.CompanyCredential{
		CompanyID:   companyID,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Login:       req.Login,
		Environment: req.Environment,
		Active:      true,
	}

	// Criptografar dados da credencial
	err = credential.SetCredentialData(req.Login, req.Password, req.Token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encrypt credential data",
		})
	}

	_, err = database.DB.NewInsert().Model(credential).Exec(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create credential",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(credential)
}

// GetCredentials lista as credenciais de uma empresa
// @Summary Listar credenciais
// @Description Lista todas as credenciais de uma empresa (requer autenticação)
// @Tags credentials
// @Produce json
// @Param company_id path int true "ID da empresa"
// @Success 200 {array} models.CompanyCredential
// @Failure 401 {object} SwaggerError "Autenticação necessária"
// @Failure 403 {object} SwaggerError "Sem permissão para esta empresa"
// @Failure 404 {object} SwaggerError "Empresa não encontrada"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Security UserToken
// @Router /companies/{company_id}/credentials [get]
func (h *CredentialHandler) GetCredentials(c *fiber.Ctx) error {
	// Obter ID da empresa
	companyIDStr := c.Params("company_id")
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	// Obter usuário do contexto
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Verificar permissões do usuário para esta empresa
	err = permissions.CanViewCredentials(c.Context(), user, companyID)
	if err != nil {
		if err == permissions.ErrCompanyNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Company not found",
			})
		}
		if err == permissions.ErrAccessDenied {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied to this company",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to validate permissions",
		})
	}

	// Buscar credenciais
	var credentials []models.CompanyCredential
	err = database.DB.NewSelect().
		Model(&credentials).
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Scan(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch credentials",
		})
	}

	return c.JSON(credentials)
}

// UpdateCredential atualiza uma credencial
// @Summary Atualizar credencial
// @Description Atualiza uma credencial existente (requer autenticação)
// @Tags credentials
// @Accept json
// @Produce json
// @Param company_id path int true "ID da empresa"
// @Param credential_id path int true "ID da credencial"
// @Param credential body UpdateCredentialRequest true "Dados para atualização"
// @Success 200 {object} models.CompanyCredential
// @Failure 400 {object} SwaggerValidationError "Erro de validação"
// @Failure 401 {object} SwaggerError "Autenticação necessária"
// @Failure 403 {object} SwaggerError "Sem permissão para esta empresa"
// @Failure 404 {object} SwaggerError "Credencial não encontrada"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Security UserToken
// @Router /companies/{company_id}/credentials/{credential_id} [patch]
func (h *CredentialHandler) UpdateCredential(c *fiber.Ctx) error {
	// Obter IDs
	companyIDStr := c.Params("company_id")
	credentialIDStr := c.Params("credential_id")

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	credentialID, err := strconv.ParseInt(credentialIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid credential ID",
		})
	}

	// Obter usuário do contexto
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Verificar permissões do usuário para esta credencial
	err = permissions.ValidateCredentialAccess(c.Context(), user, credentialID, companyID)
	if err != nil {
		if err == permissions.ErrCompanyNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Company not found",
			})
		}
		if err == permissions.ErrAccessDenied {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied to this company",
			})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Credential not found",
		})
	}

	// Buscar credencial
	credential := &models.CompanyCredential{}
	err = database.DB.NewSelect().
		Model(credential).
		Where("id = ? AND company_id = ?", credentialID, companyID).
		Scan(c.Context())

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Credential not found",
		})
	}

	// Parse do request
	var req UpdateCredentialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validar request
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": validateStruct(req),
		})
	}

	// Atualizar campos
	query := database.DB.NewUpdate().Model(credential).Where("id = ?", credentialID)

	if req.Name != nil {
		query = query.Set("name = ?", *req.Name)
		credential.Name = *req.Name
	}

	if req.Description != nil {
		query = query.Set("description = ?", *req.Description)
		credential.Description = *req.Description
	}

	if req.Login != nil {
		query = query.Set("login = ?", *req.Login)
		credential.Login = *req.Login
	}

	if req.Environment != nil {
		query = query.Set("environment = ?", *req.Environment)
		credential.Environment = *req.Environment
	}

	// Handle credential data updates
	if req.Password != nil || req.Token != nil {
		// Get current credential data
		currentLogin, currentPassword, currentToken, err := credential.GetCredentialData()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to decrypt current credential data",
			})
		}

		// Use new values if provided, otherwise keep current values
		newLogin := currentLogin
		newPassword := currentPassword
		newToken := currentToken

		if req.Login != nil {
			newLogin = *req.Login
		}
		if req.Password != nil {
			newPassword = *req.Password
		}
		if req.Token != nil {
			newToken = *req.Token
		}

		// Re-encrypt with updated data
		err = credential.SetCredentialData(newLogin, newPassword, newToken)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to encrypt updated credential data",
			})
		}

		query = query.Set("encrypted_secret = ?", credential.EncryptedSecret)
	}

	if req.Active != nil {
		query = query.Set("active = ?", *req.Active)
		credential.Active = *req.Active
	}

	// Atualizar timestamp
	query = query.Set("updated_at = CURRENT_TIMESTAMP")

	_, err = query.Exec(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update credential",
		})
	}

	return c.JSON(credential)
}

// DeleteCredential remove uma credencial
// @Summary Deletar credencial
// @Description Remove uma credencial (requer autenticação)
// @Tags credentials
// @Param company_id path int true "ID da empresa"
// @Param credential_id path int true "ID da credencial"
// @Success 204 "Credencial removida com sucesso"
// @Failure 401 {object} SwaggerError "Autenticação necessária"
// @Failure 403 {object} SwaggerError "Sem permissão para esta empresa"
// @Failure 404 {object} SwaggerError "Credencial não encontrada"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Security UserToken
// @Router /companies/{company_id}/credentials/{credential_id} [delete]
func (h *CredentialHandler) DeleteCredential(c *fiber.Ctx) error {
	// Obter IDs
	companyIDStr := c.Params("company_id")
	credentialIDStr := c.Params("credential_id")

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	credentialID, err := strconv.ParseInt(credentialIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid credential ID",
		})
	}

	// Obter usuário do contexto
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Verificar permissões do usuário para esta credencial
	err = permissions.ValidateCredentialAccess(c.Context(), user, credentialID, companyID)
	if err != nil {
		if err == permissions.ErrCompanyNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Company not found",
			})
		}
		if err == permissions.ErrAccessDenied {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied to this company",
			})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Credential not found",
		})
	}

	// Deletar credencial
	_, err = database.DB.NewDelete().
		Model((*models.CompanyCredential)(nil)).
		Where("id = ? AND company_id = ?", credentialID, companyID).
		Exec(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete credential",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
