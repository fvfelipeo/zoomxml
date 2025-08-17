package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/models"
	"github.com/zoomxml/internal/services"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles user login
// @Summary Login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.APIResponse{data=models.LoginResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if req.CNPJ == "" || req.Password == "" {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "CNPJ and password are required",
		})
	}

	// Authenticate
	response, err := h.authService.Login(req.CNPJ, req.Password)
	if err != nil {
		return c.Status(401).JSON(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

// Logout handles user logout
// @Summary Logout
// @Description Invalidate current JWT token
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get token from header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(models.APIResponse{
			Success: false,
			Error:   "Authorization header required",
		})
	}

	token := authHeader[7:] // Remove "Bearer "

	err := h.authService.Logout(token)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "Failed to logout",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Logout successful",
	})
}

// RefreshToken handles token refresh
// @Summary Refresh Token
// @Description Refresh JWT token
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.LoginResponse}
// @Failure 401 {object} models.APIResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Get token from header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(models.APIResponse{
			Success: false,
			Error:   "Authorization header required",
		})
	}

	token := authHeader[7:] // Remove "Bearer "

	response, err := h.authService.RefreshToken(token)
	if err != nil {
		return c.Status(401).JSON(models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Data:    response,
	})
}

// Me returns current user info
// @Summary Get Current User
// @Description Get current authenticated user information
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.Empresa}
// @Failure 401 {object} models.APIResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	empresa := c.Locals("empresa").(*models.Empresa)

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    empresa,
	})
}
