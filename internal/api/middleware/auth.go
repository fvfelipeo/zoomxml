package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/models"
	"github.com/zoomxml/internal/services"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(models.APIResponse{
				Success: false,
				Error:   "Authorization header required",
			})
		}

		// Check Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(models.APIResponse{
				Success: false,
				Error:   "Invalid authorization format",
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return c.Status(401).JSON(models.APIResponse{
				Success: false,
				Error:   "Token required",
			})
		}

		// Validate token
		empresa, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(401).JSON(models.APIResponse{
				Success: false,
				Error:   "Invalid or expired token",
			})
		}

		// Store empresa in context
		c.Locals("empresa", empresa)
		c.Locals("empresa_id", empresa.ID)
		c.Locals("empresa_uuid", empresa.UUID)

		return c.Next()
	}
}

// GetEmpresaFromContext gets the empresa from fiber context
func GetEmpresaFromContext(c *fiber.Ctx) *models.Empresa {
	empresa, ok := c.Locals("empresa").(*models.Empresa)
	if !ok {
		return nil
	}
	return empresa
}

// GetEmpresaIDFromContext gets the empresa ID from fiber context
func GetEmpresaIDFromContext(c *fiber.Ctx) int {
	empresaID, ok := c.Locals("empresa_id").(int)
	if !ok {
		return 0
	}
	return empresaID
}

// GetEmpresaUUIDFromContext gets the empresa UUID from fiber context
func GetEmpresaUUIDFromContext(c *fiber.Ctx) string {
	empresaUUID, ok := c.Locals("empresa_uuid").(string)
	if !ok {
		return ""
	}
	return empresaUUID
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Simple logging - in production use structured logging
		return c.Next()
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(200)
		}

		return c.Next()
	}
}
