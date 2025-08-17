package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Company representa uma empresa no sistema
type Company struct {
	bun.BaseModel `bun:"table:companies,alias:c"`

	ID   int64  `bun:"id,pk,autoincrement" json:"id"`
	Name string `bun:"name,notnull" json:"name"`
	CNPJ string `bun:"cnpj,unique,notnull" json:"cnpj"`

	// Nome fantasia
	TradeName string `bun:"trade_name" json:"trade_name,omitempty"`

	// Endereço completo
	Address    string `bun:"address" json:"address,omitempty"`
	Number     string `bun:"number" json:"number,omitempty"`
	Complement string `bun:"complement" json:"complement,omitempty"`
	District   string `bun:"district" json:"district,omitempty"`
	City       string `bun:"city" json:"city,omitempty"`
	State      string `bun:"state" json:"state,omitempty"`
	ZipCode    string `bun:"zip_code" json:"zip_code,omitempty"`

	// Contato
	Phone string `bun:"phone" json:"phone,omitempty"`
	Email string `bun:"email" json:"email,omitempty"`

	// Dados empresariais
	CompanySize        string    `bun:"company_size" json:"company_size,omitempty"`               // ME, EPP, etc
	MainActivity       string    `bun:"main_activity" json:"main_activity,omitempty"`             // Atividade principal
	SecondaryActivity  string    `bun:"secondary_activity" json:"secondary_activity,omitempty"`   // Atividades secundárias
	LegalNature        string    `bun:"legal_nature" json:"legal_nature,omitempty"`               // Natureza jurídica
	OpeningDate        string    `bun:"opening_date" json:"opening_date,omitempty"`               // Data de abertura
	RegistrationStatus string    `bun:"registration_status" json:"registration_status,omitempty"` // Situação cadastral
	Restricted         bool      `bun:"restricted,notnull,default:false" json:"restricted"`
	AutoFetch          bool      `bun:"auto_fetch,notnull,default:false" json:"auto_fetch"`
	Active             bool      `bun:"active,notnull,default:true" json:"active"`
	CreatedAt          time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt          time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relacionamentos
	Members     []CompanyMember     `bun:"rel:has-many,join:id=company_id" json:"members,omitempty"`
	Credentials []CompanyCredential `bun:"rel:has-many,join:id=company_id" json:"credentials,omitempty"`
	Documents   []Document          `bun:"rel:has-many,join:id=company_id" json:"documents,omitempty"`
}

// IsAccessibleByUser verifica se a empresa é acessível por um usuário
func (c *Company) IsAccessibleByUser(user *User) bool {
	// Admins sempre podem acessar todas as empresas
	if user.IsAdmin() {
		return true
	}

	// Se a empresa não é restrita, todos os usuários podem acessar
	if !c.Restricted {
		return true
	}

	// Para empresas restritas, verificar se o usuário é membro
	for _, member := range c.Members {
		if member.UserID == user.ID {
			return true
		}
	}

	return false
}

// BeforeAppendModel hook para atualizar timestamps
func (c *Company) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		c.CreatedAt = time.Now()
		c.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		c.UpdatedAt = time.Now()
	}
	return nil
}
