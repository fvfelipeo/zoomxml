package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// CompanyCredential representa credenciais externas de uma empresa
type CompanyCredential struct {
	bun.BaseModel `bun:"table:company_credentials,alias:cc"`

	ID                int64     `bun:"id,pk,autoincrement" json:"id"`
	CompanyID         int64     `bun:"company_id,notnull" json:"company_id"`
	Type              string    `bun:"type,notnull" json:"type"` // ex: 'SEFAZ', 'Prefeitura', 'API_Externa'
	Name              string    `bun:"name,notnull" json:"name"`
	Login             string    `bun:"login" json:"login,omitempty"`
	EncryptedSecret   string    `bun:"encrypted_secret" json:"-"` // Token/senha criptografada - não expor no JSON
	Active            bool      `bun:"active,notnull,default:true" json:"active"`
	CreatedAt         time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt         time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relacionamentos
	Company *Company `bun:"rel:belongs-to,join:company_id=id" json:"company,omitempty"`
}

// SetSecret define o segredo criptografado (implementar criptografia posteriormente)
func (cc *CompanyCredential) SetSecret(secret string) {
	// TODO: Implementar criptografia real
	cc.EncryptedSecret = secret
}

// GetSecret retorna o segredo descriptografado (implementar descriptografia posteriormente)
func (cc *CompanyCredential) GetSecret() string {
	// TODO: Implementar descriptografia real
	return cc.EncryptedSecret
}

// BeforeAppendModel hook para atualizar timestamps
func (cc *CompanyCredential) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		cc.CreatedAt = time.Now()
		cc.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		cc.UpdatedAt = time.Now()
	}
	return nil
}
