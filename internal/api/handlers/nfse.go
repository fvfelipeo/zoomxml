package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/api/middleware"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/logger"
	"github.com/zoomxml/internal/models"
	"github.com/zoomxml/internal/permissions"
	"github.com/zoomxml/internal/services"
)

// NFSeHandler handles NFSe-related HTTP requests
type NFSeHandler struct {
	nfseService *services.NFSeService
}

// NewNFSeHandler creates a new NFSe handler
func NewNFSeHandler() *NFSeHandler {
	return &NFSeHandler{
		nfseService: services.NewNFSeService(),
	}
}

// FetchNFSeRequest represents the request to fetch NFSe documents
type FetchNFSeRequest struct {
	StartDate string `json:"start_date" validate:"required"` // Format: 2006-01-02
	EndDate   string `json:"end_date" validate:"required"`   // Format: 2006-01-02
	Page      int    `json:"page,omitempty"`                 // Page number (default: 1)
}

// FetchNFSeResponse represents the response from fetching NFSe documents
type FetchNFSeResponse struct {
	Success        bool                    `json:"success"`
	Message        string                  `json:"message"`
	DocumentsCount int                     `json:"documents_count"`
	Documents      []services.NFSeDocument `json:"documents,omitempty"`
	Error          string                  `json:"error,omitempty"`
}

// FetchNFSeDocuments fetches NFSe documents for a company
// @Summary Fetch NFSe documents
// @Description Fetches NFSe documents from the municipal API for a specific company
// @Tags nfse
// @Accept json
// @Produce json
// @Param company_id path int true "Company ID"
// @Param request body FetchNFSeRequest true "Fetch request"
// @Success 200 {object} FetchNFSeResponse
// @Failure 400 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Failure 403 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/companies/{company_id}/nfse/fetch [post]
func (h *NFSeHandler) FetchNFSeDocuments(c *fiber.Ctx) error {
	// Parse company ID
	companyIDStr := c.Params("company_id")
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	// Get user from context
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Check permissions
	err = permissions.CanAccessCompany(c.Context(), user, companyID)
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

	// Parse request body
	var req FetchNFSeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set default page if not provided
	if req.Page <= 0 {
		req.Page = 1
	}

	// Validate request
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": validateStruct(req),
		})
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start_date format. Use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end_date format. Use YYYY-MM-DD",
		})
	}

	// Validate date range
	if endDate.Before(startDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "End date must be after start date",
		})
	}

	// Find company credentials for NFSe
	credentials := []models.CompanyCredential{}
	err = database.DB.NewSelect().
		Model(&credentials).
		Where("company_id = ? AND active = true", companyID).
		Where("type IN ('prefeitura_token', 'prefeitura_mixed')").
		Scan(c.Context())

	if err != nil {
		logger.ErrorWithFields("Failed to fetch company credentials", err, map[string]any{
			"operation":  "fetch_nfse",
			"company_id": companyID,
			"user_id":    user.ID,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch company credentials",
		})
	}

	if len(credentials) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No NFSe credentials found for this company",
		})
	}

	// Use the first available credential
	credential := &credentials[0]

	logger.InfoWithFields("Starting NFSe fetch", map[string]any{
		"operation":     "fetch_nfse",
		"company_id":    companyID,
		"user_id":       user.ID,
		"credential_id": credential.ID,
		"start_date":    req.StartDate,
		"end_date":      req.EndDate,
	})

	// Fetch NFSe documents
	nfseResponse, err := h.nfseService.FetchNFSeDocuments(c.Context(), credential, startDate, endDate, req.Page)
	if err != nil {
		logger.ErrorWithFields("Failed to fetch NFSe documents", err, map[string]any{
			"operation":     "fetch_nfse",
			"company_id":    companyID,
			"user_id":       user.ID,
			"credential_id": credential.ID,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(FetchNFSeResponse{
			Success: false,
			Message: "Failed to fetch NFSe documents",
			Error:   err.Error(),
		})
	}

	// Store documents if successful
	if nfseResponse.Success && len(nfseResponse.Documents) > 0 {
		err = h.nfseService.StoreNFSeDocuments(c.Context(), companyID, nfseResponse.Documents)
		if err != nil {
			logger.ErrorWithFields("Failed to store NFSe documents", err, map[string]any{
				"operation":  "fetch_nfse",
				"company_id": companyID,
				"user_id":    user.ID,
			})
			// Don't return error here, just log it - we still want to return the fetched data
		}
	}

	logger.InfoWithFields("NFSe fetch completed", map[string]any{
		"operation":       "fetch_nfse",
		"company_id":      companyID,
		"user_id":         user.ID,
		"documents_count": len(nfseResponse.Documents),
		"success":         nfseResponse.Success,
	})

	return c.Status(fiber.StatusOK).JSON(FetchNFSeResponse{
		Success:        nfseResponse.Success,
		Message:        nfseResponse.Message,
		DocumentsCount: len(nfseResponse.Documents),
		Documents:      nfseResponse.Documents,
		Error:          nfseResponse.Error,
	})
}

// GetNFSeDocuments lists stored NFSe documents for a company
// @Summary List NFSe documents
// @Description Lists stored NFSe documents for a specific company
// @Tags nfse
// @Accept json
// @Produce json
// @Param company_id path int true "Company ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Failure 403 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/companies/{company_id}/nfse [get]
func (h *NFSeHandler) GetNFSeDocuments(c *fiber.Ctx) error {
	// Parse company ID
	companyIDStr := c.Params("company_id")
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	// Get user from context
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Check permissions
	err = permissions.CanAccessCompany(c.Context(), user, companyID)
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

	// Parse pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	offset := (page - 1) * limit

	// Fetch documents
	documents := []models.Document{}
	err = database.DB.NewSelect().
		Model(&documents).
		Where("company_id = ? AND type = 'nfse'", companyID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(c.Context())

	if err != nil {
		logger.ErrorWithFields("Failed to fetch NFSe documents", err, map[string]any{
			"operation":  "get_nfse_documents",
			"company_id": companyID,
			"user_id":    user.ID,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch documents",
		})
	}

	// Count total documents
	total, err := database.DB.NewSelect().
		Model((*models.Document)(nil)).
		Where("company_id = ? AND type = 'nfse'", companyID).
		Count(c.Context())

	if err != nil {
		logger.ErrorWithFields("Failed to count NFSe documents", err, map[string]any{
			"operation":  "get_nfse_documents",
			"company_id": companyID,
			"user_id":    user.ID,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count documents",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"documents": documents,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
