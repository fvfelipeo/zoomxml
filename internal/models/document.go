package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Document representa um documento (NFS-e, etc.) no sistema
type Document struct {
	bun.BaseModel `bun:"table:documents,alias:d"`

	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	CompanyID  int64     `bun:"company_id,notnull" json:"company_id"`
	Type       string    `bun:"type,notnull" json:"type"` // ex: 'NFSe', 'NFe', 'CTe'
	Key        string    `bun:"key" json:"key,omitempty"` // Chave de acesso do documento
	Number     string    `bun:"number" json:"number,omitempty"`
	Series     string    `bun:"series" json:"series,omitempty"`
	IssueDate  time.Time `bun:"issue_date" json:"issue_date,omitempty"`
	DueDate    time.Time `bun:"due_date" json:"due_date,omitempty"`
	Amount     float64   `bun:"amount" json:"amount,omitempty"`
	Status     string    `bun:"status,notnull,default:'pending'" json:"status"` // 'pending', 'processed', 'error'
	StorageKey string    `bun:"storage_key" json:"storage_key,omitempty"`       // Chave no MinIO/S3
	Hash       string    `bun:"hash" json:"hash,omitempty"`                     // Hash do arquivo para verificação de integridade
	Metadata   string    `bun:"metadata,type:jsonb" json:"metadata,omitempty"`  // Metadados adicionais em JSON

	// NFSe specific fields for intelligent deduplication
	VerificationCode      string    `bun:"verification_code" json:"verification_code,omitempty"`
	ProviderCNPJ          string    `bun:"provider_cnpj" json:"provider_cnpj,omitempty"`
	TakerCNPJ             string    `bun:"taker_cnpj" json:"taker_cnpj,omitempty"`
	ServiceValue          float64   `bun:"service_value" json:"service_value,omitempty"`
	ServiceCode           string    `bun:"service_code" json:"service_code,omitempty"`
	MunicipalRegistration string    `bun:"municipal_registration" json:"municipal_registration,omitempty"`
	DocumentHash          string    `bun:"document_hash" json:"document_hash,omitempty"`
	IsCancelled           bool      `bun:"is_cancelled,default:false" json:"is_cancelled"`
	IsSubstituted         bool      `bun:"is_substituted,default:false" json:"is_substituted"`
	ProcessingDate        time.Time `bun:"processing_date" json:"processing_date,omitempty"`

	// Additional important NFSe fields
	Competence        string    `bun:"competence" json:"competence,omitempty"`
	RpsIssueDate      time.Time `bun:"rps_issue_date" json:"rps_issue_date,omitempty"`
	TakerName         string    `bun:"taker_name" json:"taker_name,omitempty"`
	ProviderName      string    `bun:"provider_name" json:"provider_name,omitempty"`
	ProviderTradeName string    `bun:"provider_trade_name" json:"provider_trade_name,omitempty"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relacionamentos
	Company *Company `bun:"rel:belongs-to,join:company_id=id" json:"company,omitempty"`
}

// IsProcessed verifica se o documento foi processado
func (d *Document) IsProcessed() bool {
	return d.Status == "processed"
}

// HasError verifica se o documento tem erro
func (d *Document) HasError() bool {
	return d.Status == "error"
}

// IsPending verifica se o documento está pendente
func (d *Document) IsPending() bool {
	return d.Status == "pending"
}

// BeforeAppendModel hook para atualizar timestamps
func (d *Document) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		d.CreatedAt = time.Now()
		d.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		d.UpdatedAt = time.Now()
	}
	return nil
}
