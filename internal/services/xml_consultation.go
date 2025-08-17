package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
	"github.com/zoomxml/internal/storage"
)

// XMLConsultationService handles automatic XML consultation from external APIs
type XMLConsultationService struct {
	empresaRepo     *database.EmpresaRepository
	jobRepo         *database.JobRepository
	nfseManager     *storage.NFSeMinIOManager
	storageProvider *storage.MinIOProvider
}

// NewXMLConsultationService creates a new XML consultation service
func NewXMLConsultationService(
	empresaRepo *database.EmpresaRepository,
	jobRepo *database.JobRepository,
	nfseManager *storage.NFSeMinIOManager,
	storageProvider *storage.MinIOProvider,
) *XMLConsultationService {
	return &XMLConsultationService{
		empresaRepo:     empresaRepo,
		jobRepo:         jobRepo,
		nfseManager:     nfseManager,
		storageProvider: storageProvider,
	}
}

// AutomaticConsultationRequest represents a consultation request
type AutomaticConsultationRequest struct {
	EmpresaID    int
	Competencia  string
	ForceRefresh bool
}

// ConsultationResult represents the result of an automatic consultation
type ConsultationResult struct {
	EmpresaID       int                   `json:"empresa_id"`
	EmpresaCNPJ     string                `json:"empresa_cnpj"`
	Competencia     string                `json:"competencia"`
	XMLsFound       int                   `json:"xmls_found"`
	XMLsProcessed   int                   `json:"xmls_processed"`
	XMLsErrors      int                   `json:"xmls_errors"`
	ProcessingTime  time.Duration         `json:"processing_time"`
	LastProcessedAt time.Time             `json:"last_processed_at"`
	Details         []XMLProcessingDetail `json:"details"`
	Errors          []string              `json:"errors"`
}

// XMLProcessingDetail represents details of a single XML processing
type XMLProcessingDetail struct {
	NumeroNFSe   string    `json:"numero_nfse"`
	DataEmissao  string    `json:"data_emissao"`
	Status       string    `json:"status"`
	StoragePath  string    `json:"storage_path"`
	ProcessedAt  time.Time `json:"processed_at"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// ExternalXMLData represents XML data retrieved from external API
type ExternalXMLData struct {
	NumeroNFSe  string
	DataEmissao string
	XMLContent  []byte
	ContentType string
	SourceAPI   string
	RetrievedAt time.Time
}

// PerformAutomaticConsultation performs automatic XML consultation for an empresa
func (s *XMLConsultationService) PerformAutomaticConsultation(ctx context.Context, request AutomaticConsultationRequest) (*ConsultationResult, error) {
	startTime := time.Now()

	log.Printf("🔄 Starting automatic XML consultation for empresa ID %d, competencia %s", request.EmpresaID, request.Competencia)

	// Get empresa details
	empresa, err := s.empresaRepo.GetByID(request.EmpresaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get empresa: %v", err)
	}

	if empresa == nil || !empresa.IsActive() {
		return nil, fmt.Errorf("empresa not found or inactive")
	}

	if !empresa.AutoSyncEnabled {
		return nil, fmt.Errorf("automatic sync is disabled for empresa %s", empresa.CNPJ)
	}

	// Initialize result
	result := &ConsultationResult{
		EmpresaID:       empresa.ID,
		EmpresaCNPJ:     empresa.CNPJ,
		Competencia:     request.Competencia,
		LastProcessedAt: time.Now(),
		Details:         []XMLProcessingDetail{},
		Errors:          []string{},
	}

	// Check if we should skip consultation (unless forced)
	if !request.ForceRefresh && s.shouldSkipConsultation(empresa, request.Competencia) {
		log.Printf("⏭️ Skipping consultation for empresa %s, competencia %s (recently processed)", empresa.CNPJ, request.Competencia)
		return result, nil
	}

	// Perform external API consultation
	xmlsFound, err := s.consultExternalAPI(ctx, empresa, request.Competencia)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to consult external API: %v", err)
		result.Errors = append(result.Errors, errorMsg)
		log.Printf("❌ %s", errorMsg)
		return result, err
	}

	result.XMLsFound = len(xmlsFound)
	log.Printf("📄 Found %d XMLs for empresa %s, competencia %s", len(xmlsFound), empresa.CNPJ, request.Competencia)

	// Process each XML found
	for _, xmlData := range xmlsFound {
		detail := s.processXMLData(ctx, empresa, xmlData, request.Competencia)
		result.Details = append(result.Details, detail)

		if detail.Status == "success" {
			result.XMLsProcessed++
		} else {
			result.XMLsErrors++
			if detail.ErrorMessage != "" {
				result.Errors = append(result.Errors, detail.ErrorMessage)
			}
		}
	}

	// Update empresa last sync time
	err = s.empresaRepo.UpdateLastSync(empresa.ID, time.Now())
	if err != nil {
		log.Printf("⚠️ Failed to update last sync time for empresa %s: %v", empresa.CNPJ, err)
	}

	result.ProcessingTime = time.Since(startTime)

	log.Printf("✅ Completed automatic XML consultation for empresa %s: %d found, %d processed, %d errors in %v",
		empresa.CNPJ, result.XMLsFound, result.XMLsProcessed, result.XMLsErrors, result.ProcessingTime)

	return result, nil
}

// consultExternalAPI consults external NFS-e API for XMLs
func (s *XMLConsultationService) consultExternalAPI(ctx context.Context, empresa *models.Empresa, competencia string) ([]ExternalXMLData, error) {
	log.Printf("🌐 Consulting external API for empresa %s, competencia %s", empresa.CNPJ, competencia)

	// TODO: Implement actual external API consultation based on empresa configuration
	// This would involve:
	// 1. Determining which API to use based on empresa's municipality
	// 2. Authenticating with the external API using empresa's token
	// 3. Querying for NFS-e data for the specified competencia
	// 4. Parsing the response and extracting XML data

	// For now, simulate API consultation with mock data
	mockXMLs := s.generateMockXMLData(empresa, competencia)

	// Add artificial delay to simulate API call
	time.Sleep(1 * time.Second)

	return mockXMLs, nil
}

// processXMLData processes a single XML data and stores it
func (s *XMLConsultationService) processXMLData(ctx context.Context, empresa *models.Empresa, xmlData ExternalXMLData, competencia string) XMLProcessingDetail {
	detail := XMLProcessingDetail{
		NumeroNFSe:  xmlData.NumeroNFSe,
		DataEmissao: xmlData.DataEmissao,
		ProcessedAt: time.Now(),
	}

	// Check if XML already exists (avoid duplicates)
	exists, err := s.checkXMLExists(ctx, empresa.CNPJ, competencia, xmlData.NumeroNFSe)
	if err != nil {
		detail.Status = "error"
		detail.ErrorMessage = fmt.Sprintf("Failed to check if XML exists: %v", err)
		return detail
	}

	if exists {
		detail.Status = "skipped"
		detail.ErrorMessage = "XML already exists"
		log.Printf("⏭️ Skipping XML %s for empresa %s (already exists)", xmlData.NumeroNFSe, empresa.CNPJ)
		return detail
	}

	// Store XML in MinIO
	storagePath, err := s.nfseManager.StoreXML(ctx, empresa.CNPJ, competencia, xmlData.NumeroNFSe, xmlData.XMLContent)
	if err != nil {
		detail.Status = "error"
		detail.ErrorMessage = fmt.Sprintf("Failed to store XML: %v", err)
		log.Printf("❌ Failed to store XML %s for empresa %s: %v", xmlData.NumeroNFSe, empresa.CNPJ, err)
		return detail
	}

	detail.StoragePath = storagePath
	detail.Status = "success"

	// TODO: Parse XML and store metadata in PostgreSQL
	// This would involve:
	// 1. Parsing the XML to extract NFS-e data
	// 2. Creating records in the database tables
	// 3. Linking the XML file to the database records

	log.Printf("💾 Successfully stored XML %s for empresa %s at %s", xmlData.NumeroNFSe, empresa.CNPJ, storagePath)
	return detail
}

// checkXMLExists checks if an XML already exists in storage
func (s *XMLConsultationService) checkXMLExists(ctx context.Context, cnpj, competencia, numeroNFSe string) (bool, error) {
	// Try to get the XML from storage
	_, err := s.nfseManager.GetXML(ctx, cnpj, competencia, numeroNFSe)
	if err != nil {
		// If error is "not found", then it doesn't exist
		if err.Error() == "file not found" || err.Error() == "object not found" {
			return false, nil
		}
		// Other errors should be reported
		return false, err
	}
	// If no error, then it exists
	return true, nil
}

// shouldSkipConsultation determines if consultation should be skipped
func (s *XMLConsultationService) shouldSkipConsultation(empresa *models.Empresa, competencia string) bool {
	// Skip if last sync was very recent (within the last hour)
	if empresa.LastSync != nil {
		timeSinceLastSync := time.Since(*empresa.LastSync)
		if timeSinceLastSync < time.Hour {
			return true
		}
	}

	// TODO: Add more sophisticated logic:
	// - Check if competencia is too old
	// - Check if competencia is in the future
	// - Check empresa-specific consultation rules

	return false
}

// generateMockXMLData generates mock XML data for testing
func (s *XMLConsultationService) generateMockXMLData(empresa *models.Empresa, competencia string) []ExternalXMLData {
	// Generate 1-3 mock XMLs
	count := 1 + (int(time.Now().Unix()) % 3)
	xmls := make([]ExternalXMLData, count)

	for i := 0; i < count; i++ {
		numeroNFSe := fmt.Sprintf("%06d", 1000+i)
		dataEmissao := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		// Generate mock XML content
		xmlContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<nfse>
    <numero>%s</numero>
    <dataEmissao>%s</dataEmissao>
    <prestador>
        <cnpj>%s</cnpj>
        <razaoSocial>%s</razaoSocial>
    </prestador>
    <competencia>%s</competencia>
    <valorServicos>1000.00</valorServicos>
    <valorIss>50.00</valorIss>
</nfse>`, numeroNFSe, dataEmissao, empresa.CNPJ, empresa.RazaoSocial, competencia)

		xmls[i] = ExternalXMLData{
			NumeroNFSe:  numeroNFSe,
			DataEmissao: dataEmissao,
			XMLContent:  []byte(xmlContent),
			ContentType: "application/xml",
			SourceAPI:   "mock-api",
			RetrievedAt: time.Now(),
		}
	}

	return xmls
}

// GetConsultationHistory gets consultation history for an empresa
func (s *XMLConsultationService) GetConsultationHistory(empresaID int, limit int) ([]ConsultationResult, error) {
	// TODO: Implement consultation history retrieval from database
	// This would store consultation results and allow querying them
	return []ConsultationResult{}, nil
}
