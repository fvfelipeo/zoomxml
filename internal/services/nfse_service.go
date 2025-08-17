package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zoomxml/internal/logger"
	"github.com/zoomxml/internal/models"
)

// NFSeService handles NFSe API operations
type NFSeService struct {
	client     *http.Client
	xmlManager *NFSeXMLManager
}

// PrefeituraModernaResponse represents the actual response from Prefeitura Moderna API
type PrefeituraModernaResponse struct {
	RecordCount    int                    `json:"RecordCount"`
	RecordsPerPage int                    `json:"RecordsPerPage"`
	PageCount      int                    `json:"PageCount"`
	CurrentPage    int                    `json:"CurrentPage"`
	Dados          []PrefeituraModernaDoc `json:"Dados"`
}

// PrefeituraModernaDoc represents a single NFSe document from the API
type PrefeituraModernaDoc struct {
	NrNfse        int    `json:"NrNfse"`        // Número da NFSe
	DtEmissao     string `json:"DtEmissao"`     // Data de emissão
	NrCompetencia int    `json:"NrCompetencia"` // Competência YYYYMM
	XmlCompactado string `json:"XmlCompactado"` // ZIP em Base64 contendo o XML
}

// NFSeDocument represents a processed NFSe document
type NFSeDocument struct {
	FileName    string    `json:"file_name"`   // Nome do arquivo XML
	XMLContent  string    `json:"xml_content"` // Conteúdo XML
	ProcessedAt time.Time `json:"processed_at"`
}

// NFSeProcessResult represents the result of processing NFSe documents
type NFSeProcessResult struct {
	Success        bool           `json:"success"`
	Message        string         `json:"message"`
	DocumentsCount int            `json:"documents_count"`
	Documents      []NFSeDocument `json:"documents,omitempty"`
	Error          string         `json:"error,omitempty"`
}

// NewNFSeService creates a new NFSe service instance
func NewNFSeService() *NFSeService {
	return &NFSeService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		xmlManager: NewNFSeXMLManager(),
	}
}

// extractXMLFromZip extracts XML files from a Base64 encoded ZIP
func (s *NFSeService) extractXMLFromZip(base64Zip string) ([]NFSeDocument, error) {
	// Decode Base64
	zipData, err := base64.StdEncoding.DecodeString(base64Zip)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create a reader for the ZIP data
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	var documents []NFSeDocument

	// Extract each file from the ZIP
	for _, file := range zipReader.File {
		// Open the file
		rc, err := file.Open()
		if err != nil {
			logger.ErrorWithFields("Failed to open file in ZIP", err, map[string]any{
				"operation": "extract_xml",
				"file_name": file.Name,
			})
			continue
		}

		// Read the file content
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			logger.ErrorWithFields("Failed to read file content", err, map[string]any{
				"operation": "extract_xml",
				"file_name": file.Name,
			})
			continue
		}

		// Create document
		document := NFSeDocument{
			FileName:    file.Name,
			XMLContent:  string(content),
			ProcessedAt: time.Now(),
		}

		documents = append(documents, document)

		logger.DebugWithFields("XML file extracted", map[string]any{
			"operation":    "extract_xml",
			"file_name":    file.Name,
			"content_size": len(content),
		})
	}

	return documents, nil
}

// FetchNFSeDocuments fetches NFSe documents from the municipal API
func (s *NFSeService) FetchNFSeDocuments(ctx context.Context, credential *models.CompanyCredential, startDate, endDate time.Time, page int) (*NFSeProcessResult, error) {
	// Get the API token from encrypted credentials
	_, _, token, err := credential.GetCredentialData()
	if err != nil {
		logger.ErrorWithFields("Failed to decrypt credential data", err, map[string]any{
			"operation":     "fetch_nfse",
			"credential_id": credential.ID,
			"company_id":    credential.CompanyID,
		})
		return nil, fmt.Errorf("failed to decrypt credential data: %w", err)
	}

	if token == "" {
		return nil, fmt.Errorf("API token not found in credentials")
	}

	// Build the API URL with pagination
	baseURL := "https://api-nfse-imperatriz-ma.prefeituramoderna.com.br/ws/services/xmlnfse"
	url := fmt.Sprintf("%s?dt_inicial=%s&dt_final=%s&nr_page=%d",
		baseURL,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		page,
	)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ZoomXML/1.0.0")

	logger.InfoWithFields("Making NFSe API request", map[string]any{
		"operation":     "fetch_nfse",
		"url":           url,
		"company_id":    credential.CompanyID,
		"credential_id": credential.ID,
		"start_date":    startDate.Format("2006-01-02"),
		"end_date":      endDate.Format("2006-01-02"),
	})

	// Make the request
	resp, err := s.client.Do(req)
	if err != nil {
		logger.ErrorWithFields("NFSe API request failed", err, map[string]any{
			"operation":  "fetch_nfse",
			"url":        url,
			"company_id": credential.CompanyID,
		})
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.InfoWithFields("NFSe API response received", map[string]any{
		"operation":     "fetch_nfse",
		"status_code":   resp.StatusCode,
		"company_id":    credential.CompanyID,
		"response_size": len(body),
	})

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		logger.ErrorWithFields("NFSe API returned error status", nil, map[string]any{
			"operation":   "fetch_nfse",
			"status_code": resp.StatusCode,
			"response":    string(body),
			"company_id":  credential.CompanyID,
		})
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response from Prefeitura Moderna
	var apiResponse PrefeituraModernaResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		logger.ErrorWithFields("Failed to parse JSON response", err, map[string]any{
			"operation":  "fetch_nfse",
			"company_id": credential.CompanyID,
			"response":   string(body),
		})
		return &NFSeProcessResult{
			Success: false,
			Message: "Failed to parse API response",
			Error:   err.Error(),
		}, nil
	}

	var allDocuments []NFSeDocument

	// Process each NFSe document
	for _, nfseDoc := range apiResponse.Dados {
		if nfseDoc.XmlCompactado == "" {
			logger.WarnWithFields("Empty XmlCompactado found", map[string]any{
				"operation":  "fetch_nfse",
				"company_id": credential.CompanyID,
				"nfse_nr":    nfseDoc.NrNfse,
			})
			continue
		}

		// Extract XML files from ZIP
		documents, err := s.extractXMLFromZip(nfseDoc.XmlCompactado)
		if err != nil {
			logger.ErrorWithFields("Failed to extract XML from ZIP", err, map[string]any{
				"operation":  "fetch_nfse",
				"company_id": credential.CompanyID,
				"nfse_nr":    nfseDoc.NrNfse,
			})
			continue
		}

		allDocuments = append(allDocuments, documents...)
	}

	logger.InfoWithFields("NFSe documents fetched successfully", map[string]any{
		"operation":       "fetch_nfse",
		"company_id":      credential.CompanyID,
		"documents_count": len(allDocuments),
		"page":            page,
		"total_records":   apiResponse.RecordCount,
	})

	return &NFSeProcessResult{
		Success:        true,
		Message:        fmt.Sprintf("Successfully fetched %d documents from page %d", len(allDocuments), page),
		DocumentsCount: len(allDocuments),
		Documents:      allDocuments,
	}, nil
}

// StoreNFSeDocuments stores NFSe documents using intelligent XML management with deduplication
func (s *NFSeService) StoreNFSeDocuments(ctx context.Context, companyID int64, documents []NFSeDocument) error {
	logger.InfoWithFields("Storing NFSe documents with intelligent deduplication", map[string]any{
		"operation":       "store_nfse_intelligent",
		"company_id":      companyID,
		"documents_count": len(documents),
	})

	if len(documents) == 0 {
		return nil
	}

	// Convert NFSeDocument to XMLDocument for batch processing
	xmlDocuments := make([]XMLDocument, len(documents))
	for i, doc := range documents {
		xmlDocuments[i] = XMLDocument{
			FileName: doc.FileName,
			Content:  doc.XMLContent,
		}
	}

	// Use intelligent XML manager for batch processing
	result, err := s.xmlManager.ProcessBatchXML(ctx, companyID, xmlDocuments)
	if err != nil {
		logger.ErrorWithFields("Failed to process batch XML", err, map[string]any{
			"operation":  "store_nfse_intelligent",
			"company_id": companyID,
		})
		return err
	}

	// Log detailed results
	logger.InfoWithFields("Completed intelligent NFSe document storage", map[string]any{
		"operation":           "store_nfse_intelligent",
		"company_id":          companyID,
		"total_documents":     result.TotalDocuments,
		"processed_documents": result.ProcessedDocuments,
		"duplicate_documents": result.DuplicateDocuments,
		"error_documents":     result.ErrorDocuments,
		"processing_time_ms":  result.ProcessingTime.Milliseconds(),
		"success_rate":        fmt.Sprintf("%.2f%%", float64(result.ProcessedDocuments)/float64(result.TotalDocuments)*100),
	})

	// Log individual results for debugging
	for i, docResult := range result.Results {
		if docResult.Error != nil {
			logger.ErrorWithFields("Document processing failed", docResult.Error, map[string]any{
				"operation":  "store_nfse_intelligent",
				"company_id": companyID,
				"file_name":  documents[i].FileName,
			})
		} else if docResult.IsDuplicate {
			logger.InfoWithFields("Duplicate document detected", map[string]any{
				"operation":        "store_nfse_intelligent",
				"company_id":       companyID,
				"file_name":        documents[i].FileName,
				"existing_id":      docResult.DocumentID,
				"duplicate_reason": docResult.DuplicateReason,
			})
		} else if docResult.Success {
			logger.InfoWithFields("Document processed successfully", map[string]any{
				"operation":   "store_nfse_intelligent",
				"company_id":  companyID,
				"file_name":   documents[i].FileName,
				"document_id": docResult.DocumentID,
			})
		}
	}

	return nil
}
