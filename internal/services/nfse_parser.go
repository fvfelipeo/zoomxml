package services

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"github.com/zoomxml/internal/logger"
	"github.com/zoomxml/internal/models"
)

// NFSeXMLStructure represents the complete NFSe XML structure
type NFSeXMLStructure struct {
	XMLName   xml.Name  `xml:"consultarNotaResponse"`
	ListaNfse ListaNfse `xml:"ListaNfse"`
}

type ListaNfse struct {
	ComplNfse ComplNfse `xml:"ComplNfse"`
}

type ComplNfse struct {
	Nfse             Nfse             `xml:"Nfse"`
	NfseCancelamento NfseCancelamento `xml:"NfseCancelamento"`
	NfseSubstituicao NfseSubstituicao `xml:"NfseSubstituicao"`
}

type Nfse struct {
	InfNfse InfNfse `xml:"InfNfse"`
}

type InfNfse struct {
	Numero                     string           `xml:"Numero"`
	CodigoVerificacao          string           `xml:"CodigoVerificacao"`
	AssinaturaPrestadorTomador string           `xml:"AssinaturaPrestadorTomador"`
	DataEmissao                string           `xml:"DataEmissao"`
	IdentificacaoRps           IdentificacaoRps `xml:"IdentificacaoRps"`
	DataEmissaoRps             string           `xml:"DataEmissaoRps"`
	NaturezaOperacao           string           `xml:"NaturezaOperacao"`
	OptanteSimplesNacional     string           `xml:"OptanteSimplesNacional"`
	Competencia                string           `xml:"Competencia"`
	OutrasInformacoes          string           `xml:"OutrasInformacoes"`
	Servico                    Servico          `xml:"Servico"`
	PrestadorServico           PrestadorServico `xml:"PrestadorServico"`
	TomadorServico             TomadorServico   `xml:"TomadorServico"`
}

type IdentificacaoRps struct {
	Numero string `xml:"Numero"`
	Serie  string `xml:"Serie"`
	Tipo   string `xml:"Tipo"`
}

type Servico struct {
	Valores          Valores `xml:"Valores"`
	ItemListaServico string  `xml:"ItemListaServico"`
	CodigoCnae       string  `xml:"CodigoCnae"`
	Discriminacao    string  `xml:"Discriminacao"`
	CodigoMunicipio  string  `xml:"CodigoMunicipio"`
	IBGE             string  `xml:"IBGE"`
	TOM              string  `xml:"TOM"`
}

type Valores struct {
	ValorServicos          string `xml:"ValorServicos"`
	ValorDeducoes          string `xml:"ValorDeducoes"`
	ValorPis               string `xml:"ValorPis"`
	ValorCofins            string `xml:"ValorCofins"`
	ValorInss              string `xml:"ValorInss"`
	ValorIr                string `xml:"ValorIr"`
	ValorCsll              string `xml:"ValorCsll"`
	IssRetido              string `xml:"IssRetido"`
	ValorIss               string `xml:"ValorIss"`
	OutrasRetencoes        string `xml:"OutrasRetencoes"`
	BaseCalculo            string `xml:"BaseCalculo"`
	Aliquota               string `xml:"Aliquota"`
	ValorLiquidoNfse       string `xml:"ValorLiquidoNfse"`
	DescontoCondicionado   string `xml:"DescontoCondicionado"`
	DescontoIncondicionado string `xml:"DescontoIncondicionado"`
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
	TOM             string `xml:"TOM"`
	Cep             string `xml:"Cep"`
}

type NfseCancelamento struct {
	Confirmacao Confirmacao `xml:"Confirmacao"`
}

type Confirmacao struct {
	Pedido                     Pedido                     `xml:"Pedido"`
	InfConfirmacaoCancelamento InfConfirmacaoCancelamento `xml:"InfConfirmacaoCancelamento"`
}

type Pedido struct {
	InfPedidoCancelamento InfPedidoCancelamento `xml:"InfPedidoCancelamento"`
}

type InfPedidoCancelamento struct {
	IdentificacaoNfse string `xml:"IdentificacaoNfse"`
	DataCancelamento  string `xml:"DataCancelamento"`
}

type InfConfirmacaoCancelamento struct {
	Sucesso string `xml:"Sucesso"`
}

type NfseSubstituicao struct {
	SubstituicaoNfse string `xml:"SubstituicaoNfse"`
}

// ParsedNFSeData represents the extracted and parsed NFSe data
type ParsedNFSeData struct {
	Number                string
	VerificationCode      string
	ProviderCNPJ          string
	TakerCNPJ             string
	ServiceValue          float64
	ServiceCode           string
	IssueDate             time.Time
	MunicipalRegistration string
	IsCancelled           bool
	IsSubstituted         bool
	DocumentHash          string
	FullXML               string

	// Additional important fields
	Competence        string
	RpsIssueDate      time.Time
	TakerName         string
	ProviderName      string
	ProviderTradeName string
}

// NFSeParser handles intelligent parsing and deduplication of NFSe XML documents
type NFSeParser struct{}

// NewNFSeParser creates a new NFSe parser instance
func NewNFSeParser() *NFSeParser {
	return &NFSeParser{}
}

// ParseXML parses NFSe XML content and extracts key fields
func (p *NFSeParser) ParseXML(xmlContent string) (*ParsedNFSeData, error) {
	// Validate XML structure
	if strings.TrimSpace(xmlContent) == "" {
		return nil, fmt.Errorf("empty XML content")
	}

	// Handle ISO-8859-1 encoding
	xmlContent = p.convertEncoding(xmlContent)

	var nfseXML NFSeXMLStructure
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))
	decoder.CharsetReader = p.charsetReader

	err := decoder.Decode(&nfseXML)
	if err != nil {
		logger.ErrorWithFields("Failed to parse NFSe XML", err, map[string]any{
			"operation": "parse_nfse_xml",
		})
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	// Extract data from parsed XML
	infNfse := nfseXML.ListaNfse.ComplNfse.Nfse.InfNfse

	// Parse service value
	serviceValue, err := strconv.ParseFloat(infNfse.Servico.Valores.ValorServicos, 64)
	if err != nil {
		logger.WarnWithFields("Failed to parse service value", map[string]any{
			"operation":     "parse_nfse_xml",
			"service_value": infNfse.Servico.Valores.ValorServicos,
		})
		serviceValue = 0
	}

	// Parse issue date
	issueDate, err := time.Parse("2006-01-02 15:04:05", infNfse.DataEmissao)
	if err != nil {
		logger.WarnWithFields("Failed to parse issue date", map[string]any{
			"operation":  "parse_nfse_xml",
			"issue_date": infNfse.DataEmissao,
		})
		issueDate = time.Time{}
	}

	// Get taker CNPJ (could be CNPJ or CPF)
	takerCNPJ := infNfse.TomadorServico.IdentificacaoTomador.CpfCnpj.Cnpj
	if takerCNPJ == "" {
		takerCNPJ = infNfse.TomadorServico.IdentificacaoTomador.CpfCnpj.Cpf
	}

	// Check cancellation status
	isCancelled := nfseXML.ListaNfse.ComplNfse.NfseCancelamento.Confirmacao.InfConfirmacaoCancelamento.Sucesso == "true"

	// Check substitution status
	isSubstituted := nfseXML.ListaNfse.ComplNfse.NfseSubstituicao.SubstituicaoNfse != ""

	// Parse RPS issue date
	rpsIssueDate := time.Time{}
	if infNfse.DataEmissaoRps != "" && strings.TrimSpace(infNfse.DataEmissaoRps) != "" {
		rpsIssueDate, _ = time.Parse("2006-01-02 15:04:05", strings.TrimSpace(infNfse.DataEmissaoRps))
	}

	// Generate document hash for additional validation
	documentHash := p.generateDocumentHash(infNfse.CodigoVerificacao, infNfse.Numero, infNfse.PrestadorServico.IdentificacaoPrestador.Cnpj, infNfse.DataEmissao)

	parsedData := &ParsedNFSeData{
		Number:                infNfse.Numero,
		VerificationCode:      infNfse.CodigoVerificacao,
		ProviderCNPJ:          infNfse.PrestadorServico.IdentificacaoPrestador.Cnpj,
		TakerCNPJ:             takerCNPJ,
		ServiceValue:          serviceValue,
		ServiceCode:           infNfse.Servico.ItemListaServico,
		IssueDate:             issueDate,
		MunicipalRegistration: infNfse.PrestadorServico.IdentificacaoPrestador.InscricaoMunicipal,
		IsCancelled:           isCancelled,
		IsSubstituted:         isSubstituted,
		DocumentHash:          documentHash,
		FullXML:               xmlContent,

		// Additional important fields
		Competence:        infNfse.Competencia,
		RpsIssueDate:      rpsIssueDate,
		TakerName:         infNfse.TomadorServico.RazaoSocial,
		ProviderName:      infNfse.PrestadorServico.RazaoSocial,
		ProviderTradeName: infNfse.PrestadorServico.NomeFantasia,
	}

	logger.InfoWithFields("Successfully parsed NFSe XML", map[string]any{
		"operation":         "parse_nfse_xml",
		"number":            parsedData.Number,
		"verification_code": parsedData.VerificationCode,
		"provider_cnpj":     parsedData.ProviderCNPJ,
		"service_value":     parsedData.ServiceValue,
		"is_cancelled":      parsedData.IsCancelled,
		"is_substituted":    parsedData.IsSubstituted,
	})

	return parsedData, nil
}

// generateDocumentHash creates a hash of critical fields for additional validation
func (p *NFSeParser) generateDocumentHash(verificationCode, number, providerCNPJ, issueDate string) string {
	data := fmt.Sprintf("%s|%s|%s|%s", verificationCode, number, providerCNPJ, issueDate)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ConvertToDocument converts parsed NFSe data to Document model
func (p *NFSeParser) ConvertToDocument(companyID int64, parsedData *ParsedNFSeData, storageKey string) *models.Document {
	return &models.Document{
		CompanyID:             companyID,
		Type:                  "nfse",
		Key:                   fmt.Sprintf("%s_%s", parsedData.ProviderCNPJ, parsedData.Number),
		Number:                parsedData.Number,
		IssueDate:             parsedData.IssueDate,
		Amount:                parsedData.ServiceValue,
		Status:                "processed",
		StorageKey:            storageKey,
		Metadata:              parsedData.FullXML,
		VerificationCode:      parsedData.VerificationCode,
		ProviderCNPJ:          parsedData.ProviderCNPJ,
		TakerCNPJ:             parsedData.TakerCNPJ,
		ServiceValue:          parsedData.ServiceValue,
		ServiceCode:           parsedData.ServiceCode,
		MunicipalRegistration: parsedData.MunicipalRegistration,
		DocumentHash:          parsedData.DocumentHash,
		IsCancelled:           parsedData.IsCancelled,
		IsSubstituted:         parsedData.IsSubstituted,
		ProcessingDate:        time.Now(),

		// Additional important fields
		Competence:        parsedData.Competence,
		RpsIssueDate:      parsedData.RpsIssueDate,
		TakerName:         parsedData.TakerName,
		ProviderName:      parsedData.ProviderName,
		ProviderTradeName: parsedData.ProviderTradeName,
	}
}

// convertEncoding converts ISO-8859-1 encoded XML to UTF-8
func (p *NFSeParser) convertEncoding(xmlContent string) string {
	// Check if content is already UTF-8 or doesn't specify encoding
	if !strings.Contains(xmlContent, "ISO-8859-1") && !strings.Contains(xmlContent, "iso-8859-1") {
		return xmlContent
	}

	// Convert ISO-8859-1 to UTF-8
	reader := transform.NewReader(strings.NewReader(xmlContent), charmap.ISO8859_1.NewDecoder())
	converted, err := io.ReadAll(reader)
	if err != nil {
		logger.WarnWithFields("Failed to convert encoding, using original", map[string]any{
			"operation": "convert_encoding",
			"error":     err.Error(),
		})
		return xmlContent
	}

	// Replace encoding declaration
	result := string(converted)
	result = strings.ReplaceAll(result, `encoding="ISO-8859-1"`, `encoding="UTF-8"`)
	result = strings.ReplaceAll(result, `encoding="iso-8859-1"`, `encoding="UTF-8"`)

	return result
}

// charsetReader handles different character encodings
func (p *NFSeParser) charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "iso-8859-1", "latin1":
		return transform.NewReader(input, charmap.ISO8859_1.NewDecoder()), nil
	case "windows-1252":
		return transform.NewReader(input, charmap.Windows1252.NewDecoder()), nil
	default:
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}
}
