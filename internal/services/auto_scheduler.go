package services

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
)

// AutoSchedulerService handles automatic scheduling of XML consultations
type AutoSchedulerService struct {
	cron               *cron.Cron
	empresaRepo        *database.EmpresaRepository
	jobRepo            *database.JobRepository
	xmlConsultationSvc *XMLConsultationService
	isRunning          bool
	scheduledJobs      map[int]cron.EntryID // empresa_id -> cron entry id
}

// NewAutoSchedulerService creates a new auto scheduler service
func NewAutoSchedulerService(
	empresaRepo *database.EmpresaRepository,
	jobRepo *database.JobRepository,
	xmlConsultationSvc *XMLConsultationService,
) *AutoSchedulerService {
	return &AutoSchedulerService{
		cron:               cron.New(cron.WithSeconds()),
		empresaRepo:        empresaRepo,
		jobRepo:            jobRepo,
		xmlConsultationSvc: xmlConsultationSvc,
		scheduledJobs:      make(map[int]cron.EntryID),
	}
}

// Start starts the automatic scheduler
func (s *AutoSchedulerService) Start() error {
	if s.isRunning {
		return fmt.Errorf("scheduler is already running")
	}

	log.Println("🚀 Starting automatic XML consultation scheduler")

	// Schedule initial setup job to configure all active empresas
	_, err := s.cron.AddFunc("@every 5m", s.setupActiveEmpresasSchedules)
	if err != nil {
		return fmt.Errorf("failed to schedule setup job: %v", err)
	}

	// Schedule cleanup job to remove inactive empresa schedules
	_, err = s.cron.AddFunc("@every 1h", s.cleanupInactiveSchedules)
	if err != nil {
		return fmt.Errorf("failed to schedule cleanup job: %v", err)
	}

	// Start the cron scheduler
	s.cron.Start()
	s.isRunning = true

	// Run initial setup
	s.setupActiveEmpresasSchedules()

	log.Println("✅ Automatic XML consultation scheduler started successfully")
	return nil
}

// Stop stops the automatic scheduler
func (s *AutoSchedulerService) Stop() {
	if !s.isRunning {
		return
	}

	log.Println("🛑 Stopping automatic XML consultation scheduler")

	// Stop the cron scheduler
	ctx := s.cron.Stop()
	<-ctx.Done()

	s.isRunning = false
	s.scheduledJobs = make(map[int]cron.EntryID)

	log.Println("✅ Automatic XML consultation scheduler stopped")
}

// setupActiveEmpresasSchedules sets up schedules for all active empresas
func (s *AutoSchedulerService) setupActiveEmpresasSchedules() {
	log.Println("🔧 Setting up schedules for active empresas")

	// Get all active empresas with auto-sync enabled
	empresas, err := s.getActiveEmpresasWithAutoSync()
	if err != nil {
		log.Printf("❌ Failed to get active empresas: %v", err)
		return
	}

	log.Printf("📋 Found %d active empresas with auto-sync enabled", len(empresas))

	for _, empresa := range empresas {
		s.scheduleEmpresaConsultation(empresa)
	}
}

// getActiveEmpresasWithAutoSync gets all active empresas with auto-sync enabled
func (s *AutoSchedulerService) getActiveEmpresasWithAutoSync() ([]*models.Empresa, error) {
	// TODO: Implement repository method to get empresas with auto-sync enabled
	// For now, get all active empresas and filter

	pagination := models.PaginationRequest{Page: 1, PerPage: 1000}
	empresas, _, err := s.empresaRepo.List(pagination, "active")
	if err != nil {
		return nil, err
	}

	// Filter empresas with auto-sync enabled
	var activeEmpresasWithAutoSync []*models.Empresa
	for _, empresa := range empresas {
		if empresa.AutoSyncEnabled && empresa.IsActive() {
			activeEmpresasWithAutoSync = append(activeEmpresasWithAutoSync, &empresa)
		}
	}

	return activeEmpresasWithAutoSync, nil
}

// scheduleEmpresaConsultation schedules automatic consultation for an empresa
func (s *AutoSchedulerService) scheduleEmpresaConsultation(empresa *models.Empresa) {
	// Check if already scheduled
	if _, exists := s.scheduledJobs[empresa.ID]; exists {
		return
	}

	// Determine cron schedule based on empresa's sync interval
	cronSchedule := s.buildCronSchedule(empresa.SyncIntervalHours)

	log.Printf("📅 Scheduling automatic consultation for empresa %s (%s) with interval %d hours",
		empresa.CNPJ, empresa.RazaoSocial, empresa.SyncIntervalHours)

	// Create the consultation job function
	jobFunc := func() {
		s.executeAutomaticConsultation(empresa.ID)
	}

	// Schedule the job
	entryID, err := s.cron.AddFunc(cronSchedule, jobFunc)
	if err != nil {
		log.Printf("❌ Failed to schedule consultation for empresa %s: %v", empresa.CNPJ, err)
		return
	}

	// Store the entry ID for later management
	s.scheduledJobs[empresa.ID] = entryID

	log.Printf("✅ Scheduled automatic consultation for empresa %s with schedule: %s", empresa.CNPJ, cronSchedule)
}

// buildCronSchedule builds a cron schedule string based on sync interval hours
func (s *AutoSchedulerService) buildCronSchedule(intervalHours int) string {
	if intervalHours <= 0 {
		intervalHours = 1 // Default to 1 hour
	}

	// For intervals <= 24 hours, use hourly schedule
	if intervalHours <= 24 {
		return fmt.Sprintf("0 0 */%d * * *", intervalHours) // Every N hours
	}

	// For longer intervals, use daily schedule
	days := intervalHours / 24
	return fmt.Sprintf("0 0 2 */%d * *", days) // Every N days at 2 AM
}

// executeAutomaticConsultation executes automatic consultation for an empresa
func (s *AutoSchedulerService) executeAutomaticConsultation(empresaID int) {
	log.Printf("🔄 Executing automatic consultation for empresa ID %d", empresaID)

	// Get current and previous month competencias
	competencias := s.getCompetenciasToConsult()

	for _, competencia := range competencias {
		// Create consultation request
		request := AutomaticConsultationRequest{
			EmpresaID:    empresaID,
			Competencia:  competencia,
			ForceRefresh: false,
		}

		// Create background job for the consultation
		err := s.createConsultationJob(request)
		if err != nil {
			log.Printf("❌ Failed to create consultation job for empresa %d, competencia %s: %v",
				empresaID, competencia, err)
			continue
		}

		log.Printf("📋 Created consultation job for empresa %d, competencia %s", empresaID, competencia)
	}
}

// getCompetenciasToConsult returns the competencias that should be consulted
func (s *AutoSchedulerService) getCompetenciasToConsult() []string {
	now := time.Now()
	competencias := []string{}

	// Current month
	currentCompetencia := now.Format("2006-01")
	competencias = append(competencias, currentCompetencia)

	// Previous month (if we're early in the current month)
	if now.Day() <= 5 {
		previousMonth := now.AddDate(0, -1, 0)
		previousCompetencia := previousMonth.Format("2006-01")
		competencias = append(competencias, previousCompetencia)
	}

	return competencias
}

// createConsultationJob creates a background job for XML consultation
func (s *AutoSchedulerService) createConsultationJob(request AutomaticConsultationRequest) error {
	// Create job parameters
	parameters := map[string]interface{}{
		"empresa_id":    request.EmpresaID,
		"competencia":   request.Competencia,
		"force_refresh": request.ForceRefresh,
		"job_type":      "automatic_xml_consultation",
	}

	// Create the job with high priority (1) and schedule for immediate execution
	_, err := s.jobRepo.Create(
		request.EmpresaID,
		models.JobTypeXMLConsultation,
		1, // High priority for automatic jobs
		time.Now(),
		parameters,
	)

	return err
}

// cleanupInactiveSchedules removes schedules for inactive empresas
func (s *AutoSchedulerService) cleanupInactiveSchedules() {
	log.Println("🧹 Cleaning up inactive empresa schedules")

	// Get current active empresas
	activeEmpresasMap := make(map[int]bool)
	activeEmpresas, err := s.getActiveEmpresasWithAutoSync()
	if err != nil {
		log.Printf("❌ Failed to get active empresas for cleanup: %v", err)
		return
	}

	for _, empresa := range activeEmpresas {
		activeEmpresasMap[empresa.ID] = true
	}

	// Remove schedules for empresas that are no longer active
	for empresaID, entryID := range s.scheduledJobs {
		if !activeEmpresasMap[empresaID] {
			s.cron.Remove(entryID)
			delete(s.scheduledJobs, empresaID)
			log.Printf("🗑️ Removed schedule for inactive empresa ID %d", empresaID)
		}
	}
}

// AddEmpresaSchedule manually adds a schedule for a specific empresa
func (s *AutoSchedulerService) AddEmpresaSchedule(empresaID int) error {
	empresa, err := s.empresaRepo.GetByID(empresaID)
	if err != nil {
		return fmt.Errorf("failed to get empresa: %v", err)
	}

	if empresa == nil || !empresa.IsActive() || !empresa.AutoSyncEnabled {
		return fmt.Errorf("empresa is not eligible for automatic scheduling")
	}

	s.scheduleEmpresaConsultation(empresa)
	return nil
}

// RemoveEmpresaSchedule manually removes a schedule for a specific empresa
func (s *AutoSchedulerService) RemoveEmpresaSchedule(empresaID int) error {
	if entryID, exists := s.scheduledJobs[empresaID]; exists {
		s.cron.Remove(entryID)
		delete(s.scheduledJobs, empresaID)
		log.Printf("🗑️ Removed schedule for empresa ID %d", empresaID)
		return nil
	}
	return fmt.Errorf("no schedule found for empresa ID %d", empresaID)
}

// GetScheduledEmpresas returns the list of currently scheduled empresas
func (s *AutoSchedulerService) GetScheduledEmpresas() []int {
	empresaIDs := make([]int, 0, len(s.scheduledJobs))
	for empresaID := range s.scheduledJobs {
		empresaIDs = append(empresaIDs, empresaID)
	}
	return empresaIDs
}

// IsRunning returns whether the scheduler is currently running
func (s *AutoSchedulerService) IsRunning() bool {
	return s.isRunning
}
