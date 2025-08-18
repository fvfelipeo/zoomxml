package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/api/middleware"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
)

// CompanyHandler gerencia as rotas de empresas
type CompanyHandler struct{}

// NewCompanyHandler cria uma nova instância do handler de empresas
func NewCompanyHandler() *CompanyHandler {
	return &CompanyHandler{}
}

// CreateCompanyRequest representa a requisição para criar empresa
type CreateCompanyRequest struct {
	// Dados básicos obrigatórios
	Name string `json:"name" validate:"required,min=2,max=255"`
	CNPJ string `json:"cnpj" validate:"required,min=14,max=18"`

	// Nome fantasia
	TradeName string `json:"trade_name,omitempty"`

	// Endereço completo
	Address    string `json:"address,omitempty"`
	Number     string `json:"number,omitempty"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	ZipCode    string `json:"zip_code,omitempty"`

	// Contato
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`

	// Dados empresariais
	CompanySize        string `json:"company_size,omitempty"`        // ME, EPP, etc
	MainActivity       string `json:"main_activity,omitempty"`       // Atividade principal
	SecondaryActivity  string `json:"secondary_activity,omitempty"`  // Atividades secundárias
	LegalNature        string `json:"legal_nature,omitempty"`        // Natureza jurídica
	OpeningDate        string `json:"opening_date,omitempty"`        // Data de abertura
	RegistrationStatus string `json:"registration_status,omitempty"` // Situação cadastral

	// Configurações do sistema
	Restricted bool `json:"restricted"`
	AutoFetch  bool `json:"auto_fetch"`
}

// UpdateCompanyRequest representa a requisição para atualizar empresa
type UpdateCompanyRequest struct {
	// Dados básicos
	Name      *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	CNPJ      *string `json:"cnpj,omitempty" validate:"omitempty,min=14,max=18"`
	TradeName *string `json:"trade_name,omitempty"`

	// Endereço
	Address    *string `json:"address,omitempty"`
	Number     *string `json:"number,omitempty"`
	Complement *string `json:"complement,omitempty"`
	District   *string `json:"district,omitempty"`
	City       *string `json:"city,omitempty"`
	State      *string `json:"state,omitempty"`
	ZipCode    *string `json:"zip_code,omitempty"`

	// Contato
	Phone *string `json:"phone,omitempty"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`

	// Dados empresariais
	CompanySize        *string `json:"company_size,omitempty"`
	MainActivity       *string `json:"main_activity,omitempty"`
	SecondaryActivity  *string `json:"secondary_activity,omitempty"`
	LegalNature        *string `json:"legal_nature,omitempty"`
	OpeningDate        *string `json:"opening_date,omitempty"`
	RegistrationStatus *string `json:"registration_status,omitempty"`

	// Configurações
	Restricted *bool `json:"restricted,omitempty"`
	AutoFetch  *bool `json:"auto_fetch,omitempty"`
	Active     *bool `json:"active,omitempty"`
}

// CreateCompany cria uma nova empresa
// @Summary Criar empresa
// @Description Cria uma nova empresa no sistema (requer autenticação)
// @Tags companies
// @Accept json
// @Produce json
// @Param company body CreateCompanyRequest true "Dados da empresa"
// @Success 201 {object} SwaggerCompany
// @Failure 400 {object} SwaggerValidationError "Erro de validação"
// @Failure 401 {object} SwaggerError "Autenticação necessária"
// @Failure 409 {object} SwaggerError "CNPJ já existe"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Security UserToken
// @Router /companies [post]
func (h *CompanyHandler) CreateCompany(c *fiber.Ctx) error {
	var req CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validar entrada
	if err := validateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err,
		})
	}

	// Verificar se CNPJ já existe
	exists, err := database.DB.NewSelect().
		Model((*models.Company)(nil)).
		Where("cnpj = ?", req.CNPJ).
		Exists(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "CNPJ already exists",
		})
	}

	// Criar empresa
	company := &models.Company{
		Name:      req.Name,
		CNPJ:      req.CNPJ,
		TradeName: req.TradeName,

		// Endereço
		Address:    req.Address,
		Number:     req.Number,
		Complement: req.Complement,
		District:   req.District,
		City:       req.City,
		State:      req.State,
		ZipCode:    req.ZipCode,

		// Contato
		Phone: req.Phone,
		Email: req.Email,

		// Dados empresariais
		CompanySize:        req.CompanySize,
		MainActivity:       req.MainActivity,
		SecondaryActivity:  req.SecondaryActivity,
		LegalNature:        req.LegalNature,
		OpeningDate:        req.OpeningDate,
		RegistrationStatus: req.RegistrationStatus,

		// Configurações
		Restricted: req.Restricted,
		AutoFetch:  req.AutoFetch,
		Active:     true,
	}

	_, err = database.DB.NewInsert().Model(company).Exec(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create company",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(company)
}

// GetCompanies lista empresas com base nas regras de visibilidade
// @Summary Listar empresas
// @Description Lista empresas conforme regras de visibilidade (públicas para todos, restritas apenas para membros/admin)
// @Tags companies
// @Produce json
// @Param active query string false "Filtrar por status (true/false) - apenas admin"
// @Param restricted query string false "Filtrar por tipo (true/false) - apenas admin"
// @Param page query int false "Página (padrão: 1)"
// @Param limit query int false "Itens por página (padrão: 20)"
// @Success 200 {object} SwaggerCompaniesResponse "Lista de empresas com paginação"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Router /companies [get]
func (h *CompanyHandler) GetCompanies(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)

	var companies []models.Company
	query := database.DB.NewSelect().Model(&companies)

	// Aplicar regras de visibilidade
	if user == nil {
		// Usuário não autenticado - apenas empresas não restritas
		query = query.Where("restricted = false AND active = true")
	} else if !user.IsAdmin() {
		// Usuário comum - empresas não restritas + empresas onde é membro
		query = query.Where(`
			(restricted = false AND active = true) OR
			(id IN (
				SELECT cm.company_id FROM company_members cm
				JOIN companies c2 ON cm.company_id = c2.id
				WHERE cm.user_id = ? AND c2.active = true
			))
		`, user.ID)
	}
	// Admin vê todas as empresas (sem filtro adicional)

	// Filtros opcionais
	if active := c.Query("active"); active != "" && user != nil && user.IsAdmin() {
		switch active {
		case "true":
			query = query.Where("active = true")
		case "false":
			query = query.Where("active = false")
		}
	}

	if restricted := c.Query("restricted"); restricted != "" && user != nil && user.IsAdmin() {
		switch restricted {
		case "true":
			query = query.Where("restricted = true")
		case "false":
			query = query.Where("restricted = false")
		}
	}

	// Paginação
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	query = query.Limit(limit).Offset(offset).Order("id ASC")

	err := query.Scan(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch companies",
		})
	}

	// Contar total (aplicando os mesmos filtros)
	countQuery := database.DB.NewSelect().Model((*models.Company)(nil))
	if user == nil {
		countQuery = countQuery.Where("restricted = false AND active = true")
	} else if !user.IsAdmin() {
		countQuery = countQuery.Where(`
			(restricted = false AND active = true) OR 
			(id IN (
				SELECT company_id FROM company_members 
				WHERE user_id = ?
			))
		`, user.ID)
	}

	total, err := countQuery.Count(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count companies",
		})
	}

	return c.JSON(fiber.Map{
		"companies": companies,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetCompany obtém uma empresa específica
func (h *CompanyHandler) GetCompany(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	user := middleware.GetUserFromContext(c)

	company := &models.Company{}
	query := database.DB.NewSelect().Model(company).Where("id = ?", id)

	// Aplicar regras de visibilidade
	if user == nil {
		query = query.Where("restricted = false AND active = true")
	} else if !user.IsAdmin() {
		query = query.Where(`
			(restricted = false AND active = true) OR
			(id IN (
				SELECT cm.company_id FROM company_members cm
				JOIN companies c2 ON cm.company_id = c2.id
				WHERE cm.user_id = ? AND c2.active = true
			))
		`, user.ID)
	}

	err = query.Scan(c.Context())
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Company not found or access denied",
		})
	}

	return c.JSON(company)
}

// UpdateCompany atualiza uma empresa
func (h *CompanyHandler) UpdateCompany(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	var req UpdateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validar entrada
	if err := validateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err,
		})
	}

	user := middleware.GetUserFromContext(c)

	// Verificar acesso à empresa
	company := &models.Company{}
	accessQuery := database.DB.NewSelect().Model(company).Where("id = ?", id)

	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	if !user.IsAdmin() {
		accessQuery = accessQuery.Where(`
			(restricted = false) OR 
			(id IN (
				SELECT company_id FROM company_members 
				WHERE user_id = ?
			))
		`, user.ID)
	}

	err = accessQuery.Scan(c.Context())
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Company not found or access denied",
		})
	}

	// Atualizar campos
	query := database.DB.NewUpdate().Model(company).Where("id = ?", id)

	if req.Name != nil {
		query = query.Set("name = ?", *req.Name)
		company.Name = *req.Name
	}

	if req.CNPJ != nil {
		// Verificar se CNPJ já existe (exceto para a própria empresa)
		exists, err := database.DB.NewSelect().
			Model((*models.Company)(nil)).
			Where("cnpj = ? AND id != ?", *req.CNPJ, id).
			Exists(c.Context())

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Database error",
			})
		}

		if exists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "CNPJ already exists",
			})
		}

		query = query.Set("cnpj = ?", *req.CNPJ)
		company.CNPJ = *req.CNPJ
	}

	if req.Address != nil {
		query = query.Set("address = ?", *req.Address)
		company.Address = *req.Address
	}

	if req.City != nil {
		query = query.Set("city = ?", *req.City)
		company.City = *req.City
	}

	if req.State != nil {
		query = query.Set("state = ?", *req.State)
		company.State = *req.State
	}

	if req.ZipCode != nil {
		query = query.Set("zip_code = ?", *req.ZipCode)
		company.ZipCode = *req.ZipCode
	}

	if req.TradeName != nil {
		query = query.Set("trade_name = ?", *req.TradeName)
		company.TradeName = *req.TradeName
	}

	if req.Number != nil {
		query = query.Set("number = ?", *req.Number)
		company.Number = *req.Number
	}

	if req.Complement != nil {
		query = query.Set("complement = ?", *req.Complement)
		company.Complement = *req.Complement
	}

	if req.District != nil {
		query = query.Set("district = ?", *req.District)
		company.District = *req.District
	}

	if req.Phone != nil {
		query = query.Set("phone = ?", *req.Phone)
		company.Phone = *req.Phone
	}

	if req.Email != nil {
		query = query.Set("email = ?", *req.Email)
		company.Email = *req.Email
	}

	// Dados empresariais
	if req.CompanySize != nil {
		query = query.Set("company_size = ?", *req.CompanySize)
		company.CompanySize = *req.CompanySize
	}

	if req.MainActivity != nil {
		query = query.Set("main_activity = ?", *req.MainActivity)
		company.MainActivity = *req.MainActivity
	}

	if req.SecondaryActivity != nil {
		query = query.Set("secondary_activity = ?", *req.SecondaryActivity)
		company.SecondaryActivity = *req.SecondaryActivity
	}

	if req.LegalNature != nil {
		query = query.Set("legal_nature = ?", *req.LegalNature)
		company.LegalNature = *req.LegalNature
	}

	if req.OpeningDate != nil {
		query = query.Set("opening_date = ?", *req.OpeningDate)
		company.OpeningDate = *req.OpeningDate
	}

	if req.RegistrationStatus != nil {
		query = query.Set("registration_status = ?", *req.RegistrationStatus)
		company.RegistrationStatus = *req.RegistrationStatus
	}

	// Apenas admin pode alterar restricted e active
	if user.IsAdmin() {
		if req.Restricted != nil {
			query = query.Set("restricted = ?", *req.Restricted)
			company.Restricted = *req.Restricted
		}

		if req.Active != nil {
			query = query.Set("active = ?", *req.Active)
			company.Active = *req.Active
		}
	}

	if req.AutoFetch != nil {
		query = query.Set("auto_fetch = ?", *req.AutoFetch)
		company.AutoFetch = *req.AutoFetch
	}

	_, err = query.Exec(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update company",
		})
	}

	return c.JSON(company)
}

// DeleteCompany remove uma empresa (apenas admin)
func (h *CompanyHandler) DeleteCompany(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	// Verificar se empresa existe
	exists, err := database.DB.NewSelect().
		Model((*models.Company)(nil)).
		Where("id = ?", id).
		Exists(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Company not found",
		})
	}

	_, err = database.DB.NewDelete().
		Model((*models.Company)(nil)).
		Where("id = ?", id).
		Exec(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete company",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
