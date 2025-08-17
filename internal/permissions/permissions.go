package permissions

import (
	"context"
	"errors"

	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrCompanyNotFound = errors.New("company not found")
	ErrAccessDenied    = errors.New("access denied")
)

// CanAccessCompany checks if a user can access a specific company
func CanAccessCompany(ctx context.Context, user *models.User, companyID int64) error {
	if user == nil {
		return ErrUserNotFound
	}

	// Admin users can access all companies
	if user.IsAdmin() {
		return nil
	}

	// Check if company exists
	company := &models.Company{}
	err := database.DB.NewSelect().
		Model(company).
		Where("id = ? AND active = true", companyID).
		Scan(ctx)

	if err != nil {
		return ErrCompanyNotFound
	}

	// If company is not restricted, any authenticated user can access it
	if !company.Restricted {
		return nil
	}

	// For restricted companies, check if user is a member
	exists, err := database.DB.NewSelect().
		Model((*models.CompanyMember)(nil)).
		Where("user_id = ? AND company_id = ?", user.ID, companyID).
		Exists(ctx)

	if err != nil {
		return err
	}

	if !exists {
		return ErrAccessDenied
	}

	return nil
}

// CanManageCredentials checks if a user can manage credentials for a company
func CanManageCredentials(ctx context.Context, user *models.User, companyID int64) error {
	// For now, credential management has the same permissions as company access
	// This can be extended in the future for more granular permissions
	return CanAccessCompany(ctx, user, companyID)
}

// CanViewCredentials checks if a user can view credentials for a company
func CanViewCredentials(ctx context.Context, user *models.User, companyID int64) error {
	// For now, viewing credentials has the same permissions as company access
	// This can be extended in the future for more granular permissions
	return CanAccessCompany(ctx, user, companyID)
}

// CanCreateCredentials checks if a user can create credentials for a company
func CanCreateCredentials(ctx context.Context, user *models.User, companyID int64) error {
	// For now, creating credentials has the same permissions as company access
	// This can be extended in the future for more granular permissions
	return CanAccessCompany(ctx, user, companyID)
}

// CanUpdateCredentials checks if a user can update credentials for a company
func CanUpdateCredentials(ctx context.Context, user *models.User, companyID int64) error {
	// For now, updating credentials has the same permissions as company access
	// This can be extended in the future for more granular permissions
	return CanAccessCompany(ctx, user, companyID)
}

// CanDeleteCredentials checks if a user can delete credentials for a company
func CanDeleteCredentials(ctx context.Context, user *models.User, companyID int64) error {
	// For now, deleting credentials has the same permissions as company access
	// This can be extended in the future for more granular permissions
	return CanAccessCompany(ctx, user, companyID)
}

// GetAccessibleCompanies returns a list of company IDs that the user can access
func GetAccessibleCompanies(ctx context.Context, user *models.User) ([]int64, error) {
	if user == nil {
		return nil, ErrUserNotFound
	}

	var companyIDs []int64

	// Admin users can access all companies
	if user.IsAdmin() {
		err := database.DB.NewSelect().
			Model((*models.Company)(nil)).
			Column("id").
			Where("active = true").
			Scan(ctx, &companyIDs)
		return companyIDs, err
	}

	// Get all non-restricted companies
	var publicCompanyIDs []int64
	err := database.DB.NewSelect().
		Model((*models.Company)(nil)).
		Column("id").
		Where("active = true AND restricted = false").
		Scan(ctx, &publicCompanyIDs)

	if err != nil {
		return nil, err
	}

	// Get restricted companies where user is a member
	var memberCompanyIDs []int64
	err = database.DB.NewSelect().
		Model((*models.CompanyMember)(nil)).
		Column("company_id").
		Where("user_id = ?", user.ID).
		Scan(ctx, &memberCompanyIDs)

	if err != nil {
		return nil, err
	}

	// Combine both lists
	companyIDs = append(companyIDs, publicCompanyIDs...)
	companyIDs = append(companyIDs, memberCompanyIDs...)

	return companyIDs, nil
}

// ValidateCredentialAccess validates that a user can access a specific credential
func ValidateCredentialAccess(ctx context.Context, user *models.User, credentialID, companyID int64) error {
	// First check if user can access the company
	err := CanAccessCompany(ctx, user, companyID)
	if err != nil {
		return err
	}

	// Verify that the credential belongs to the specified company
	exists, err := database.DB.NewSelect().
		Model((*models.CompanyCredential)(nil)).
		Where("id = ? AND company_id = ?", credentialID, companyID).
		Exists(ctx)

	if err != nil {
		return err
	}

	if !exists {
		return errors.New("credential not found or does not belong to the specified company")
	}

	return nil
}
