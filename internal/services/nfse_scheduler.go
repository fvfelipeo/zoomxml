package services

import (
	"context"
	"time"

	"github.com/zoomxml/config"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/logger"
	"github.com/zoomxml/internal/models"
)

// NFSeScheduler handles automatic NFSe document fetching
type NFSeScheduler struct {
	nfseService *NFSeService
	ticker      *time.Ticker
	stopChan    chan bool
	running     bool
	config      *config.Config
}

// NewNFSeScheduler creates a new NFSe scheduler
func NewNFSeScheduler() *NFSeScheduler {
	return &NFSeScheduler{
		nfseService: NewNFSeService(),
		stopChan:    make(chan bool),
		running:     false,
		config:      config.Get(),
	}
}

// Start begins the automatic NFSe fetching process
func (s *NFSeScheduler) Start() error {
	if !s.config.NFSeScheduler.Enabled {
		logger.InfoWithFields("NFSe scheduler is disabled", map[string]any{
			"operation": "start_scheduler",
		})
		return nil
	}

	if s.running {
		logger.WarnWithFields("NFSe scheduler already running", map[string]any{
			"operation": "start_scheduler",
		})
		return nil
	}

	// Parse interval from config
	interval, err := time.ParseDuration(s.config.NFSeScheduler.Interval)
	if err != nil {
		logger.ErrorWithFields("Invalid scheduler interval", err, map[string]any{
			"operation": "start_scheduler",
			"interval":  s.config.NFSeScheduler.Interval,
		})
		return err
	}

	s.ticker = time.NewTicker(interval)
	s.running = true

	logger.InfoWithFields("Starting NFSe scheduler", map[string]any{
		"operation":       "start_scheduler",
		"interval":        interval.String(),
		"fetch_days_back": s.config.NFSeScheduler.FetchDaysBack,
		"max_pages":       s.config.NFSeScheduler.MaxPagesPerRun,
	})

	go s.run()
	return nil
}

// Stop stops the automatic NFSe fetching process
func (s *NFSeScheduler) Stop() {
	if !s.running {
		return
	}

	logger.InfoWithFields("Stopping NFSe scheduler", map[string]any{
		"operation": "stop_scheduler",
	})

	s.stopChan <- true
	s.ticker.Stop()
	s.running = false
}

// run is the main scheduler loop
func (s *NFSeScheduler) run() {
	// Run immediately on start
	s.fetchAllCompanies()

	for {
		select {
		case <-s.ticker.C:
			s.fetchAllCompanies()
		case <-s.stopChan:
			logger.InfoWithFields("NFSe scheduler stopped", map[string]any{
				"operation": "scheduler_stopped",
			})
			return
		}
	}
}

// fetchAllCompanies fetches NFSe documents for all companies with auto_fetch enabled
func (s *NFSeScheduler) fetchAllCompanies() {
	ctx := context.Background()

	logger.InfoWithFields("Starting scheduled NFSe fetch for all companies", map[string]any{
		"operation":       "scheduled_fetch",
		"fetch_days_back": s.config.NFSeScheduler.FetchDaysBack,
	})

	// Get all companies with auto_fetch enabled
	companies := []models.Company{}
	err := database.DB.NewSelect().
		Model(&companies).
		Where("auto_fetch = true AND active = true").
		Scan(ctx)

	if err != nil {
		logger.ErrorWithFields("Failed to fetch companies for scheduled NFSe fetch", err, map[string]any{
			"operation": "scheduled_fetch",
		})
		return
	}

	logger.InfoWithFields("Found companies for scheduled fetch", map[string]any{
		"operation":       "scheduled_fetch",
		"companies_count": len(companies),
	})

	// Process each company
	successCount := 0
	for _, company := range companies {
		if s.fetchCompanyDocuments(ctx, &company) {
			successCount++
		}
	}

	logger.InfoWithFields("Completed scheduled NFSe fetch", map[string]any{
		"operation":         "scheduled_fetch",
		"companies_total":   len(companies),
		"companies_success": successCount,
	})
}

// fetchCompanyDocuments fetches NFSe documents for a specific company
func (s *NFSeScheduler) fetchCompanyDocuments(ctx context.Context, company *models.Company) bool {
	logger.InfoWithFields("Fetching NFSe documents for company", map[string]any{
		"operation":    "fetch_company_documents",
		"company_id":   company.ID,
		"company_name": company.Name,
		"company_cnpj": company.CNPJ,
	})

	// Get company credentials - use only token-based credentials
	credentials := []models.CompanyCredential{}
	err := database.DB.NewSelect().
		Model(&credentials).
		Where("company_id = ? AND active = true", company.ID).
		Where("type = 'prefeitura_token'").
		Scan(ctx)

	if err != nil {
		logger.ErrorWithFields("Failed to fetch company credentials", err, map[string]any{
			"operation":  "fetch_company_documents",
			"company_id": company.ID,
		})
		return false
	}

	if len(credentials) == 0 {
		logger.WarnWithFields("No NFSe credentials found for company", map[string]any{
			"operation":  "fetch_company_documents",
			"company_id": company.ID,
		})
		return false
	}

	logger.InfoWithFields("Found credentials for company", map[string]any{
		"operation":         "fetch_company_documents",
		"company_id":        company.ID,
		"credentials_count": len(credentials),
		"credential_types":  getCredentialTypes(credentials),
	})

	// Use the first available credential (now prioritized by token availability)
	credential := &credentials[0]

	logger.InfoWithFields("Selected credential for API call", map[string]any{
		"operation":       "fetch_company_documents",
		"company_id":      company.ID,
		"credential_id":   credential.ID,
		"credential_type": credential.Type,
	})

	// Calculate date range based on config
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -s.config.NFSeScheduler.FetchDaysBack)

	// Calculate actual days difference for verification
	daysDiff := int(endDate.Sub(startDate).Hours() / 24)

	logger.InfoWithFields("Fetching documents for date range", map[string]any{
		"operation":        "fetch_company_documents",
		"company_id":       company.ID,
		"start_date":       startDate.Format("2006-01-02"),
		"end_date":         endDate.Format("2006-01-02"),
		"config_days_back": s.config.NFSeScheduler.FetchDaysBack,
		"calculated_days":  daysDiff,
	})

	totalDocuments := 0
	// Try to fetch multiple pages
	for page := 1; page <= s.config.NFSeScheduler.MaxPagesPerRun; page++ {
		logger.InfoWithFields("Fetching NFSe documents page", map[string]any{
			"operation":       "fetch_company_documents",
			"company_id":      company.ID,
			"page":            page,
			"credential_id":   credential.ID,
			"credential_type": credential.Type,
		})

		result, err := s.nfseService.FetchNFSeDocuments(ctx, credential, startDate, endDate, page)
		if err != nil {
			logger.ErrorWithFields("Failed to fetch NFSe documents", err, map[string]any{
				"operation":     "fetch_company_documents",
				"company_id":    company.ID,
				"page":          page,
				"credential_id": credential.ID,
				"error_details": err.Error(),
			})
			break
		}

		if !result.Success {
			logger.WarnWithFields("NFSe fetch was not successful", map[string]any{
				"operation":  "fetch_company_documents",
				"company_id": company.ID,
				"page":       page,
				"result":     result,
			})
			break
		}

		if len(result.Documents) == 0 {
			logger.InfoWithFields("No more documents found", map[string]any{
				"operation":  "fetch_company_documents",
				"company_id": company.ID,
				"page":       page,
			})
			break
		}

		// Store documents
		logger.InfoWithFields("Storing NFSe documents", map[string]any{
			"operation":       "fetch_company_documents",
			"company_id":      company.ID,
			"page":            page,
			"documents_count": len(result.Documents),
		})

		err = s.nfseService.StoreNFSeDocuments(ctx, company.ID, result.Documents)
		if err != nil {
			logger.ErrorWithFields("Failed to store NFSe documents", err, map[string]any{
				"operation":     "fetch_company_documents",
				"company_id":    company.ID,
				"page":          page,
				"error_details": err.Error(),
			})
		} else {
			totalDocuments += len(result.Documents)
			logger.InfoWithFields("Successfully stored NFSe documents", map[string]any{
				"operation":       "fetch_company_documents",
				"company_id":      company.ID,
				"page":            page,
				"documents_count": len(result.Documents),
				"total_so_far":    totalDocuments,
			})
		}

		// If we got less than 100 documents (max per page), we're done
		if len(result.Documents) < 100 {
			break
		}

		// Add delay between pages to be respectful to the API
		if s.config.NFSeScheduler.APIDelaySeconds > 0 {
			time.Sleep(time.Duration(s.config.NFSeScheduler.APIDelaySeconds) * time.Second)
		}
	}

	success := totalDocuments > 0
	logger.InfoWithFields("Completed NFSe fetch for company", map[string]any{
		"operation":       "fetch_company_documents",
		"company_id":      company.ID,
		"company_name":    company.Name,
		"company_cnpj":    company.CNPJ,
		"total_documents": totalDocuments,
		"success":         success,
	})

	return success
}

// IsRunning returns whether the scheduler is currently running
func (s *NFSeScheduler) IsRunning() bool {
	return s.running
}

// FetchCompanyNow immediately fetches NFSe documents for a specific company
func (s *NFSeScheduler) FetchCompanyNow(ctx context.Context, companyID int64) error {
	// Get company
	company := &models.Company{}
	err := database.DB.NewSelect().
		Model(company).
		Where("id = ? AND active = true", companyID).
		Scan(ctx)

	if err != nil {
		return err
	}

	// Fetch documents
	s.fetchCompanyDocuments(ctx, company)
	return nil
}

// GetStatus returns the current status of the scheduler
func (s *NFSeScheduler) GetStatus() map[string]any {
	return map[string]any{
		"running":           s.running,
		"enabled":           s.config.NFSeScheduler.Enabled,
		"interval":          s.config.NFSeScheduler.Interval,
		"fetch_days_back":   s.config.NFSeScheduler.FetchDaysBack,
		"max_pages_per_run": s.config.NFSeScheduler.MaxPagesPerRun,
		"api_delay_seconds": s.config.NFSeScheduler.APIDelaySeconds,
	}
}

// getCredentialTypes returns a slice of credential types for logging
func getCredentialTypes(credentials []models.CompanyCredential) []string {
	types := make([]string, len(credentials))
	for i, cred := range credentials {
		types[i] = cred.Type
	}
	return types
}
