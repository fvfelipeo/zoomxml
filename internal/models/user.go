package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// User representa um usuário do sistema
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	Name      string    `bun:"name,notnull" json:"name"`
	Email     string    `bun:"email,unique,notnull" json:"email"`
	Password  string    `bun:"password,notnull" json:"-"`               // Senha para frontend - não expor no JSON
	Token     string    `bun:"token,unique,notnull" json:"-"`           // Token de acesso para API - não expor no JSON
	Role      string    `bun:"role,notnull,default:'user'" json:"role"` // 'admin' ou 'user'
	Active    bool      `bun:"active,notnull,default:true" json:"active"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relacionamentos
	CompanyMembers []CompanyMember `bun:"rel:has-many,join:id=user_id" json:"company_members,omitempty"`
	AuditLogs      []AuditLog      `bun:"rel:has-many,join:id=actor_id" json:"audit_logs,omitempty"`
}

// IsAdmin verifica se o usuário é admin
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// CanAccessCompany verifica se o usuário pode acessar uma empresa
func (u *User) CanAccessCompany(companyID int64, companyRestricted bool) bool {
	// Admins sempre podem acessar todas as empresas
	if u.IsAdmin() {
		return true
	}

	// Se a empresa não é restrita, todos os usuários podem acessar
	if !companyRestricted {
		return true
	}

	// Para empresas restritas, verificar se o usuário é membro
	for _, member := range u.CompanyMembers {
		if member.CompanyID == companyID {
			return true
		}
	}

	return false
}

// BeforeAppendModel hook para atualizar timestamps
func (u *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		u.UpdatedAt = time.Now()
	}
	return nil
}
