package models

import (
	"fmt"
	"time"
)

// ConsultarNotaResponse represents the XML response structure
type ConsultarNotaResponse struct {
	XMLName   string    `xml:"consultarNotaResponse"`
	ListaNfse ListaNfse `xml:"ListaNfse"`
}

type ListaNfse struct {
	ComplNfse ComplNfse `xml:"ComplNfse"`
}

type ComplNfse struct {
	Nfse Nfse `xml:"Nfse"`
}

type Nfse struct {
	InfNfse InfNfse `xml:"InfNfse"`
}

type InfNfse struct {
	Numero            string           `xml:"Numero"`
	CodigoVerificacao string           `xml:"CodigoVerificacao"`
	DataEmissao       string           `xml:"DataEmissao"`
	Competencia       string           `xml:"Competencia"`
	NaturezaOperacao  string           `xml:"NaturezaOperacao"`
	PrestadorServico  PrestadorServico `xml:"PrestadorServico"`
	TomadorServico    TomadorServico   `xml:"TomadorServico"`
	Servico           Servico          `xml:"Servico"`
	OutrasInformacoes string           `xml:"OutrasInformacoes"`
}

type PrestadorServico struct {
	IdentificacaoPrestador IdentificacaoPrestador `xml:"IdentificacaoPrestador"`
	RazaoSocial            string                 `xml:"RazaoSocial"`
	NomeFantasia           string                 `xml:"NomeFantasia"`
	Endereco               Endereco               `xml:"Endereco"`
}

type IdentificacaoPrestador struct {
	Cnpj               string `xml:"Cnpj"`
	InscricaoMunicipal string `xml:"InscricaoMunicipal"`
}

type TomadorServico struct {
	IdentificacaoTomador IdentificacaoTomador `xml:"IdentificacaoTomador"`
	RazaoSocial          string               `xml:"RazaoSocial"`
	Endereco             Endereco             `xml:"Endereco"`
}

type IdentificacaoTomador struct {
	CpfCnpj CpfCnpj `xml:"CpfCnpj"`
}

type CpfCnpj struct {
	Cnpj string `xml:"Cnpj"`
	Cpf  string `xml:"Cpf"`
}

type Endereco struct {
	Endereco        string `xml:"Endereco"`
	Numero          string `xml:"Numero"`
	Complemento     string `xml:"Complemento"`
	Bairro          string `xml:"Bairro"`
	CodigoMunicipio string `xml:"CodigoMunicipio"`
	IBGE            string `xml:"IBGE"`
	Cep             string `xml:"Cep"`
}

type Servico struct {
	Valores          Valores `xml:"Valores"`
	ItemListaServico string  `xml:"ItemListaServico"`
	CodigoCnae       string  `xml:"CodigoCnae"`
	Discriminacao    string  `xml:"Discriminacao"`
	CodigoMunicipio  string  `xml:"CodigoMunicipio"`
}

type Valores struct {
	ValorServicos    string `xml:"ValorServicos"`
	ValorDeducoes    string `xml:"ValorDeducoes"`
	ValorIss         string `xml:"ValorIss"`
	ValorLiquidoNfse string `xml:"ValorLiquidoNfse"`
	BaseCalculo      string `xml:"BaseCalculo"`
	Aliquota         string `xml:"Aliquota"`
	IssRetido        string `xml:"IssRetido"`
}

// NFSeMetadata represents comprehensive NFS-e metadata
type NFSeMetadata struct {
	ID                    int       `json:"id" db:"id"`
	UUID                  string    `json:"uuid" db:"uuid"`
	NumeroNFSe            string    `json:"numero_nfse" db:"numero_nfse"`
	ContentHash           string    `json:"content_hash" db:"content_hash"`
	FilePath              string    `json:"file_path" db:"file_path"`
	SourceZipFile         string    `json:"source_zip_file" db:"source_zip_file"`
	DataEmissao           time.Time `json:"data_emissao" db:"data_emissao"`
	Competencia           string    `json:"competencia" db:"competencia"`
	CompetenciaFormatada  string    `json:"competencia_formatada" db:"competencia_formatada"`
	PrestadorCNPJ         string    `json:"prestador_cnpj" db:"prestador_cnpj"`
	PrestadorRazao        string    `json:"prestador_razao" db:"prestador_razao"`
	PrestadorNomeFantasia string    `json:"prestador_nome_fantasia" db:"prestador_nome_fantasia"`
	TomadorCNPJ           string    `json:"tomador_cnpj" db:"tomador_cnpj"`
	TomadorCPF            string    `json:"tomador_cpf" db:"tomador_cpf"`
	TomadorRazao          string    `json:"tomador_razao" db:"tomador_razao"`
	ValorServicos         float64   `json:"valor_servicos" db:"valor_servicos"`
	ValorISS              float64   `json:"valor_iss" db:"valor_iss"`
	ValorLiquido          float64   `json:"valor_liquido" db:"valor_liquido"`
	Aliquota              float64   `json:"aliquota" db:"aliquota"`
	BaseCalculo           float64   `json:"base_calculo" db:"base_calculo"`
	CodigoVerificacao     string    `json:"codigo_verificacao" db:"codigo_verificacao"`
	NaturezaOperacao      int       `json:"natureza_operacao" db:"natureza_operacao"`
	ItemListaServico      string    `json:"item_lista_servico" db:"item_lista_servico"`
	CodigoCNAE            string    `json:"codigo_cnae" db:"codigo_cnae"`
	Discriminacao         string    `json:"discriminacao" db:"discriminacao"`
	FileSize              int64     `json:"file_size" db:"file_size"`
	ProcessedAt           time.Time `json:"processed_at" db:"processed_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
	Version               int       `json:"version" db:"version"`
	Status                string    `json:"status" db:"status"`
	ErrorMessage          string    `json:"error_message,omitempty" db:"error_message"`
}

// ProcessingLog represents processing history
type ProcessingLog struct {
	ID          int       `json:"id" db:"id"`
	UUID        string    `json:"uuid" db:"uuid"`
	Operation   string    `json:"operation" db:"operation"`
	SourceFile  string    `json:"source_file" db:"source_file"`
	TargetFile  string    `json:"target_file" db:"target_file"`
	Status      string    `json:"status" db:"status"`
	Message     string    `json:"message" db:"message"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
	Duration    int64     `json:"duration_ms" db:"duration_ms"`
	BatchID     string    `json:"batch_id" db:"batch_id"`
}

// PrestadorCache represents cached prestador information
type PrestadorCache struct {
	CNPJ               string    `json:"cnpj" db:"cnpj"`
	RazaoSocial        string    `json:"razao_social" db:"razao_social"`
	NomeFantasia       string    `json:"nome_fantasia" db:"nome_fantasia"`
	InscricaoMunicipal string    `json:"inscricao_municipal" db:"inscricao_municipal"`
	FirstSeen          time.Time `json:"first_seen" db:"first_seen"`
	LastSeen           time.Time `json:"last_seen" db:"last_seen"`
	NFSeCount          int       `json:"nfse_count" db:"nfse_count"`
	TotalValorServicos float64   `json:"total_valor_servicos" db:"total_valor_servicos"`
	TotalValorISS      float64   `json:"total_valor_iss" db:"total_valor_iss"`
	Status             string    `json:"status" db:"status"`
}

// ProcessingStats tracks processing statistics
type ProcessingStats struct {
	TotalFiles     int
	ProcessedFiles int
	NewFiles       int
	UpdatedFiles   int
	DuplicateFiles int
	ErrorFiles     int
	StartTime      time.Time
	BatchID        string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ConsultaXMLRequest - Parâmetros para consulta de XML
type ConsultaXMLRequest struct {
	NumeroInicial string `json:"nr_inicial,omitempty"`     // Número inicial da NFS-e
	NumeroFinal   string `json:"nr_final,omitempty"`       // Número final da NFS-e
	DataInicial   string `json:"dt_inicial,omitempty"`     // Data inicial (YYYY-MM-DD)
	DataFinal     string `json:"dt_final,omitempty"`       // Data final (YYYY-MM-DD)
	Competencia   string `json:"nr_competencia,omitempty"` // Competência (YYYY MM)
	Pagina        int    `json:"nr_page,omitempty"`        // Número da página (padrão: 1)
}

// NFSeXMLResponse - Resposta da consulta XML
type NFSeXMLResponse struct {
	Success        bool          `json:"success"`
	Message        *ErrorMessage `json:"Message,omitempty"`
	RecordCount    int           `json:"RecordCount,omitempty"`
	RecordsPerPage int           `json:"RecordsPerPage,omitempty"`
	PageCount      int           `json:"PageCount,omitempty"`
	CurrentPage    int           `json:"CurrentPage,omitempty"`
	Dados          []NFSeXMLItem `json:"Dados,omitempty"`
}

// NFSeXMLItem - Item individual de NFS-e
type NFSeXMLItem struct {
	NrNfse        int    `json:"NrNfse"`
	DtEmissao     string `json:"DtEmissao"`
	NrCompetencia int    `json:"NrCompetencia"`
	XmlCompactado string `json:"XmlCompactado"`
}

// ErrorMessage - Estrutura de erro da API
type ErrorMessage struct {
	Kind        string   `json:"Kind"`
	Code        string   `json:"Code"`
	Message     string   `json:"Message"`
	Detail      string   `json:"Detail"`
	DetailError []string `json:"DetailError,omitempty"`
}

// FormatarResposta - Formata a resposta para exibição
func (r *NFSeXMLResponse) FormatarResposta() string {
	if !r.Success && r.Message != nil {
		return fmt.Sprintf("❌ Erro: %s - %s\nDetalhes: %s",
			r.Message.Code, r.Message.Message, r.Message.Detail)
	}

	if r.Success && len(r.Dados) == 0 {
		return "ℹ️  Nenhuma NFS-e encontrada no período especificado"
	}

	result := "✅ Consulta realizada com sucesso!\n"
	result += fmt.Sprintf("📄 Total de registros: %d\n", r.RecordCount)
	result += fmt.Sprintf("📑 Página atual: %d de %d\n", r.CurrentPage, r.PageCount)
	result += fmt.Sprintf("📋 Registros por página: %d\n", r.RecordsPerPage)

	if len(r.Dados) > 0 {
		result += fmt.Sprintf("\n📋 NFS-e encontradas:\n")
		for i, nfse := range r.Dados {
			result += fmt.Sprintf("  %d. NFS-e: %d | Data: %s | Competência: %d\n",
				i+1, nfse.NrNfse, nfse.DtEmissao, nfse.NrCompetencia)
			if nfse.XmlCompactado != "" {
				result += fmt.Sprintf("      📦 XML: %d caracteres\n", len(nfse.XmlCompactado))
			}
		}
	}

	return result
}

// GetXMLCompactado - Retorna o XML compactado se disponível
func (r *NFSeXMLResponse) GetXMLCompactado() string {
	// Verificar nos itens individuais
	if len(r.Dados) > 0 {
		var xmls []string
		for _, nfse := range r.Dados {
			if nfse.XmlCompactado != "" {
				xmls = append(xmls, nfse.XmlCompactado)
			}
		}
		if len(xmls) > 0 {
			// Retornar o primeiro XML ou concatenar todos
			return xmls[0]
		}
	}

	return ""
}

// GetAllXMLCompactado - Retorna todos os XMLs compactados
func (r *NFSeXMLResponse) GetAllXMLCompactado() []string {
	var xmls []string
	for _, nfse := range r.Dados {
		if nfse.XmlCompactado != "" {
			xmls = append(xmls, nfse.XmlCompactado)
		}
	}
	return xmls
}
