package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

var reDigits = regexp.MustCompile(`\D`)

// CNPJData representa os dados retornados pela API do CNPJá
type CNPJData struct {
	CNPJ                string   `json:"cnpj"`
	Name                string   `json:"name"`
	TradeName           string   `json:"trade_name"`
	Address             string   `json:"address"`
	Number              string   `json:"number"`
	Complement          string   `json:"complement"`
	District            string   `json:"district"`
	City                string   `json:"city"`
	State               string   `json:"state"`
	ZipCode             string   `json:"zip_code"`
	Phone               string   `json:"phone"`
	Email               string   `json:"email"`
	CompanySize         string   `json:"company_size"`
	MainActivity        string   `json:"main_activity"`
	SecondaryActivities []string `json:"secondary_activities"`
	LegalNature         string   `json:"legal_nature"`
	OpeningDate         string   `json:"opening_date"`
	RegistrationStatus  string   `json:"registration_status"`
}

type CNPJService struct {
	client *http.Client
}

func NewCNPJService() *CNPJService {
	return &CNPJService{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// limparCNPJ remove todos os caracteres não numéricos
func (s *CNPJService) limparCNPJ(cnpj string) string {
	return reDigits.ReplaceAllString(cnpj, "")
}

// validarCNPJ valida se o CNPJ é válido usando o algoritmo oficial
func (s *CNPJService) validarCNPJ(cnpj string) bool {
	cnpj = s.limparCNPJ(cnpj)
	if len(cnpj) != 14 {
		return false
	}

	// Bloqueia sequências iguais (11111111111111, etc.)
	allEqual := true
	for i := 1; i < 14; i++ {
		if cnpj[i] != cnpj[0] {
			allEqual = false
			break
		}
	}
	if allEqual {
		return false
	}

	// Calcula os dígitos verificadores
	calc := func(x int) int {
		n, pos := 0, x-7
		for i := 0; i < x; i++ {
			digit, _ := strconv.Atoi(string(cnpj[i]))
			n += digit * pos
			pos--
			if pos < 2 {
				pos = 9
			}
		}
		r := n % 11
		if r < 2 {
			return 0
		}
		return 11 - r
	}

	d1 := calc(12)
	d2 := calc(13)

	digit12, _ := strconv.Atoi(string(cnpj[12]))
	digit13, _ := strconv.Atoi(string(cnpj[13]))

	return d1 == digit12 && d2 == digit13
}

// ConsultarCNPJ consulta os dados do CNPJ na API do CNPJá
func (s *CNPJService) ConsultarCNPJ(ctx context.Context, cnpjRaw string) (*CNPJData, error) {
	cnpj := s.limparCNPJ(cnpjRaw)

	if !s.validarCNPJ(cnpj) {
		return nil, errors.New("CNPJ inválido")
	}

	url := fmt.Sprintf("https://open.cnpja.com/office/%s", cnpj)

	log.Info().
		Str("cnpj", cnpj).
		Str("url", url).
		Msg("Consultando CNPJ na API externa")

	backoff := 300 * time.Millisecond
	for tentativa := 0; tentativa < 5; tentativa++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar requisição: %w", err)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("erro na requisição HTTP: %w", err)
		}
		defer resp.Body.Close()

		// Se recebeu 429 (rate limit), tenta novamente com backoff
		if resp.StatusCode == http.StatusTooManyRequests && tentativa < 4 {
			log.Warn().
				Int("tentativa", tentativa+1).
				Dur("espera", backoff).
				Msg("Rate limit atingido, aguardando...")

			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			return nil, errors.New("CNPJ não encontrado")
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API retornou status %d", resp.StatusCode)
		}

		var rawData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
			return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
		}

		// Mapear os dados da API para nossa estrutura
		cnpjData := &CNPJData{}

		// CNPJ
		if taxId, ok := rawData["taxId"].(string); ok {
			cnpjData.CNPJ = taxId
		}

		// Nome da empresa (company.name)
		if company, ok := rawData["company"].(map[string]interface{}); ok {
			if name, ok := company["name"].(string); ok && name != "" {
				cnpjData.Name = name
			}
			// Natureza jurídica
			if nature, ok := company["nature"].(map[string]interface{}); ok {
				if text, ok := nature["text"].(string); ok && text != "" {
					cnpjData.LegalNature = text
				}
			}
			// Porte da empresa
			if size, ok := company["size"].(map[string]interface{}); ok {
				if text, ok := size["text"].(string); ok && text != "" {
					cnpjData.CompanySize = text
				}
			}
		}

		// Nome fantasia
		if alias, ok := rawData["alias"].(string); ok && alias != "" {
			cnpjData.TradeName = alias
		}

		// Endereço
		if address, ok := rawData["address"].(map[string]interface{}); ok {
			if street, ok := address["street"].(string); ok && street != "" {
				cnpjData.Address = street
			}
			if number, ok := address["number"].(string); ok && number != "" {
				cnpjData.Number = number
			}
			if details, ok := address["details"].(string); ok && details != "" {
				cnpjData.Complement = details
			}
			if district, ok := address["district"].(string); ok && district != "" {
				cnpjData.District = district
			}
			if city, ok := address["city"].(string); ok && city != "" {
				cnpjData.City = city
			}
			if state, ok := address["state"].(string); ok && state != "" {
				cnpjData.State = state
			}
			if zip, ok := address["zip"].(string); ok && zip != "" {
				cnpjData.ZipCode = zip
			}
		}

		// Telefones
		if phones, ok := rawData["phones"].([]interface{}); ok && len(phones) > 0 {
			if phone, ok := phones[0].(map[string]interface{}); ok {
				if area, ok := phone["area"].(string); ok {
					if number, ok := phone["number"].(string); ok {
						cnpjData.Phone = fmt.Sprintf("(%s) %s", area, number)
					}
				}
			}
		}

		// Emails
		if emails, ok := rawData["emails"].([]interface{}); ok && len(emails) > 0 {
			if email, ok := emails[0].(map[string]interface{}); ok {
				if address, ok := email["address"].(string); ok && address != "" {
					cnpjData.Email = address
				}
			}
		}

		// Atividade principal
		if mainActivity, ok := rawData["mainActivity"].(map[string]interface{}); ok {
			if text, ok := mainActivity["text"].(string); ok && text != "" {
				cnpjData.MainActivity = text
			}
		}

		// Atividades secundárias
		if sideActivities, ok := rawData["sideActivities"].([]interface{}); ok && len(sideActivities) > 0 {
			if activity, ok := sideActivities[0].(map[string]interface{}); ok {
				if text, ok := activity["text"].(string); ok && text != "" {
					cnpjData.SecondaryActivities = append(cnpjData.SecondaryActivities, text)
				}
			}
		}

		// Data de fundação
		if founded, ok := rawData["founded"].(string); ok && founded != "" {
			cnpjData.OpeningDate = founded
		}

		// Status
		if status, ok := rawData["status"].(map[string]interface{}); ok {
			if text, ok := status["text"].(string); ok && text != "" {
				cnpjData.RegistrationStatus = text
			}
		}

		log.Info().
			Str("cnpj", cnpj).
			Str("name", cnpjData.Name).
			Msg("CNPJ consultado com sucesso")

		return cnpjData, nil
	}

	return nil, errors.New("limite de tentativas excedido")
}

// FormatarCNPJ formata o CNPJ com máscara
func (s *CNPJService) FormatarCNPJ(cnpj string) string {
	cnpj = s.limparCNPJ(cnpj)
	if len(cnpj) != 14 {
		return cnpj
	}
	return fmt.Sprintf("%s.%s.%s/%s-%s",
		cnpj[0:2], cnpj[2:5], cnpj[5:8], cnpj[8:12], cnpj[12:14])
}
