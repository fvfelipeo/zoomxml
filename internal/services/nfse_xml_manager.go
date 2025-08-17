package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/logger"
	"github.com/zoomxml/internal/models"
	"github.com/zoomxml/internal/storage"
)

// ProcessingResult represents the result of XML processing
type ProcessingResult struct {
	Success         bool
	DocumentID      int64
	IsDuplicate     bool
	DuplicateReason string
	ProcessingTime  time.Duration
	Error           error
}

// BatchProcessingResult represents the result of batch XML processing
type BatchProcessingResult struct {
	TotalDocuments     int
	ProcessedDocuments int
	DuplicateDocuments int
	ErrorDocuments     int
	ProcessingTime     time.Duration
	Results            []ProcessingResult
	Statistics         map[string]any
}

// NFSeXMLManager handles intelligent XML management with deduplication
type NFSeXMLManager struct {
	parser       *NFSeParser
	deduplicator *NFSeDeduplicator
}

// NewNFSeXMLManager creates a new NFSe XML manager instance
func NewNFSeXMLManager() *NFSeXMLManager {
	return &NFSeXMLManager{
		parser:       NewNFSeParser(),
		deduplicator: NewNFSeDeduplicator(),
	}
}

// generateOrganizedStorageKey creates an organized storage path: year/competence/cnpj/filename
// Example: 2025/012025/34194865000158/filename.xml
func (m *NFSeXMLManager) generateOrganizedStorageKey(parsedData *ParsedNFSeData, fileName string) string {
	// Extract year from issue date
	year := parsedData.IssueDate.Format("2006")

	// Clean competence (remove spaces, slashes, etc.) and format as MMYYYY
	competence := strings.ReplaceAll(parsedData.Competence, "/", "")
	competence = strings.ReplaceAll(competence, " ", "")
	competence = strings.ReplaceAll(competence, ":", "")

	// If competence is in format "DD/MM/YYYY HH:MM:SS", extract MM and YYYY
	if len(competence) >= 8 {
		// Try to parse different formats
		if strings.Contains(parsedData.Competence, "/") {
			parts := strings.Split(parsedData.Competence, "/")
			if len(parts) >= 3 {
				month := strings.TrimSpace(parts[1])
				yearPart := strings.TrimSpace(parts[2])
				if len(yearPart) >= 4 {
					yearPart = yearPart[:4] // Take first 4 characters as year
				}
				if len(month) == 1 {
					month = "0" + month // Pad month with zero
				}
				competence = month + yearPart
			}
		}
	}

	// If competence is still not in MMYYYY format, use issue date
	if len(competence) != 6 {
		competence = parsedData.IssueDate.Format("012006") // MM + YYYY
	}

	// Clean CNPJ (remove dots, slashes, spaces)
	cleanCNPJ := regexp.MustCompile(`[^0-9]`).ReplaceAllString(parsedData.ProviderCNPJ, "")

	// Generate organized path: year/competence/cnpj/filename
	return fmt.Sprintf("nfse/%s/%s/%s/%s", year, competence, cleanCNPJ, fileName)
}

// ProcessSingleXML processes a single NFSe XML document with intelligent deduplication
func (m *NFSeXMLManager) ProcessSingleXML(ctx context.Context, companyID int64, xmlContent, fileName string) (*ProcessingResult, error) {
	startTime := time.Now()

	logger.InfoWithFields("Starting single XML processing", map[string]any{
		"operation":  "process_single_xml",
		"company_id": companyID,
		"file_name":  fileName,
	})

	result := &ProcessingResult{}

	// Step 1: Parse XML content
	parsedData, err := m.parser.ParseXML(xmlContent)
	if err != nil {
		result.Error = fmt.Errorf("failed to parse XML: %v", err)
		result.ProcessingTime = time.Since(startTime)
		logger.ErrorWithFields("Failed to parse XML", err, map[string]any{
			"operation":  "process_single_xml",
			"company_id": companyID,
			"file_name":  fileName,
		})
		return result, nil
	}

	// Step 2: Check for duplicates
	duplicateCheck, err := m.deduplicator.CheckForDuplicates(ctx, companyID, parsedData)
	if err != nil {
		result.Error = fmt.Errorf("failed to check duplicates: %v", err)
		result.ProcessingTime = time.Since(startTime)
		logger.ErrorWithFields("Failed to check duplicates", err, map[string]any{
			"operation":         "process_single_xml",
			"company_id":        companyID,
			"verification_code": parsedData.VerificationCode,
		})
		return result, nil
	}

	if duplicateCheck.IsDuplicate {
		result.IsDuplicate = true
		result.DuplicateReason = duplicateCheck.Reason
		result.DocumentID = duplicateCheck.ExistingDocument.ID
		result.ProcessingTime = time.Since(startTime)

		logger.InfoWithFields("Duplicate document detected", map[string]any{
			"operation":         "process_single_xml",
			"company_id":        companyID,
			"verification_code": parsedData.VerificationCode,
			"existing_id":       duplicateCheck.ExistingDocument.ID,
			"check_method":      duplicateCheck.CheckMethod,
			"reason":            duplicateCheck.Reason,
		})
		return result, nil
	}

	// Step 3: Store XML in MinIO with organized path
	storageKey := m.generateOrganizedStorageKey(parsedData, fileName)
	err = storage.Storage.UploadFile(ctx, "nfse-storage", storageKey, []byte(xmlContent), "application/xml")
	if err != nil {
		result.Error = fmt.Errorf("failed to store XML: %v", err)
		result.ProcessingTime = time.Since(startTime)
		logger.ErrorWithFields("Failed to store XML in MinIO", err, map[string]any{
			"operation":   "process_single_xml",
			"company_id":  companyID,
			"storage_key": storageKey,
		})
		return result, nil
	}

	// Step 4: Convert to document model and save to database
	document := m.parser.ConvertToDocument(companyID, parsedData, storageKey)

	_, err = database.DB.NewInsert().Model(document).Exec(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to save document: %v", err)
		result.ProcessingTime = time.Since(startTime)
		logger.ErrorWithFields("Failed to save document to database", err, map[string]any{
			"operation":         "process_single_xml",
			"company_id":        companyID,
			"verification_code": parsedData.VerificationCode,
		})
		return result, nil
	}

	result.Success = true
	result.DocumentID = document.ID
	result.ProcessingTime = time.Since(startTime)

	logger.InfoWithFields("Successfully processed XML document", map[string]any{
		"operation":         "process_single_xml",
		"company_id":        companyID,
		"document_id":       document.ID,
		"verification_code": parsedData.VerificationCode,
		"processing_time":   result.ProcessingTime.Milliseconds(),
		"storage_key":       storageKey,
	})

	return result, nil
}

// ProcessBatchXML processes multiple NFSe XML documents with optimized batch operations
func (m *NFSeXMLManager) ProcessBatchXML(ctx context.Context, companyID int64, xmlDocuments []XMLDocument) (*BatchProcessingResult, error) {
	startTime := time.Now()

	logger.InfoWithFields("Starting batch XML processing", map[string]any{
		"operation":       "process_batch_xml",
		"company_id":      companyID,
		"documents_count": len(xmlDocuments),
	})

	result := &BatchProcessingResult{
		TotalDocuments: len(xmlDocuments),
		Results:        make([]ProcessingResult, len(xmlDocuments)),
	}

	if len(xmlDocuments) == 0 {
		result.ProcessingTime = time.Since(startTime)
		return result, nil
	}

	// Step 1: Parse all XML documents
	parsedDataList := make([]*ParsedNFSeData, 0, len(xmlDocuments))
	parseErrors := make(map[int]error)

	for i, xmlDoc := range xmlDocuments {
		parsedData, err := m.parser.ParseXML(xmlDoc.Content)
		if err != nil {
			parseErrors[i] = err
			result.Results[i] = ProcessingResult{
				Error: fmt.Errorf("failed to parse XML: %v", err),
			}
			result.ErrorDocuments++
			continue
		}
		parsedDataList = append(parsedDataList, parsedData)
	}

	// Step 2: Batch check for duplicates
	duplicateResults, err := m.deduplicator.BatchCheckForDuplicates(ctx, companyID, parsedDataList)
	if err != nil {
		logger.ErrorWithFields("Failed to batch check duplicates", err, map[string]any{
			"operation":  "process_batch_xml",
			"company_id": companyID,
		})
		return nil, err
	}

	// Step 3: Process non-duplicate documents
	documentsToInsert := make([]*models.Document, 0)
	storageOperations := make([]StorageOperation, 0)

	parsedIndex := 0
	for i, xmlDoc := range xmlDocuments {
		// Skip documents that failed parsing
		if _, hasError := parseErrors[i]; hasError {
			continue
		}

		parsedData := parsedDataList[parsedIndex]
		duplicateCheck := duplicateResults[parsedIndex]
		parsedIndex++

		if duplicateCheck.IsDuplicate {
			result.Results[i] = ProcessingResult{
				IsDuplicate:     true,
				DuplicateReason: duplicateCheck.Reason,
				DocumentID:      duplicateCheck.ExistingDocument.ID,
			}
			result.DuplicateDocuments++
			continue
		}

		// Prepare for storage and database insertion with organized path
		storageKey := m.generateOrganizedStorageKey(parsedData, xmlDoc.FileName)
		document := m.parser.ConvertToDocument(companyID, parsedData, storageKey)

		documentsToInsert = append(documentsToInsert, document)
		storageOperations = append(storageOperations, StorageOperation{
			Key:     storageKey,
			Content: xmlDoc.Content,
			Index:   i,
		})
	}

	// Step 4: Batch upload to MinIO
	err = m.batchUploadToStorage(ctx, storageOperations)
	if err != nil {
		logger.ErrorWithFields("Failed to batch upload to storage", err, map[string]any{
			"operation":  "process_batch_xml",
			"company_id": companyID,
		})
		// Mark storage operations as failed
		for _, op := range storageOperations {
			result.Results[op.Index] = ProcessingResult{
				Error: fmt.Errorf("failed to store XML: %v", err),
			}
			result.ErrorDocuments++
		}
	} else {
		// Step 5: Batch insert to database
		if len(documentsToInsert) > 0 {
			_, err = database.DB.NewInsert().Model(&documentsToInsert).Exec(ctx)
			if err != nil {
				logger.ErrorWithFields("Failed to batch insert documents", err, map[string]any{
					"operation":       "process_batch_xml",
					"company_id":      companyID,
					"documents_count": len(documentsToInsert),
				})
				// Mark all as failed
				for _, op := range storageOperations {
					result.Results[op.Index] = ProcessingResult{
						Error: fmt.Errorf("failed to save document: %v", err),
					}
					result.ErrorDocuments++
				}
			} else {
				// Mark all as successful
				for i, op := range storageOperations {
					result.Results[op.Index] = ProcessingResult{
						Success:    true,
						DocumentID: documentsToInsert[i].ID,
					}
					result.ProcessedDocuments++
				}
			}
		}
	}

	result.ProcessingTime = time.Since(startTime)

	// Generate statistics
	result.Statistics = map[string]any{
		"total_documents":     result.TotalDocuments,
		"processed_documents": result.ProcessedDocuments,
		"duplicate_documents": result.DuplicateDocuments,
		"error_documents":     result.ErrorDocuments,
		"processing_time_ms":  result.ProcessingTime.Milliseconds(),
		"success_rate":        float64(result.ProcessedDocuments) / float64(result.TotalDocuments) * 100,
	}

	logger.InfoWithFields("Completed batch XML processing", result.Statistics)

	return result, nil
}

// XMLDocument represents an XML document to be processed
type XMLDocument struct {
	FileName string
	Content  string
}

// StorageOperation represents a storage operation
type StorageOperation struct {
	Key     string
	Content string
	Index   int
}

// batchUploadToStorage uploads multiple files to storage efficiently
func (m *NFSeXMLManager) batchUploadToStorage(ctx context.Context, operations []StorageOperation) error {
	for _, op := range operations {
		err := storage.Storage.UploadFile(ctx, "nfse-storage", op.Key, []byte(op.Content), "application/xml")
		if err != nil {
			return fmt.Errorf("failed to upload %s: %v", op.Key, err)
		}
	}
	return nil
}

// GetProcessingStatistics returns processing statistics for a company
func (m *NFSeXMLManager) GetProcessingStatistics(ctx context.Context, companyID int64, days int) (map[string]any, error) {
	duplicateStats, err := m.deduplicator.GetDuplicateStatistics(ctx, companyID, days)
	if err != nil {
		return nil, err
	}

	since := time.Now().AddDate(0, 0, -days)

	var processingStats struct {
		TotalProcessed    int64   `bun:"total_processed"`
		AvgProcessingTime float64 `bun:"avg_processing_time"`
		RecentDocuments   int64   `bun:"recent_documents"`
	}

	err = database.DB.NewSelect().
		Model((*models.Document)(nil)).
		ColumnExpr("COUNT(*) as total_processed").
		ColumnExpr("COUNT(*) FILTER (WHERE created_at >= ?) as recent_documents", since).
		Where("company_id = ? AND type = 'nfse'", companyID).
		Scan(ctx, &processingStats)

	if err != nil {
		return nil, fmt.Errorf("failed to get processing statistics: %v", err)
	}

	stats := map[string]any{
		"processing": map[string]any{
			"total_processed":  processingStats.TotalProcessed,
			"recent_documents": processingStats.RecentDocuments,
		},
		"deduplication": duplicateStats,
		"period_days":   days,
	}

	return stats, nil
}
