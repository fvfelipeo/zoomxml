package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"github.com/zoomxml/internal/api/handlers"
	"github.com/zoomxml/internal/api/middleware"
	"github.com/zoomxml/internal/api/routes"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
	"github.com/zoomxml/internal/services"
	"github.com/zoomxml/internal/storage"
)

// ZoomXMLService represents the main service
type ZoomXMLService struct {
	// Database
	db          *sql.DB
	empresaRepo *database.EmpresaRepository
	authRepo    *database.AuthRepository
	jobRepo     *database.JobRepository

	// Services
	authService        *services.AuthService
	scheduler          *cron.Cron
	xmlConsultationSvc *services.XMLConsultationService
	autoScheduler      *services.AutoSchedulerService

	// Storage
	storageProvider *storage.MinIOProvider
	nfseManager     *storage.NFSeMinIOManager

	// HTTP Server
	app *fiber.App

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewZoomXMLService creates a new ZoomXML service
func NewZoomXMLService() (*ZoomXMLService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	service := &ZoomXMLService{
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize database
	if err := service.initDatabase(); err != nil {
		return nil, err
	}

	// Initialize storage
	if err := service.initStorage(); err != nil {
		return nil, err
	}

	// Initialize services
	if err := service.initServices(); err != nil {
		return nil, err
	}

	// Initialize HTTP server
	if err := service.initHTTPServer(); err != nil {
		return nil, err
	}

	// Initialize scheduler
	if err := service.initScheduler(); err != nil {
		return nil, err
	}

	return service, nil
}

// initDatabase initializes database connections and repositories
func (s *ZoomXMLService) initDatabase() error {
	dbConfig := models.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "nfse_metadata"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := connectDB(dbConfig)
	if err != nil {
		return err
	}

	s.db = db
	s.empresaRepo = database.NewEmpresaRepository(db)
	s.authRepo = database.NewAuthRepository(db)
	s.jobRepo = database.NewJobRepository(db)

	log.Println("✅ Database initialized")
	return nil
}

// initStorage initializes MinIO storage
func (s *ZoomXMLService) initStorage() error {
	storageConfig := storage.StorageConfig{
		Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:  getEnv("MINIO_ACCESS_KEY", "admin"),
		SecretKey:  getEnv("MINIO_SECRET_KEY", "password123"),
		BucketName: getEnv("MINIO_BUCKET", "nfse-storage"),
		UseSSL:     false,
	}

	provider, err := storage.NewMinIOProvider(storageConfig)
	if err != nil {
		return err
	}

	s.storageProvider = provider
	s.nfseManager = storage.NewNFSeMinIOManager(provider)

	log.Println("✅ Storage initialized")
	return nil
}

// initServices initializes business services
func (s *ZoomXMLService) initServices() error {
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	s.authService = services.NewAuthService(s.empresaRepo, s.authRepo, jwtSecret)

	// Initialize XML consultation service
	s.xmlConsultationSvc = services.NewXMLConsultationService(
		s.empresaRepo,
		s.jobRepo,
		s.nfseManager,
		s.storageProvider,
	)

	// Initialize auto scheduler service
	s.autoScheduler = services.NewAutoSchedulerService(
		s.empresaRepo,
		s.jobRepo,
		s.xmlConsultationSvc,
	)

	log.Println("✅ Services initialized")
	return nil
}

// initHTTPServer initializes the HTTP server
func (s *ZoomXMLService) initHTTPServer() error {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(models.APIResponse{
				Success: false,
				Error:   err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.LoggingMiddleware())

	// Initialize handler dependencies
	handlerDeps := handlers.HandlerDependencies{
		AuthService:     s.authService,
		EmpresaRepo:     s.empresaRepo,
		AuthRepo:        s.authRepo,
		JobRepo:         s.jobRepo,
		StorageProvider: s.storageProvider,
		NFSeManager:     s.nfseManager,
	}

	// Validate dependencies
	if err := handlers.ValidateHandlerDependencies(handlerDeps); err != nil {
		return fmt.Errorf("invalid handler dependencies: %v", err)
	}

	// Create handler container
	handlerContainer := handlers.NewHandlerContainer(handlerDeps)

	// Setup routes using the organized route module
	routeConfig := routes.RouteConfig{
		Handlers:    handlerContainer,
		AuthService: s.authService,
	}

	routes.SetupRoutes(app, routeConfig)

	s.app = app
	log.Println("✅ HTTP server initialized")
	return nil
}

// initScheduler initializes the cron scheduler
func (s *ZoomXMLService) initScheduler() error {
	s.scheduler = cron.New(cron.WithSeconds())

	// Schedule automatic sync every hour
	_, err := s.scheduler.AddFunc("0 0 * * * *", func() {
		s.runAutomaticSync()
	})
	if err != nil {
		return err
	}

	// Schedule token cleanup every day at midnight
	_, err = s.scheduler.AddFunc("0 0 0 * * *", func() {
		s.authService.CleanupExpiredTokens()
	})
	if err != nil {
		return err
	}

	// Schedule job cleanup every week
	_, err = s.scheduler.AddFunc("0 0 0 * * 0", func() {
		oldTime := time.Now().AddDate(0, 0, -30) // 30 days ago
		s.jobRepo.CleanupOldJobs(oldTime)
	})
	if err != nil {
		return err
	}

	log.Println("✅ Scheduler initialized")
	return nil
}

// Start starts all services
func (s *ZoomXMLService) Start() error {
	log.Println("🚀 Starting ZoomXML Service...")

	// Start scheduler
	s.scheduler.Start()
	log.Println("⏰ Scheduler started")

	// Start auto scheduler for XML consultations
	err := s.autoScheduler.Start()
	if err != nil {
		return fmt.Errorf("failed to start auto scheduler: %v", err)
	}
	log.Println("🤖 Auto scheduler started")

	// Start background workers
	s.wg.Add(1)
	go s.runJobProcessor()

	// Start HTTP server
	port := getEnv("PORT", "8080")
	log.Printf("🌐 Starting HTTP server on port %s", port)
	log.Printf("📚 Health check: http://localhost:%s/health", port)
	log.Printf("🔐 Auth endpoint: http://localhost:%s/api/v1/auth/login", port)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.app.Listen(":" + port); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Println("✅ ZoomXML Service started successfully")
	return nil
}

// Stop stops all services gracefully
func (s *ZoomXMLService) Stop() error {
	log.Println("🛑 Stopping ZoomXML Service...")

	// Cancel context to stop background workers
	s.cancel()

	// Stop scheduler
	s.scheduler.Stop()
	log.Println("⏰ Scheduler stopped")

	// Stop auto scheduler
	s.autoScheduler.Stop()
	log.Println("🤖 Auto scheduler stopped")

	// Stop HTTP server
	if err := s.app.Shutdown(); err != nil {
		log.Printf("Error stopping HTTP server: %v", err)
	}
	log.Println("🌐 HTTP server stopped")

	// Wait for background workers to finish
	s.wg.Wait()
	log.Println("⚙️ Background workers stopped")

	// Close database
	if s.db != nil {
		s.db.Close()
		log.Println("🗄️ Database connection closed")
	}

	log.Println("✅ ZoomXML Service stopped")
	return nil
}

// runAutomaticSync runs automatic sync for all active empresas
func (s *ZoomXMLService) runAutomaticSync() {
	log.Println("🔄 Running automatic sync...")

	empresas, err := s.empresaRepo.GetEmpresasForSync()
	if err != nil {
		log.Printf("❌ Error getting empresas for sync: %v", err)
		return
	}

	if len(empresas) == 0 {
		log.Println("ℹ️  No empresas need sync")
		return
	}

	log.Printf("📋 Found %d empresas for sync", len(empresas))

	for _, empresa := range empresas {
		// Create sync job
		parameters := map[string]interface{}{
			"empresa_id":   empresa.ID,
			"empresa_uuid": empresa.UUID,
			"sync_type":    "automatic",
		}

		_, err := s.jobRepo.Create(empresa.ID, models.JobTypeSyncNFSe, 5, time.Now(), parameters)
		if err != nil {
			log.Printf("❌ Error creating sync job for empresa %s: %v", empresa.CNPJ, err)
			continue
		}

		log.Printf("✅ Created sync job for empresa %s", empresa.CNPJ)
	}
}

// runJobProcessor processes jobs from the queue
func (s *ZoomXMLService) runJobProcessor() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	log.Println("⚙️ Job processor started")

	for {
		select {
		case <-s.ctx.Done():
			log.Println("⚙️ Job processor stopping...")
			return
		case <-ticker.C:
			s.processJobs()
		}
	}
}

// processJobs processes pending jobs
func (s *ZoomXMLService) processJobs() {
	jobs, err := s.jobRepo.GetPendingJobs(5) // Process up to 5 jobs at once
	if err != nil {
		log.Printf("❌ Error getting pending jobs: %v", err)
		return
	}

	if len(jobs) == 0 {
		return // No jobs to process
	}

	log.Printf("⚙️ Processing %d jobs", len(jobs))

	for _, job := range jobs {
		s.processJob(&job)
	}
}

// processJob processes a single job
func (s *ZoomXMLService) processJob(job *models.ProcessingJob) {
	log.Printf("🔄 Processing job %s (type: %s)", job.UUID, job.JobType)

	// Mark job as started
	if err := s.jobRepo.StartJob(job.ID); err != nil {
		log.Printf("❌ Error starting job %s: %v", job.UUID, err)
		return
	}

	var err error
	var result map[string]interface{}

	switch job.JobType {
	case models.JobTypeSyncNFSe:
		result, err = s.processSyncJob(job)
	case models.JobTypeProcessXML:
		result, err = s.processXMLJob(job)
	case models.JobTypeGenerateReport:
		result, err = s.processReportJob(job)
	case models.JobTypeXMLConsultation:
		result, err = s.processXMLConsultationJob(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.JobType)
	}

	if err != nil {
		log.Printf("❌ Job %s failed: %v", job.UUID, err)
		s.jobRepo.FailJob(job.ID, err.Error())

		// Retry if possible
		if job.CanRetry() {
			retryAt := time.Now().Add(time.Duration(job.RetryCount+1) * time.Minute)
			s.jobRepo.RetryJob(job.ID, retryAt)
			log.Printf("🔄 Job %s scheduled for retry at %v", job.UUID, retryAt)
		}
	} else {
		log.Printf("✅ Job %s completed successfully", job.UUID)
		s.jobRepo.CompleteJob(job.ID, result)
	}
}

// processSyncJob processes NFS-e sync job with automatic XML consultation
func (s *ZoomXMLService) processSyncJob(job *models.ProcessingJob) (map[string]interface{}, error) {
	empresaID, ok := job.Parameters["empresa_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid empresa_id in job parameters")
	}

	empresa, err := s.empresaRepo.GetByID(int(empresaID))
	if err != nil {
		return nil, fmt.Errorf("failed to get empresa: %v", err)
	}

	if empresa == nil || !empresa.IsActive() {
		return nil, fmt.Errorf("empresa not found or inactive")
	}

	log.Printf("🔄 Starting manual sync for empresa %s", empresa.CNPJ)

	// For manual sync, trigger immediate XML consultation
	competencias := s.getCompetenciasToConsult(empresa)
	totalProcessed := 0
	totalErrors := 0

	for _, competencia := range competencias {
		request := services.AutomaticConsultationRequest{
			EmpresaID:    int(empresaID),
			Competencia:  competencia,
			ForceRefresh: true, // Force refresh for manual sync
		}

		result, err := s.xmlConsultationSvc.PerformAutomaticConsultation(context.Background(), request)
		if err != nil {
			log.Printf("❌ Failed to consult competencia %s: %v", competencia, err)
			totalErrors++
			continue
		}

		totalProcessed += result.XMLsProcessed
		totalErrors += result.XMLsErrors
	}

	// Update last sync time
	err = s.empresaRepo.UpdateLastSync(empresa.ID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to update last sync: %v", err)
	}

	result := map[string]interface{}{
		"empresa_id":       empresa.ID,
		"empresa_cnpj":     empresa.CNPJ,
		"competencias":     competencias,
		"processed_at":     time.Now(),
		"total_processed":  totalProcessed,
		"total_errors":     totalErrors,
		"duration_seconds": time.Since(job.CreatedAt).Seconds(),
		"status":           "completed",
	}

	log.Printf("✅ Completed manual sync for empresa %s: %d processed, %d errors", empresa.CNPJ, totalProcessed, totalErrors)
	return result, nil
}

// performAutomaticXMLConsultation performs automatic consultation and storage of XMLs
func (s *ZoomXMLService) performAutomaticXMLConsultation(empresa *models.Empresa) (map[string]interface{}, error) {
	startTime := time.Now()

	// Get competências to consult (current month and previous months if needed)
	competencias := s.getCompetenciasToConsult(empresa)

	totalXMLsFound := 0
	totalProcessed := 0
	totalErrors := 0

	// Process each competência
	for _, competencia := range competencias {
		log.Printf("📋 Consulting XMLs for empresa %s, competencia %s", empresa.CNPJ, competencia)

		// Consult XMLs from API for this competência
		xmlsFound, err := s.consultXMLsFromAPI(empresa, competencia)
		if err != nil {
			log.Printf("❌ Error consulting XMLs for competencia %s: %v", competencia, err)
			continue
		}

		totalXMLsFound += len(xmlsFound)

		// Process each XML found
		for _, xmlData := range xmlsFound {
			err := s.processAndStoreXML(empresa, xmlData, competencia)
			if err != nil {
				log.Printf("❌ Error processing XML %s for empresa %s: %v", xmlData.NumeroNFSe, empresa.CNPJ, err)
				totalErrors++
			} else {
				totalProcessed++
			}
		}

		log.Printf("📄 Competencia %s: %d XMLs found, %d processed", competencia, len(xmlsFound), len(xmlsFound)-totalErrors)
	}

	duration := time.Since(startTime)

	result := map[string]interface{}{
		"empresa_id":       empresa.ID,
		"empresa_cnpj":     empresa.CNPJ,
		"competencias":     competencias,
		"processed_at":     time.Now(),
		"total_xmls_found": totalXMLsFound,
		"total_processed":  totalProcessed,
		"total_errors":     totalErrors,
		"duration_seconds": duration.Seconds(),
		"status":           "completed",
	}

	log.Printf("📊 XML consultation summary for %s: %d competências, %d XMLs found, %d processed, %d errors",
		empresa.CNPJ, len(competencias), totalXMLsFound, totalProcessed, totalErrors)

	return result, nil
}

// getCompetenciasToConsult returns the list of competências to consult for an empresa
func (s *ZoomXMLService) getCompetenciasToConsult(empresa *models.Empresa) []string {
	competencias := []string{}
	currentTime := time.Now()

	// Always include current month
	competencias = append(competencias, currentTime.Format("2006-01"))

	// If it's early in the month (first 5 days), also include previous month
	if currentTime.Day() <= 5 {
		previousMonth := currentTime.AddDate(0, -1, 0)
		competencias = append(competencias, previousMonth.Format("2006-01"))
	}

	// If this is the first sync for the empresa, include last 3 months
	if empresa.LastSync == nil {
		log.Printf("🔄 First sync for empresa %s, including last 3 months", empresa.CNPJ)
		for i := 1; i <= 3; i++ {
			pastMonth := currentTime.AddDate(0, -i, 0)
			pastCompetencia := pastMonth.Format("2006-01")

			// Avoid duplicates
			found := false
			for _, existing := range competencias {
				if existing == pastCompetencia {
					found = true
					break
				}
			}
			if !found {
				competencias = append(competencias, pastCompetencia)
			}
		}
	}

	log.Printf("📅 Competências to consult for empresa %s: %v", empresa.CNPJ, competencias)
	return competencias
}

// consultXMLsFromAPI consults XMLs from the external API
func (s *ZoomXMLService) consultXMLsFromAPI(empresa *models.Empresa, competencia string) ([]XMLData, error) {
	// This would be replaced with actual API client implementation
	// For now, simulate finding some XMLs

	log.Printf("🌐 Consulting API for empresa %s, competencia %s", empresa.CNPJ, competencia)

	// Simulate API call delay
	time.Sleep(1 * time.Second)

	// Simulate finding XMLs (replace with actual API implementation)
	xmlsFound := []XMLData{
		{
			NumeroNFSe:  "000001",
			DataEmissao: time.Now().Format("2006-01-02"),
			XMLContent:  []byte(`<?xml version="1.0" encoding="UTF-8"?><nfse><numero>000001</numero></nfse>`),
			ContentType: "application/xml",
		},
		// Add more simulated XMLs as needed
	}

	log.Printf("📄 Found %d XMLs for empresa %s", len(xmlsFound), empresa.CNPJ)
	return xmlsFound, nil
}

// processAndStoreXML processes and stores an XML in MinIO
func (s *ZoomXMLService) processAndStoreXML(empresa *models.Empresa, xmlData XMLData, competencia string) error {
	// Store XML in MinIO
	xmlPath, err := s.nfseManager.StoreXML(context.Background(), empresa.CNPJ, competencia, xmlData.NumeroNFSe, xmlData.XMLContent)
	if err != nil {
		return fmt.Errorf("failed to store XML in MinIO: %v", err)
	}

	log.Printf("💾 Stored XML %s for empresa %s at path: %s", xmlData.NumeroNFSe, empresa.CNPJ, xmlPath)

	// TODO: Parse XML and store metadata in PostgreSQL
	// This would involve:
	// 1. Parsing the XML to extract NFS-e data
	// 2. Storing metadata in the database
	// 3. Creating relationships between empresa and NFS-e

	return nil
}

// XMLData represents XML data from API consultation
type XMLData struct {
	NumeroNFSe  string
	DataEmissao string
	XMLContent  []byte
	ContentType string
}

// processXMLJob processes XML processing job
func (s *ZoomXMLService) processXMLJob(job *models.ProcessingJob) (map[string]interface{}, error) {
	// TODO: Implement XML processing logic
	time.Sleep(1 * time.Second)

	result := map[string]interface{}{
		"processed_at": time.Now(),
		"status":       "completed",
	}

	return result, nil
}

// processReportJob processes report generation job
func (s *ZoomXMLService) processReportJob(job *models.ProcessingJob) (map[string]interface{}, error) {
	// TODO: Implement report generation logic
	time.Sleep(1 * time.Second)

	result := map[string]interface{}{
		"processed_at": time.Now(),
		"status":       "completed",
	}

	return result, nil
}

// processXMLConsultationJob processes automatic XML consultation job
func (s *ZoomXMLService) processXMLConsultationJob(job *models.ProcessingJob) (map[string]interface{}, error) {
	// Extract parameters
	empresaID, ok := job.Parameters["empresa_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid empresa_id in job parameters")
	}

	competencia, ok := job.Parameters["competencia"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid competencia in job parameters")
	}

	forceRefresh, _ := job.Parameters["force_refresh"].(bool)

	log.Printf("🔄 Processing XML consultation job for empresa %d, competencia %s", int(empresaID), competencia)

	// Create consultation request
	request := services.AutomaticConsultationRequest{
		EmpresaID:    int(empresaID),
		Competencia:  competencia,
		ForceRefresh: forceRefresh,
	}

	// Perform the consultation
	result, err := s.xmlConsultationSvc.PerformAutomaticConsultation(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("failed to perform XML consultation: %v", err)
	}

	// Convert result to map for job result
	jobResult := map[string]interface{}{
		"empresa_id":        result.EmpresaID,
		"empresa_cnpj":      result.EmpresaCNPJ,
		"competencia":       result.Competencia,
		"xmls_found":        result.XMLsFound,
		"xmls_processed":    result.XMLsProcessed,
		"xmls_errors":       result.XMLsErrors,
		"processing_time":   result.ProcessingTime.Seconds(),
		"last_processed_at": result.LastProcessedAt,
		"status":            "completed",
	}

	if len(result.Errors) > 0 {
		jobResult["errors"] = result.Errors
	}

	log.Printf("✅ Completed XML consultation job for empresa %d: %d found, %d processed, %d errors",
		int(empresaID), result.XMLsFound, result.XMLsProcessed, result.XMLsErrors)

	return jobResult, nil
}

// HTTP Handlers for additional endpoints

// handleManualSync handles manual sync trigger
func (s *ZoomXMLService) handleManualSync(c *fiber.Ctx) error {
	empresaID := middleware.GetEmpresaIDFromContext(c)
	if empresaID == 0 {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid empresa context",
		})
	}

	// Create manual sync job
	parameters := map[string]interface{}{
		"empresa_id": empresaID,
		"sync_type":  "manual",
	}

	job, err := s.jobRepo.Create(empresaID, models.JobTypeSyncNFSe, 10, time.Now(), parameters)
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

// handleListJobs handles listing jobs for an empresa
func (s *ZoomXMLService) handleListJobs(c *fiber.Ctx) error {
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

	jobs, total, err := s.jobRepo.List(pagination, &empresaID, status, jobType)
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

// handleGetStats handles getting stats for an empresa
func (s *ZoomXMLService) handleGetStats(c *fiber.Ctx) error {
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

// handleListStoredXMLs lists all stored XMLs for the authenticated empresa
func (s *ZoomXMLService) handleListStoredXMLs(c *fiber.Ctx) error {
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
	files, err := s.storageProvider.List(c.Context(), prefix)
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

// handleListXMLsByCompetencia lists XMLs by competencia for the authenticated empresa
func (s *ZoomXMLService) handleListXMLsByCompetencia(c *fiber.Ctx) error {
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

	files, err := s.storageProvider.List(c.Context(), prefix)
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

// handleGetStoredXML gets a specific stored XML
func (s *ZoomXMLService) handleGetStoredXML(c *fiber.Ctx) error {
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
	xmlData, err := s.nfseManager.GetXML(c.Context(), empresa.CNPJ, competencia, numero)
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

// handleDownloadXML downloads a specific XML file
func (s *ZoomXMLService) handleDownloadXML(c *fiber.Ctx) error {
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
	xmlData, err := s.nfseManager.GetXML(c.Context(), empresa.CNPJ, competencia, numero)
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

// Utility functions

// connectDB connects to PostgreSQL database
func connectDB(config models.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ Connected to PostgreSQL database")
	return db, nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// main function
func main() {
	log.Println("🚀 ZoomXML Multi-Enterprise NFS-e Service")
	log.Println("=" + strings.Repeat("=", 45))

	// Create service
	service, err := NewZoomXMLService()
	if err != nil {
		log.Fatalf("❌ Failed to create service: %v", err)
	}

	// Start service
	if err := service.Start(); err != nil {
		log.Fatalf("❌ Failed to start service: %v", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Stop service gracefully
	if err := service.Stop(); err != nil {
		log.Printf("❌ Error during shutdown: %v", err)
	}
}
