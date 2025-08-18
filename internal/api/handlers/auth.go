package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler gerencia as rotas de autenticação
type AuthHandler struct{}

// NewAuthHandler cria uma nova instância do handler de autenticação
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse representa a resposta de login
type LoginResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Active    bool   `json:"active"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// checkPassword verifica se a senha fornecida corresponde ao hash
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Login autentica um usuário com email e senha
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
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

	// Buscar usuário por email
	user := &models.User{}
	err := database.DB.NewSelect().
		Model(user).
		Where("email = ? AND active = true", req.Email).
		Scan(c.Context())

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Verificar senha
	if !checkPassword(req.Password, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Retornar dados do usuário com token
	response := LoginResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Active:    user.Active,
		Token:     user.Token,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.JSON(response)
}

// Logout invalida o token do usuário (opcional - regenera o token)
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Obter usuário do contexto (definido pelo middleware de auth)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Gerar novo token para invalidar o atual
	newToken := generateToken()

	// Atualizar token no banco
	_, err := database.DB.NewUpdate().
		Model(user).
		Set("token = ?", newToken).
		Where("id = ?", user.ID).
		Exec(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to logout",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Logout successful",
	})
}

// GetProfile retorna o perfil do usuário autenticado
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	// Obter usuário do contexto (definido pelo middleware de auth)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	return c.JSON(user)
}
