package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/zoomxml/internal/crypto"
)

// CompanyCredential representa credenciais externas de uma empresa
type CompanyCredential struct {
	bun.BaseModel `bun:"table:company_credentials,alias:cc"`

	ID              int64     `bun:"id,pk,autoincrement" json:"id"`
	CompanyID       int64     `bun:"company_id,notnull" json:"company_id"`
	Type            string    `bun:"type,notnull" json:"type"` // ex: 'prefeitura_user_pass', 'prefeitura_token', 'prefeitura_mixed'
	Name            string    `bun:"name,notnull" json:"name"`
	Description     string    `bun:"description" json:"description,omitempty"`
	Login           string    `bun:"login" json:"login,omitempty"`
	Environment     string    `bun:"environment" json:"environment,omitempty"` // production, staging, development
	EncryptedSecret string    `bun:"encrypted_secret" json:"-"`                // Token/senha criptografada - não expor no JSON
	Active          bool      `bun:"active,notnull,default:true" json:"active"`
	CreatedAt       time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relacionamentos
	Company *Company `bun:"rel:belongs-to,join:company_id=id" json:"company,omitempty"`
}

// SetSecret define o segredo criptografado
func (cc *CompanyCredential) SetSecret(secret string) error {
	encrypted, err := crypto.Encrypt(secret)
	if err != nil {
		return err
	}
	cc.EncryptedSecret = encrypted
	return nil
}

// GetSecret retorna o segredo descriptografado
func (cc *CompanyCredential) GetSecret() (string, error) {
	return crypto.Decrypt(cc.EncryptedSecret)
}

// SetCredentialData encrypts and sets credential data based on type
func (cc *CompanyCredential) SetCredentialData(login, password, token string) error {
	encrypted, err := crypto.EncryptCredentialData(cc.Type, login, password, token)
	if err != nil {
		return err
	}
	cc.EncryptedSecret = encrypted
	return nil
}

// GetCredentialData decrypts and returns credential data
func (cc *CompanyCredential) GetCredentialData() (login, password, token string, err error) {
	return crypto.DecryptCredentialData(cc.Type, cc.EncryptedSecret)
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
