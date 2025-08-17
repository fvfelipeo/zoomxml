package services

import (
	"context"
	"fmt"
	"time"

	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/logger"
	"github.com/zoomxml/internal/models"
)

// DuplicateCheckResult represents the result of duplicate detection
type DuplicateCheckResult struct {
	IsDuplicate      bool
	ExistingDocument *models.Document
	CheckMethod      string
	Reason           string
}

// NFSeDeduplicator handles intelligent duplicate detection for NFSe documents
type NFSeDeduplicator struct{}

// NewNFSeDeduplicator creates a new NFSe deduplicator instance
func NewNFSeDeduplicator() *NFSeDeduplicator {
	return &NFSeDeduplicator{}
}

// CheckForDuplicates performs comprehensive duplicate detection using multiple strategies
func (d *NFSeDeduplicator) CheckForDuplicates(ctx context.Context, companyID int64, parsedData *ParsedNFSeData) (*DuplicateCheckResult, error) {
	logger.DebugWithFields("Starting duplicate check", map[string]any{
		"operation":         "check_duplicates",
		"company_id":        companyID,
		"verification_code": parsedData.VerificationCode,
		"number":            parsedData.Number,
		"provider_cnpj":     parsedData.ProviderCNPJ,
	})

	// Strategy 1: Primary check by verification code (most reliable)
	if parsedData.VerificationCode != "" {
		result, err := d.checkByVerificationCode(ctx, companyID, parsedData.VerificationCode)
		if err != nil {
			return nil, err
		}
		if result.IsDuplicate {
			logger.InfoWithFields("Duplicate found by verification code", map[string]any{
				"operation":         "check_duplicates",
				"company_id":        companyID,
				"verification_code": parsedData.VerificationCode,
				"existing_id":       result.ExistingDocument.ID,
			})
			return result, nil
		}
	}

	// Strategy 2: Secondary check by NFSe number + provider CNPJ + issue date
	result, err := d.checkByCompositeKey(ctx, companyID, parsedData)
	if err != nil {
		return nil, err
	}
	if result.IsDuplicate {
		logger.InfoWithFields("Duplicate found by composite key", map[string]any{
			"operation":     "check_duplicates",
			"company_id":    companyID,
			"number":        parsedData.Number,
			"provider_cnpj": parsedData.ProviderCNPJ,
			"existing_id":   result.ExistingDocument.ID,
		})
		return result, nil
	}

	// Strategy 3: Tertiary check by document hash
	if parsedData.DocumentHash != "" {
		result, err := d.checkByDocumentHash(ctx, companyID, parsedData.DocumentHash)
		if err != nil {
			return nil, err
		}
		if result.IsDuplicate {
			logger.InfoWithFields("Duplicate found by document hash", map[string]any{
				"operation":     "check_duplicates",
				"company_id":    companyID,
				"document_hash": parsedData.DocumentHash,
				"existing_id":   result.ExistingDocument.ID,
			})
			return result, nil
		}
	}

	// No duplicates found
	logger.DebugWithFields("No duplicates found", map[string]any{
		"operation":         "check_duplicates",
		"company_id":        companyID,
		"verification_code": parsedData.VerificationCode,
		"number":            parsedData.Number,
	})

	return &DuplicateCheckResult{
		IsDuplicate: false,
		CheckMethod: "comprehensive",
		Reason:      "no duplicates found",
	}, nil
}

// checkByVerificationCode checks for duplicates using verification code (primary key)
func (d *NFSeDeduplicator) checkByVerificationCode(ctx context.Context, companyID int64, verificationCode string) (*DuplicateCheckResult, error) {
	var existingDoc models.Document
	
	err := database.DB.NewSelect().
		Model(&existingDoc).
		Where("company_id = ? AND verification_code = ? AND verification_code != ''", companyID, verificationCode).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return &DuplicateCheckResult{
				IsDuplicate: false,
				CheckMethod: "verification_code",
				Reason:      "no matching verification code",
			}, nil
		}
		return nil, fmt.Errorf("failed to check verification code: %v", err)
	}

	return &DuplicateCheckResult{
		IsDuplicate:      true,
		ExistingDocument: &existingDoc,
		CheckMethod:      "verification_code",
		Reason:           fmt.Sprintf("matching verification code: %s", verificationCode),
	}, nil
}

// checkByCompositeKey checks for duplicates using NFSe number + provider CNPJ + issue date
func (d *NFSeDeduplicator) checkByCompositeKey(ctx context.Context, companyID int64, parsedData *ParsedNFSeData) (*DuplicateCheckResult, error) {
	var existingDoc models.Document
	
	// Format issue date for comparison (ignore time component for date matching)
	issueDate := parsedData.IssueDate.Format("2006-01-02")
	
	err := database.DB.NewSelect().
		Model(&existingDoc).
		Where("company_id = ? AND number = ? AND provider_cnpj = ? AND DATE(issue_date) = ?", 
			companyID, parsedData.Number, parsedData.ProviderCNPJ, issueDate).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return &DuplicateCheckResult{
				IsDuplicate: false,
				CheckMethod: "composite_key",
				Reason:      "no matching composite key",
			}, nil
		}
		return nil, fmt.Errorf("failed to check composite key: %v", err)
	}

	return &DuplicateCheckResult{
		IsDuplicate:      true,
		ExistingDocument: &existingDoc,
		CheckMethod:      "composite_key",
		Reason:           fmt.Sprintf("matching number: %s, provider: %s, date: %s", parsedData.Number, parsedData.ProviderCNPJ, issueDate),
	}, nil
}

// checkByDocumentHash checks for duplicates using document hash
func (d *NFSeDeduplicator) checkByDocumentHash(ctx context.Context, companyID int64, documentHash string) (*DuplicateCheckResult, error) {
	var existingDoc models.Document
	
	err := database.DB.NewSelect().
		Model(&existingDoc).
		Where("company_id = ? AND document_hash = ? AND document_hash != ''", companyID, documentHash).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return &DuplicateCheckResult{
				IsDuplicate: false,
				CheckMethod: "document_hash",
				Reason:      "no matching document hash",
			}, nil
		}
		return nil, fmt.Errorf("failed to check document hash: %v", err)
	}

	return &DuplicateCheckResult{
		IsDuplicate:      true,
		ExistingDocument: &existingDoc,
		CheckMethod:      "document_hash",
		Reason:           fmt.Sprintf("matching document hash: %s", documentHash),
	}, nil
}

// BatchCheckForDuplicates performs duplicate detection for multiple documents efficiently
func (d *NFSeDeduplicator) BatchCheckForDuplicates(ctx context.Context, companyID int64, parsedDataList []*ParsedNFSeData) (map[int]*DuplicateCheckResult, error) {
	results := make(map[int]*DuplicateCheckResult)
	
	if len(parsedDataList) == 0 {
		return results, nil
	}

	logger.InfoWithFields("Starting batch duplicate check", map[string]any{
		"operation":      "batch_check_duplicates",
		"company_id":     companyID,
		"documents_count": len(parsedDataList),
	})

	// Collect all verification codes and numbers for batch query
	verificationCodes := make([]string, 0, len(parsedDataList))
	numbers := make([]string, 0, len(parsedDataList))
	documentHashes := make([]string, 0, len(parsedDataList))

	for _, data := range parsedDataList {
		if data.VerificationCode != "" {
			verificationCodes = append(verificationCodes, data.VerificationCode)
		}
		if data.Number != "" {
			numbers = append(numbers, data.Number)
		}
		if data.DocumentHash != "" {
			documentHashes = append(documentHashes, data.DocumentHash)
		}
	}

	// Batch query for existing documents
	var existingDocs []models.Document
	query := database.DB.NewSelect().
		Model(&existingDocs).
		Where("company_id = ?", companyID)

	if len(verificationCodes) > 0 {
		query = query.WhereOr("verification_code IN (?)", verificationCodes)
	}
	if len(numbers) > 0 {
		query = query.WhereOr("number IN (?)", numbers)
	}
	if len(documentHashes) > 0 {
		query = query.WhereOr("document_hash IN (?)", documentHashes)
	}

	err := query.Scan(ctx)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return nil, fmt.Errorf("failed to batch check duplicates: %v", err)
	}

	// Create lookup maps for efficient matching
	verificationCodeMap := make(map[string]*models.Document)
	compositeKeyMap := make(map[string]*models.Document)
	documentHashMap := make(map[string]*models.Document)

	for i := range existingDocs {
		doc := &existingDocs[i]
		if doc.VerificationCode != "" {
			verificationCodeMap[doc.VerificationCode] = doc
		}
		if doc.Number != "" && doc.ProviderCNPJ != "" {
			compositeKey := fmt.Sprintf("%s|%s|%s", doc.Number, doc.ProviderCNPJ, doc.IssueDate.Format("2006-01-02"))
			compositeKeyMap[compositeKey] = doc
		}
		if doc.DocumentHash != "" {
			documentHashMap[doc.DocumentHash] = doc
		}
	}

	// Check each document for duplicates
	for i, data := range parsedDataList {
		// Check by verification code first
		if data.VerificationCode != "" {
			if existingDoc, exists := verificationCodeMap[data.VerificationCode]; exists {
				results[i] = &DuplicateCheckResult{
					IsDuplicate:      true,
					ExistingDocument: existingDoc,
					CheckMethod:      "verification_code",
					Reason:           fmt.Sprintf("matching verification code: %s", data.VerificationCode),
				}
				continue
			}
		}

		// Check by composite key
		compositeKey := fmt.Sprintf("%s|%s|%s", data.Number, data.ProviderCNPJ, data.IssueDate.Format("2006-01-02"))
		if existingDoc, exists := compositeKeyMap[compositeKey]; exists {
			results[i] = &DuplicateCheckResult{
				IsDuplicate:      true,
				ExistingDocument: existingDoc,
				CheckMethod:      "composite_key",
				Reason:           fmt.Sprintf("matching composite key: %s", compositeKey),
			}
			continue
		}

		// Check by document hash
		if data.DocumentHash != "" {
			if existingDoc, exists := documentHashMap[data.DocumentHash]; exists {
				results[i] = &DuplicateCheckResult{
					IsDuplicate:      true,
					ExistingDocument: existingDoc,
					CheckMethod:      "document_hash",
					Reason:           fmt.Sprintf("matching document hash: %s", data.DocumentHash),
				}
				continue
			}
		}

		// No duplicate found
		results[i] = &DuplicateCheckResult{
			IsDuplicate: false,
			CheckMethod: "batch_comprehensive",
			Reason:      "no duplicates found",
		}
	}

	duplicateCount := 0
	for _, result := range results {
		if result.IsDuplicate {
			duplicateCount++
		}
	}

	logger.InfoWithFields("Completed batch duplicate check", map[string]any{
		"operation":        "batch_check_duplicates",
		"company_id":       companyID,
		"documents_count":  len(parsedDataList),
		"duplicates_found": duplicateCount,
	})

	return results, nil
}

// GetDuplicateStatistics returns statistics about duplicate detection
func (d *NFSeDeduplicator) GetDuplicateStatistics(ctx context.Context, companyID int64, days int) (map[string]any, error) {
	since := time.Now().AddDate(0, 0, -days)
	
	var stats struct {
		TotalDocuments    int64 `bun:"total_documents"`
		UniqueDocuments   int64 `bun:"unique_documents"`
		CancelledDocuments int64 `bun:"cancelled_documents"`
		SubstitutedDocuments int64 `bun:"substituted_documents"`
	}

	err := database.DB.NewSelect().
		Model((*models.Document)(nil)).
		ColumnExpr("COUNT(*) as total_documents").
		ColumnExpr("COUNT(DISTINCT verification_code) as unique_documents").
		ColumnExpr("COUNT(*) FILTER (WHERE is_cancelled = true) as cancelled_documents").
		ColumnExpr("COUNT(*) FILTER (WHERE is_substituted = true) as substituted_documents").
		Where("company_id = ? AND type = 'nfse' AND created_at >= ?", companyID, since).
		Scan(ctx, &stats)

	if err != nil {
		return nil, fmt.Errorf("failed to get duplicate statistics: %v", err)
	}

	return map[string]any{
		"total_documents":      stats.TotalDocuments,
		"unique_documents":     stats.UniqueDocuments,
		"cancelled_documents":  stats.CancelledDocuments,
		"substituted_documents": stats.SubstitutedDocuments,
		"potential_duplicates": stats.TotalDocuments - stats.UniqueDocuments,
		"period_days":          days,
	}, nil
}
