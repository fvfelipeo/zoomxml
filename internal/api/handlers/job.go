package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/api/middleware"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
)

// JobHandler handles job-related operations
type JobHandler struct {
	jobRepo *database.JobRepository
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobRepo *database.JobRepository) *JobHandler {
	return &JobHandler{
		jobRepo: jobRepo,
	}
}

// HandleGetJob gets a specific job by ID
func (h *JobHandler) HandleGetJob(c *fiber.Ctx) error {
	empresaID := middleware.GetEmpresaIDFromContext(c)
	if empresaID == 0 {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	jobIDStr := c.Params("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid job ID",
		})
	}

	job, err := h.jobRepo.GetByID(jobID)
	if err != nil {
		return c.Status(404).JSON(models.APIResponse{
			Success: false,
			Error:   "Job not found",
		})
	}

	// Check if job belongs to the authenticated empresa
	if job.EmpresaID != empresaID {
		return c.Status(403).JSON(models.APIResponse{
			Success: false,
			Error:   "Access denied",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    job,
	})
}

// HandleCancelJob cancels a specific job
func (h *JobHandler) HandleCancelJob(c *fiber.Ctx) error {
	empresaID := middleware.GetEmpresaIDFromContext(c)
	if empresaID == 0 {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	jobIDStr := c.Params("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid job ID",
		})
	}

	job, err := h.jobRepo.GetByID(jobID)
	if err != nil {
		return c.Status(404).JSON(models.APIResponse{
			Success: false,
			Error:   "Job not found",
		})
	}

	// Check if job belongs to the authenticated empresa
	if job.EmpresaID != empresaID {
		return c.Status(403).JSON(models.APIResponse{
			Success: false,
			Error:   "Access denied",
		})
	}

	// Check if job can be cancelled
	if job.Status == models.JobStatusCompleted || job.Status == models.JobStatusFailed {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Cannot cancel completed or failed job",
		})
	}

	// Cancel the job
	err = h.jobRepo.CancelJob(jobID)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "Failed to cancel job",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Job cancelled successfully",
	})
}

// HandleRetryJob retries a failed job
func (h *JobHandler) HandleRetryJob(c *fiber.Ctx) error {
	empresaID := middleware.GetEmpresaIDFromContext(c)
	if empresaID == 0 {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	jobIDStr := c.Params("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid job ID",
		})
	}

	job, err := h.jobRepo.GetByID(jobID)
	if err != nil {
		return c.Status(404).JSON(models.APIResponse{
			Success: false,
			Error:   "Job not found",
		})
	}

	// Check if job belongs to the authenticated empresa
	if job.EmpresaID != empresaID {
		return c.Status(403).JSON(models.APIResponse{
			Success: false,
			Error:   "Access denied",
		})
	}

	// Check if job can be retried
	if job.Status != models.JobStatusFailed {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Only failed jobs can be retried",
		})
	}

	// Check retry limit
	if !job.CanRetry() {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Job has exceeded maximum retry attempts",
		})
	}

	// Retry the job
	retryAt := time.Now().Add(time.Duration(job.RetryCount+1) * time.Minute)
	err = h.jobRepo.RetryJob(jobID, retryAt)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "Failed to retry job",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Job scheduled for retry",
		Data: map[string]interface{}{
			"retry_at":    retryAt,
			"retry_count": job.RetryCount + 1,
		},
	})
}

// HandleListJobsByStatus lists jobs by status
func (h *JobHandler) HandleListJobsByStatus(c *fiber.Ctx) error {
	empresaID := middleware.GetEmpresaIDFromContext(c)
	if empresaID == 0 {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	status := c.Params("status")
	if status == "" {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Status is required",
		})
	}

	// Parse pagination
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
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

// HandleGetJobStats gets job statistics for the empresa
func (h *JobHandler) HandleGetJobStats(c *fiber.Ctx) error {
	empresaID := middleware.GetEmpresaIDFromContext(c)
	if empresaID == 0 {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	// TODO: Implement actual job statistics calculation
	// This would query the database for job statistics

	stats := map[string]interface{}{
		"total_jobs":     0,
		"pending_jobs":   0,
		"running_jobs":   0,
		"completed_jobs": 0,
		"failed_jobs":    0,
		"cancelled_jobs": 0,
		"success_rate":   0.0,
		"avg_duration":   0.0,
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    stats,
	})
}
