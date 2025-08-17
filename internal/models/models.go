package models

import (
	"github.com/uptrace/bun"
)

// RegisterModels registra todos os modelos no banco de dados
func RegisterModels(db *bun.DB) {
	db.RegisterModel(
		(*User)(nil),
		(*Company)(nil),
		(*CompanyMember)(nil),
		(*CompanyCredential)(nil),
		(*Document)(nil),
		(*AuditLog)(nil),
	)
}

// GetAllModels retorna uma lista de todos os modelos para migrações
func GetAllModels() []interface{} {
	return []interface{}{
		(*User)(nil),
		(*Company)(nil),
		(*CompanyMember)(nil),
		(*CompanyCredential)(nil),
		(*Document)(nil),
		(*AuditLog)(nil),
	}
}
