package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler gerencia as rotas de usuários
type UserHandler struct{}

// NewUserHandler cria uma nova instância do handler de usuários
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// generateToken gera um token aleatório
func generateToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// hashPassword gera um hash bcrypt da senha
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CreateUserRequest representa a requisição para criar usuário
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"` // Senha obrigatória para frontend
	Token    string `json:"token,omitempty"`                    // Token opcional - se não fornecido, será gerado automaticamente
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

// UpdateUserRequest representa a requisição para atualizar usuário
type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8"` // Nova senha opcional
	Token    *string `json:"token,omitempty"`                               // Novo token opcional
	Role     *string `json:"role,omitempty" validate:"omitempty,oneof=admin user"`
	Active   *bool   `json:"active,omitempty"`
}

// CreateUser cria um novo usuário (apenas admin)
// @Summary Criar usuário
// @Description Cria um novo usuário no sistema (apenas admin com token especial)
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "Dados do usuário"
// @Success 201 {object} SwaggerUserWithToken
// @Failure 400 {object} SwaggerValidationError "Erro de validação"
// @Failure 401 {object} SwaggerError "Token de admin necessário"
// @Failure 409 {object} SwaggerError "Email já existe"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Security AdminToken
// @Router /users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
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

	// Verificar se email já existe
	exists, err := database.DB.NewSelect().
		Model((*models.User)(nil)).
		Where("email = ?", req.Email).
		Exists(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	// Hash da senha
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	// Usar token fornecido ou gerar um novo
	userToken := req.Token
	if userToken == "" {
		userToken = generateToken()
	}

	// Criar usuário
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Token:    userToken,
		Role:     req.Role,
		Active:   true,
	}

	_, err = database.DB.NewInsert().Model(user).Exec(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Retornar usuário criado com token
	response := fiber.Map{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"active":     user.Active,
		"token":      user.Token, // Incluir token na resposta de criação
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetUsers lista todos os usuários (apenas admin)
// @Summary Listar usuários
// @Description Lista todos os usuários do sistema com paginação e filtros
// @Tags users
// @Produce json
// @Param role query string false "Filtrar por papel (admin/user)"
// @Param active query string false "Filtrar por status (true/false)"
// @Param page query int false "Página (padrão: 1)"
// @Param limit query int false "Itens por página (padrão: 20)"
// @Success 200 {object} SwaggerUsersResponse "Lista de usuários com paginação"
// @Failure 401 {object} SwaggerError "Token de admin necessário"
// @Failure 500 {object} SwaggerError "Erro interno"
// @Security AdminToken
// @Router /users [get]
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	var users []models.User

	query := database.DB.NewSelect().Model(&users)

	// Filtros opcionais
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	if active := c.Query("active"); active != "" {
		switch active {
		case "true":
			query = query.Where("active = true")
		case "false":
			query = query.Where("active = false")
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
			"error": "Failed to fetch users",
		})
	}

	// Contar total
	total, err := database.DB.NewSelect().Model((*models.User)(nil)).Count(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count users",
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetUser obtém um usuário específico (apenas admin)
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	user := &models.User{}
	err = database.DB.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(c.Context())

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

// UpdateUser atualiza um usuário (apenas admin)
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req UpdateUserRequest
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

	// Buscar usuário existente
	user := &models.User{}
	err = database.DB.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(c.Context())

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Atualizar campos
	query := database.DB.NewUpdate().Model(user).Where("id = ?", id)

	if req.Name != nil {
		query = query.Set("name = ?", *req.Name)
		user.Name = *req.Name
	}

	if req.Email != nil {
		// Verificar se email já existe (exceto para o próprio usuário)
		exists, err := database.DB.NewSelect().
			Model((*models.User)(nil)).
			Where("email = ? AND id != ?", *req.Email, id).
			Exists(c.Context())

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Database error",
			})
		}

		if exists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Email already exists",
			})
		}

		query = query.Set("email = ?", *req.Email)
		user.Email = *req.Email
	}

	if req.Password != nil {
		hashedPassword, err := hashPassword(*req.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to hash password",
			})
		}
		query = query.Set("password = ?", hashedPassword)
		user.Password = hashedPassword
	}

	if req.Token != nil {
		query = query.Set("token = ?", *req.Token)
		user.Token = *req.Token
	}

	if req.Role != nil {
		query = query.Set("role = ?", *req.Role)
		user.Role = *req.Role
	}

	if req.Active != nil {
		query = query.Set("active = ?", *req.Active)
		user.Active = *req.Active
	}

	_, err = query.Exec(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	return c.JSON(user)
}

// DeleteUser remove um usuário (apenas admin)
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Verificar se usuário existe
	exists, err := database.DB.NewSelect().
		Model((*models.User)(nil)).
		Where("id = ?", id).
		Exists(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	_, err = database.DB.NewDelete().
		Model((*models.User)(nil)).
		Where("id = ?", id).
		Exec(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
