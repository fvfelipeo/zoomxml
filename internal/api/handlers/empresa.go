package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
)

// EmpresaHandler handles empresa endpoints
type EmpresaHandler struct {
	empresaRepo *database.EmpresaRepository
}

// NewEmpresaHandler creates a new empresa handler
func NewEmpresaHandler(empresaRepo *database.EmpresaRepository) *EmpresaHandler {
	return &EmpresaHandler{
		empresaRepo: empresaRepo,
	}
}

// Create creates a new empresa
// @Summary Create Empresa
// @Description Create a new empresa
// @Tags empresas
// @Accept json
// @Produce json
// @Param request body models.EmpresaCreateRequest true "Empresa data"
// @Success 201 {object} models.APIResponse{data=models.Empresa}
// @Failure 400 {object} models.APIResponse
// @Router /empresas [post]
func (h *EmpresaHandler) Create(c *fiber.Ctx) error {
	var req models.EmpresaCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Set defaults
	if req.SyncIntervalHours == 0 {
		req.SyncIntervalHours = 24
	}

	empresa, err := h.empresaRepo.Create(&req)
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(201).JSON(models.APIResponse{
		Success: true,
		Message: "Empresa created successfully",
		Data:    empresa,
	})
}

// List lists empresas with pagination
// @Summary List Empresas
// @Description List empresas with pagination
// @Tags empresas
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} models.APIResponse{data=[]models.Empresa}
// @Router /empresas [get]
func (h *EmpresaHandler) List(c *fiber.Ctx) error {
	// Parse pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	status := c.Query("status")

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

	empresas, total, err := h.empresaRepo.List(pagination, status)
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
		Data:    empresas,
		Meta:    meta,
	})
}

// GetByID gets an empresa by ID
// @Summary Get Empresa
// @Description Get empresa by ID
// @Tags empresas
// @Produce json
// @Param id path int true "Empresa ID"
// @Success 200 {object} models.APIResponse{data=models.Empresa}
// @Failure 404 {object} models.APIResponse
// @Router /empresas/{id} [get]
func (h *EmpresaHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa ID",
		})
	}

	empresa, err := h.empresaRepo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if empresa == nil {
		return c.Status(404).JSON(models.APIResponse{
			Success: false,
			Error:   "Empresa not found",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    empresa,
	})
}

// Update updates an empresa
// @Summary Update Empresa
// @Description Update empresa by ID
// @Tags empresas
// @Accept json
// @Produce json
// @Param id path int true "Empresa ID"
// @Param request body models.EmpresaUpdateRequest true "Update data"
// @Success 200 {object} models.APIResponse{data=models.Empresa}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /empresas/{id} [put]
func (h *EmpresaHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa ID",
		})
	}

	var req models.EmpresaUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	empresa, err := h.empresaRepo.Update(id, &req)
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Empresa updated successfully",
		Data:    empresa,
	})
}

// Delete deletes an empresa
// @Summary Delete Empresa
// @Description Delete empresa by ID (soft delete)
// @Tags empresas
// @Produce json
// @Param id path int true "Empresa ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Router /empresas/{id} [delete]
func (h *EmpresaHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa ID",
		})
	}

	err = h.empresaRepo.Delete(id)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Empresa deleted successfully",
	})
}
