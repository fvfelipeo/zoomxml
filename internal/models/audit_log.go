package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// AuditLog representa um log de auditoria do sistema
type AuditLog struct {
	bun.BaseModel `bun:"table:audit_logs,alias:al"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	ActorID   int64     `bun:"actor_id,notnull" json:"actor_id"` // ID do usuário que executou a ação
	Action    string    `bun:"action,notnull" json:"action"` // ex: 'CREATE', 'UPDATE', 'DELETE'
	Entity    string    `bun:"entity,notnull" json:"entity"` // ex: 'User', 'Company', 'Document'
	EntityID  int64     `bun:"entity_id" json:"entity_id,omitempty"` // ID da entidade afetada
	Details   string    `bun:"details,type:jsonb" json:"details,omitempty"` // Detalhes da ação em JSON
	IPAddress string    `bun:"ip_address" json:"ip_address,omitempty"`
	UserAgent string    `bun:"user_agent" json:"user_agent,omitempty"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`

	// Relacionamentos
	Actor *User `bun:"rel:belongs-to,join:actor_id=id" json:"actor,omitempty"`
}

// BeforeAppendModel hook para definir timestamp
func (al *AuditLog) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		al.CreatedAt = time.Now()
	}
	return nil
}
