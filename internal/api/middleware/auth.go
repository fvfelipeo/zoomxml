package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/config"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/models"
)

// UserContextKey é a chave para armazenar o usuário no contexto
type UserContextKey string

const UserKey UserContextKey = "user"

// AuthMiddleware middleware para autenticação com token simples
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extrair token do header "token" ou "Authorization"
		tokenString := c.Get("token")
		if tokenString == "" {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token header required",
				})
			}

			// Verificar formato "Bearer <token>" ou apenas "<token>"
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				tokenString = authHeader
			}
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token required",
			})
		}

		// Buscar usuário pelo token no banco de dados
		user := &models.User{}
		err := database.DB.NewSelect().
			Model(user).
			Where("token = ? AND active = true", tokenString).
			Scan(c.Context())

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token or user not found",
			})
		}

		// Adicionar usuário ao contexto
		c.Locals(string(UserKey), user)

		return c.Next()
	}
}

// AdminTokenMiddleware middleware para validação do token de admin
func AdminTokenMiddleware() fiber.Handler {
	cfg := config.Get()

	return func(c *fiber.Ctx) error {
		// Extrair token do header "token" ou "Authorization"
		tokenString := c.Get("token")
		if tokenString == "" {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token header required",
				})
			}

			// Verificar formato "Bearer <token>" ou apenas "<token>"
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				tokenString = authHeader
			}
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token required",
			})
		}

		// Verificar se é o token de admin
		if tokenString != cfg.Auth.AdminToken {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid admin token",
			})
		}

		return c.Next()
	}
}

// AdminOnlyMiddleware middleware que permite apenas usuários admin
func AdminOnlyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUserFromContext(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		if !user.IsAdmin() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}

		return c.Next()
	}
}

// OptionalAuthMiddleware middleware de autenticação opcional
func OptionalAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extrair token do header "token" ou "Authorization"
		tokenString := c.Get("token")
		if tokenString == "" {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				// Sem token, continuar sem usuário
				return c.Next()
			}

			// Verificar formato "Bearer <token>" ou apenas "<token>"
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				tokenString = authHeader
			}
		}

		if tokenString == "" {
			// Token vazio, continuar sem usuário
			return c.Next()
		}

		// Buscar usuário pelo token no banco de dados
		user := &models.User{}
		err := database.DB.NewSelect().
			Model(user).
			Where("token = ? AND active = true", tokenString).
			Scan(c.Context())

		if err != nil {
			// Token inválido ou usuário não encontrado, continuar sem usuário
			return c.Next()
		}

		// Adicionar usuário ao contexto
		c.Locals(string(UserKey), user)

		return c.Next()
	}
}

// GetUserFromContext extrai o usuário do contexto
func GetUserFromContext(c *fiber.Ctx) *models.User {
	user, ok := c.Locals(string(UserKey)).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// GetUserFromGoContext extrai o usuário do contexto Go
func GetUserFromGoContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
