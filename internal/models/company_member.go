package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// CompanyMember representa o vínculo entre usuário e empresa (apenas para empresas restritas)
type CompanyMember struct {
	bun.BaseModel `bun:"table:company_members,alias:cm"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID    int64     `bun:"user_id,notnull" json:"user_id"`
	CompanyID int64     `bun:"company_id,notnull" json:"company_id"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relacionamentos
	User    *User    `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Company *Company `bun:"rel:belongs-to,join:company_id=id" json:"company,omitempty"`
}

// BeforeAppendModel hook para atualizar timestamps
func (cm *CompanyMember) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		cm.CreatedAt = time.Now()
		cm.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		cm.UpdatedAt = time.Now()
	}
	return nil
}
