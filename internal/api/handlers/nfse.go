package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/api/middleware"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
	"github.com/zoomxml/internal/storage"
)

// NFSeHandler handles NFS-e related operations
type NFSeHandler struct {
	jobRepo         *database.JobRepository
	empresaRepo     *database.EmpresaRepository
	storageProvider *storage.MinIOProvider
	nfseManager     *storage.NFSeMinIOManager
}

// NewNFSeHandler creates a new NFS-e handler
func NewNFSeHandler(
	jobRepo *database.JobRepository,
	empresaRepo *database.EmpresaRepository,
	storageProvider *storage.MinIOProvider,
	nfseManager *storage.NFSeMinIOManager,
) *NFSeHandler {
	return &NFSeHandler{
		jobRepo:         jobRepo,
		empresaRepo:     empresaRepo,
		storageProvider: storageProvider,
		nfseManager:     nfseManager,
	}
}

// HandleManualSync handles manual sync trigger
func (h *NFSeHandler) HandleManualSync(c *fiber.Ctx) error {
	empresaID := middleware.GetEmpresaIDFromContext(c)
	if empresaID == 0 {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	// Create manual sync job
	parameters := map[string]interface{}{
		"empresa_id": float64(empresaID), // Convert to float64 for JSON compatibility
		"sync_type":  "manual",
	}

	job, err := h.jobRepo.Create(empresaID, models.JobTypeSyncNFSe, 10, time.Now(), parameters)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "Failed to create sync job",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Sync job created successfully",
		Data:    job,
	})
}

// HandleListJobs handles listing jobs for an empresa
func (h *NFSeHandler) HandleListJobs(c *fiber.Ctx) error {
	empresaID := middleware.GetEmpresaIDFromContext(c)
	if empresaID == 0 {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	// Parse pagination
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	status := c.Query("status")
	jobType := c.Query("job_type")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	pagination := models.PaginationRequest{
		Page:    page,
		PerPage: perPage,
	}

	jobs, total, err := h.jobRepo.List(pagination, &empresaID, status, jobType)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	meta := &models.APIMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: pagination.CalculateTotalPages(total),
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    jobs,
		Meta:    meta,
	})
}

// HandleGetStats handles getting stats for an empresa
func (h *NFSeHandler) HandleGetStats(c *fiber.Ctx) error {
	empresaUUID := middleware.GetEmpresaUUIDFromContext(c)
	if empresaUUID == "" {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	// TODO: Implement actual stats calculation
	// This would query the database for NFS-e statistics

	stats := map[string]interface{}{
		"total_nfse":           0,
		"total_valor_servicos": 0.0,
		"total_valor_iss":      0.0,
		"last_processed":       time.Now(),
		"total_prestadores":    0,
		"total_competencias":   0,
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    stats,
	})
}

// HandleListStoredXMLs lists all stored XMLs for the authenticated empresa
func (h *NFSeHandler) HandleListStoredXMLs(c *fiber.Ctx) error {
	empresaUUID := middleware.GetEmpresaUUIDFromContext(c)
	empresa := middleware.GetEmpresaFromContext(c)
	if empresaUUID == "" || empresa == nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	// Parse pagination
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// List all XMLs for this empresa
	prefix := empresa.CNPJ + "/"
	files, err := h.storageProvider.List(c.Context(), prefix)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "Failed to list XMLs",
		})
	}

	// Filter only XML files
	xmlFiles := []storage.FileInfo{}
	for _, file := range files {
		if strings.HasSuffix(file.Path, ".xml") && strings.Contains(file.Path, "/xml/") {
			xmlFiles = append(xmlFiles, file)
		}
	}

	// Apply pagination
	start := (page - 1) * perPage
	end := start + perPage
	if start > len(xmlFiles) {
		start = len(xmlFiles)
	}
	if end > len(xmlFiles) {
		end = len(xmlFiles)
	}

	paginatedFiles := xmlFiles[start:end]

	meta := &models.APIMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      len(xmlFiles),
		TotalPages: (len(xmlFiles) + perPage - 1) / perPage,
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    paginatedFiles,
		Meta:    meta,
	})
}

// HandleListXMLsByCompetencia lists XMLs by competencia for the authenticated empresa
func (h *NFSeHandler) HandleListXMLsByCompetencia(c *fiber.Ctx) error {
	empresaUUID := middleware.GetEmpresaUUIDFromContext(c)
	empresa := middleware.GetEmpresaFromContext(c)
	competencia := c.Params("competencia")

	if empresaUUID == "" || empresa == nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	if competencia == "" {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Competencia is required",
		})
	}

	// Build path for competencia
	ano := competencia[:4]
	mes := competencia[5:7]
	prefix := fmt.Sprintf("%s/%s/%s/xml/", empresa.CNPJ, ano, mes)

	files, err := h.storageProvider.List(c.Context(), prefix)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "Failed to list XMLs for competencia",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    files,
		Meta: &models.APIMeta{
			Page:       1,
			PerPage:    len(files),
			Total:      len(files),
			TotalPages: 1,
		},
	})
}

// HandleGetStoredXML gets a specific stored XML
func (h *NFSeHandler) HandleGetStoredXML(c *fiber.Ctx) error {
	empresaUUID := middleware.GetEmpresaUUIDFromContext(c)
	empresa := middleware.GetEmpresaFromContext(c)
	competencia := c.Params("competencia")
	numero := c.Params("numero")

	if empresaUUID == "" || empresa == nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	if competencia == "" || numero == "" {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Competencia and numero are required",
		})
	}

	// Try to get XML from storage
	xmlData, err := h.nfseManager.GetXML(c.Context(), empresa.CNPJ, competencia, numero)
	if err != nil {
		return c.Status(404).JSON(models.APIResponse{
			Success: false,
			Error:   "XML not found",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"competencia": competencia,
			"numero":      numero,
			"xml_content": string(xmlData),
			"size":        len(xmlData),
		},
	})
}

// HandleDownloadXML downloads a specific XML file
func (h *NFSeHandler) HandleDownloadXML(c *fiber.Ctx) error {
	empresaUUID := middleware.GetEmpresaUUIDFromContext(c)
	empresa := middleware.GetEmpresaFromContext(c)
	competencia := c.Params("competencia")
	numero := c.Params("numero")

	if empresaUUID == "" || empresa == nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	if competencia == "" || numero == "" {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Competencia and numero are required",
		})
	}

	// Try to get XML from storage
	xmlData, err := h.nfseManager.GetXML(c.Context(), empresa.CNPJ, competencia, numero)
	if err != nil {
		return c.Status(404).JSON(models.APIResponse{
			Success: false,
			Error:   "XML not found",
		})
	}

	// Set headers for file download
	filename := fmt.Sprintf("nfse_%s_%s.xml", numero, competencia)
	c.Set("Content-Type", "application/xml")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Length", fmt.Sprintf("%d", len(xmlData)))

	return c.Send(xmlData)
}
