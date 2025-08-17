package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zoomxml/internal/models"
)

// JobRepository handles database operations for processing jobs
type JobRepository struct {
	db *sql.DB
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Create creates a new processing job
func (r *JobRepository) Create(empresaID int, jobType string, priority int, scheduledAt time.Time, parameters map[string]interface{}) (*models.ProcessingJob, error) {
	// Marshal parameters to JSON
	paramJSON, err := json.Marshal(parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %v", err)
	}

	query := `
		INSERT INTO nfse.processing_jobs (
			empresa_id, job_type, priority, scheduled_at, parameters
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, uuid, created_at, updated_at`

	var job models.ProcessingJob
	err = r.db.QueryRow(query, empresaID, jobType, priority, scheduledAt, paramJSON).Scan(
		&job.ID, &job.UUID, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %v", err)
	}

	// Fill in the rest of the fields
	job.EmpresaID = empresaID
	job.JobType = jobType
	job.Priority = priority
	job.ScheduledAt = scheduledAt
	job.Parameters = parameters
	job.Status = models.JobStatusPending
	job.MaxRetries = 3

	return &job, nil
}

// GetByID gets a job by ID
func (r *JobRepository) GetByID(id int) (*models.ProcessingJob, error) {
	query := `
		SELECT id, uuid, empresa_id, job_type, status, priority, scheduled_at,
			   started_at, completed_at, parameters, result, error_message,
			   retry_count, max_retries, created_at, updated_at
		FROM nfse.processing_jobs WHERE id = $1`

	return r.scanJob(r.db.QueryRow(query, id))
}

// GetByUUID gets a job by UUID
func (r *JobRepository) GetByUUID(uuid string) (*models.ProcessingJob, error) {
	query := `
		SELECT id, uuid, empresa_id, job_type, status, priority, scheduled_at,
			   started_at, completed_at, parameters, result, error_message,
			   retry_count, max_retries, created_at, updated_at
		FROM nfse.processing_jobs WHERE uuid = $1`

	return r.scanJob(r.db.QueryRow(query, uuid))
}

// List lists jobs with pagination and filters
func (r *JobRepository) List(pagination models.PaginationRequest, empresaID *int, status string, jobType string) ([]models.ProcessingJob, int, error) {
	// Build WHERE clause
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	conditions := []string{}
	if empresaID != nil {
		conditions = append(conditions, fmt.Sprintf("empresa_id = $%d", argIndex))
		args = append(args, *empresaID)
		argIndex++
	}
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}
	if jobType != "" {
		conditions = append(conditions, fmt.Sprintf("job_type = $%d", argIndex))
		args = append(args, jobType)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM nfse.processing_jobs" + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %v", err)
	}

	// Get paginated results
	query := `
		SELECT id, uuid, empresa_id, job_type, status, priority, scheduled_at,
			   started_at, completed_at, parameters, result, error_message,
			   retry_count, max_retries, created_at, updated_at
		FROM nfse.processing_jobs` + whereClause + `
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, pagination.PerPage, pagination.CalculateOffset())

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list jobs: %v", err)
	}
	defer rows.Close()

	var jobs []models.ProcessingJob
	for rows.Next() {
		job, err := r.scanJob(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan job: %v", err)
		}
		jobs = append(jobs, *job)
	}

	return jobs, total, nil
}

// GetPendingJobs gets jobs that are ready to be processed
func (r *JobRepository) GetPendingJobs(limit int) ([]models.ProcessingJob, error) {
	query := `
		SELECT id, uuid, empresa_id, job_type, status, priority, scheduled_at,
			   started_at, completed_at, parameters, result, error_message,
			   retry_count, max_retries, created_at, updated_at
		FROM nfse.processing_jobs
		WHERE status = $1 AND scheduled_at <= $2
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT $3`

	rows, err := r.db.Query(query, models.JobStatusPending, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %v", err)
	}
	defer rows.Close()

	var jobs []models.ProcessingJob
	for rows.Next() {
		job, err := r.scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %v", err)
		}
		jobs = append(jobs, *job)
	}

	return jobs, nil
}

// UpdateStatus updates the status of a job
func (r *JobRepository) UpdateStatus(id int, status string) error {
	query := "UPDATE nfse.processing_jobs SET status = $1, updated_at = $2 WHERE id = $3"
	_, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %v", err)
	}
	return nil
}

// StartJob marks a job as started
func (r *JobRepository) StartJob(id int) error {
	query := `
		UPDATE nfse.processing_jobs
		SET status = $1, started_at = $2, updated_at = $3
		WHERE id = $4`
	_, err := r.db.Exec(query, models.JobStatusRunning, time.Now(), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to start job: %v", err)
	}
	return nil
}

// CompleteJob marks a job as completed with result
func (r *JobRepository) CompleteJob(id int, result map[string]interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	query := `
		UPDATE nfse.processing_jobs
		SET status = $1, completed_at = $2, result = $3, updated_at = $4
		WHERE id = $5`
	_, err = r.db.Exec(query, models.JobStatusCompleted, time.Now(), resultJSON, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to complete job: %v", err)
	}
	return nil
}

// FailJob marks a job as failed with error message
func (r *JobRepository) FailJob(id int, errorMessage string) error {
	query := `
		UPDATE nfse.processing_jobs
		SET status = $1, completed_at = $2, error_message = $3,
		    retry_count = retry_count + 1, updated_at = $4
		WHERE id = $5`
	_, err := r.db.Exec(query, models.JobStatusFailed, time.Now(), errorMessage, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to fail job: %v", err)
	}
	return nil
}

// RetryJob resets a job for retry
func (r *JobRepository) RetryJob(id int, scheduledAt time.Time) error {
	query := `
		UPDATE nfse.processing_jobs
		SET status = $1, scheduled_at = $2, started_at = NULL, completed_at = NULL,
		    error_message = '', updated_at = $3
		WHERE id = $4`
	_, err := r.db.Exec(query, models.JobStatusPending, scheduledAt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to retry job: %v", err)
	}
	return nil
}

// CancelJob cancels a job
func (r *JobRepository) CancelJob(id int) error {
	query := `
		UPDATE nfse.processing_jobs
		SET status = $1, completed_at = $2, updated_at = $3
		WHERE id = $4 AND status IN ($5, $6)`
	_, err := r.db.Exec(query, models.JobStatusCancelled, time.Now(), time.Now(), id,
		models.JobStatusPending, models.JobStatusRunning)
	if err != nil {
		return fmt.Errorf("failed to cancel job: %v", err)
	}
	return nil
}

// CleanupOldJobs removes old completed/failed jobs
func (r *JobRepository) CleanupOldJobs(olderThan time.Time) error {
	query := `
		DELETE FROM nfse.processing_jobs
		WHERE status IN ($1, $2, $3) AND completed_at < $4`
	_, err := r.db.Exec(query, models.JobStatusCompleted, models.JobStatusFailed,
		models.JobStatusCancelled, olderThan)
	if err != nil {
		return fmt.Errorf("failed to cleanup old jobs: %v", err)
	}
	return nil
}

// scanJob scans a row into a ProcessingJob struct
func (r *JobRepository) scanJob(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.ProcessingJob, error) {
	var job models.ProcessingJob
	var paramJSON, resultJSON []byte
	var startedAt, completedAt sql.NullTime

	err := scanner.Scan(
		&job.ID,
		&job.UUID,
		&job.EmpresaID,
		&job.JobType,
		&job.Status,
		&job.Priority,
		&job.ScheduledAt,
		&startedAt,
		&completedAt,
		&paramJSON,
		&resultJSON,
		&job.ErrorMessage,
		&job.RetryCount,
		&job.MaxRetries,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan job: %v", err)
	}

	// Handle nullable timestamps
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	// Unmarshal JSON fields
	if len(paramJSON) > 0 {
		err = json.Unmarshal(paramJSON, &job.Parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %v", err)
		}
	}

	if len(resultJSON) > 0 {
		err = json.Unmarshal(resultJSON, &job.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %v", err)
		}
	}

	return &job, nil
}
