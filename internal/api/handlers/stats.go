package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
)

// StatsHandler gerencia as rotas de estatísticas
type StatsHandler struct{}

// NewStatsHandler cria uma nova instância do handler de estatísticas
func NewStatsHandler() *StatsHandler {
	return &StatsHandler{}
}

// DashboardStatsResponse representa a resposta das estatísticas do dashboard
type DashboardStatsResponse struct {
	Companies struct {
		Total      int `json:"total"`
		Active     int `json:"active"`
		Restricted int `json:"restricted"`
		AutoFetch  int `json:"auto_fetch"`
		ThisWeek   int `json:"this_week"`
	} `json:"companies"`
	Documents struct {
		Total     int `json:"total"`
		Processed int `json:"processed"`
		Pending   int `json:"pending"`
		Errors    int `json:"errors"`
		Today     int `json:"today"`
	} `json:"documents"`
	Users struct {
		Total  int `json:"total"`
		Active int `json:"active"`
		Admins int `json:"admins"`
	} `json:"users"`
	RecentActivity struct {
		DocumentsToday     int `json:"documents_today"`
		CompaniesThisWeek  int `json:"companies_this_week"`
		LastSyncTime       *time.Time `json:"last_sync_time,omitempty"`
	} `json:"recent_activity"`
}

// GetDashboardStats retorna estatísticas para o dashboard
// @Summary Estatísticas do dashboard
// @Description Retorna estatísticas gerais do sistema para o dashboard
// @Tags stats
// @Produce json
// @Success 200 {object} DashboardStatsResponse "Estatísticas do dashboard"
// @Failure 401 {object} SwaggerError "Token inválido"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Security BearerAuth
// @Router /stats/dashboard [get]
func (h *StatsHandler) GetDashboardStats(c *fiber.Ctx) error {
	var stats DashboardStatsResponse

	// Estatísticas de empresas
	var companies []models.Company
	err := database.DB.NewSelect().Model(&companies).Scan(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch companies",
		})
	}

	stats.Companies.Total = len(companies)
	weekAgo := time.Now().AddDate(0, 0, -7)
	
	for _, company := range companies {
		if company.Active {
			stats.Companies.Active++
		}
		if company.Restricted {
			stats.Companies.Restricted++
		}
		if company.AutoFetch {
			stats.Companies.AutoFetch++
		}
		if company.CreatedAt.After(weekAgo) {
			stats.Companies.ThisWeek++
		}
	}

	// Estatísticas de documentos
	var documents []models.Document
	err = database.DB.NewSelect().Model(&documents).Scan(c.Context())
	if err != nil {
		// Se não conseguir buscar documentos, continuar com zeros
		stats.Documents.Total = 0
		stats.Documents.Processed = 0
		stats.Documents.Pending = 0
		stats.Documents.Errors = 0
		stats.Documents.Today = 0
	} else {
		stats.Documents.Total = len(documents)
		today := time.Now().Truncate(24 * time.Hour)
		
		for _, doc := range documents {
			switch doc.Status {
			case "processed":
				stats.Documents.Processed++
			case "pending":
				stats.Documents.Pending++
			case "error":
				stats.Documents.Errors++
			}
			
			if doc.CreatedAt.After(today) {
				stats.Documents.Today++
			}
		}
	}

	// Estatísticas de usuários (apenas para admins)
	user, ok := c.Locals("user").(*models.User)
	if ok && user.IsAdmin() {
		var users []models.User
		err = database.DB.NewSelect().Model(&users).Scan(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch users",
			})
		}

		stats.Users.Total = len(users)
		for _, u := range users {
			if u.Active {
				stats.Users.Active++
			}
			if u.IsAdmin() {
				stats.Users.Admins++
			}
		}
	}

	// Atividade recente
	stats.RecentActivity.DocumentsToday = stats.Documents.Today
	stats.RecentActivity.CompaniesThisWeek = stats.Companies.ThisWeek

	// Buscar último tempo de sincronização (se houver documentos)
	if stats.Documents.Total > 0 {
		var lastDoc models.Document
		err = database.DB.NewSelect().
			Model(&lastDoc).
			Order("created_at DESC").
			Limit(1).
			Scan(c.Context())
		if err == nil {
			stats.RecentActivity.LastSyncTime = &lastDoc.CreatedAt
		}
	}

	return c.JSON(stats)
}

// GetCompanyStats retorna estatísticas de uma empresa específica
// @Summary Estatísticas de empresa
// @Description Retorna estatísticas detalhadas de uma empresa específica
// @Tags stats
// @Produce json
// @Param id path int true "ID da empresa"
// @Success 200 {object} map[string]interface{} "Estatísticas da empresa"
// @Failure 401 {object} SwaggerError "Token inválido"
// @Failure 404 {object} SwaggerError "Empresa não encontrada"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Security BearerAuth
// @Router /stats/companies/{id} [get]
func (h *StatsHandler) GetCompanyStats(c *fiber.Ctx) error {
	companyID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	// Verificar se a empresa existe
	var company models.Company
	err = database.DB.NewSelect().
		Model(&company).
		Where("id = ?", companyID).
		Scan(c.Context())
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Company not found",
		})
	}

	// Buscar documentos da empresa
	var documents []models.Document
	err = database.DB.NewSelect().
		Model(&documents).
		Where("company_id = ?", companyID).
		Scan(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch company documents",
		})
	}

	// Calcular estatísticas
	stats := map[string]interface{}{
		"company": company,
		"documents": map[string]interface{}{
			"total":     len(documents),
			"processed": 0,
			"pending":   0,
			"errors":    0,
			"this_month": 0,
		},
	}

	thisMonth := time.Now().AddDate(0, -1, 0)
	docStats := stats["documents"].(map[string]interface{})
	
	for _, doc := range documents {
		switch doc.Status {
		case "processed":
			docStats["processed"] = docStats["processed"].(int) + 1
		case "pending":
			docStats["pending"] = docStats["pending"].(int) + 1
		case "error":
			docStats["errors"] = docStats["errors"].(int) + 1
		}
		
		if doc.CreatedAt.After(thisMonth) {
			docStats["this_month"] = docStats["this_month"].(int) + 1
		}
	}

	return c.JSON(stats)
}
