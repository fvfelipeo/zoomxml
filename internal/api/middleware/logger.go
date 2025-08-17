package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/logger"
)

// LoggerMiddleware creates a custom logging middleware using our zerolog logger
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get user ID if available
		var userID *int64
		if user := GetUserFromContext(c); user != nil {
			userID = &user.ID
		}

		// Log the request
		logger.LogAPIRequest(
			c.Context(),
			c.Method(),
			c.Path(),
			userID,
			c.Response().StatusCode(),
			duration,
		)

		return err
	}
}

// LoggerConfig holds configuration for the logger middleware
type LoggerConfig struct {
	// Skip defines a function to skip middleware
	Skip func(c *fiber.Ctx) bool

	// TimeFormat defines the time format for timestamps
	TimeFormat string

	// TimeZone defines the timezone for timestamps
	TimeZone string

	// Custom logger function
	CustomLogger func(c *fiber.Ctx, duration time.Duration)
}

// LoggerWithConfig creates a custom logging middleware with configuration
func LoggerWithConfig(config LoggerConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip logging if configured
		if config.Skip != nil && config.Skip(c) {
			return c.Next()
		}

		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Use custom logger if provided
		if config.CustomLogger != nil {
			config.CustomLogger(c, duration)
		} else {
			// Use default logger
			var userID *int64
			if user := GetUserFromContext(c); user != nil {
				userID = &user.ID
			}

			logger.LogAPIRequest(
				c.Context(),
				c.Method(),
				c.Path(),
				userID,
				c.Response().StatusCode(),
				duration,
			)
		}

		return err
	}
}

// HealthCheckSkipper skips logging for health check endpoints
func HealthCheckSkipper(c *fiber.Ctx) bool {
	return c.Path() == "/health" || c.Path() == "/metrics"
}

// StaticFileSkipper skips logging for static files
func StaticFileSkipper(c *fiber.Ctx) bool {
	path := c.Path()
	return len(path) > 4 && (
		path[len(path)-4:] == ".css" ||
		path[len(path)-3:] == ".js" ||
		path[len(path)-4:] == ".png" ||
		path[len(path)-4:] == ".jpg" ||
		path[len(path)-5:] == ".jpeg" ||
		path[len(path)-4:] == ".gif" ||
		path[len(path)-4:] == ".ico" ||
		path[len(path)-4:] == ".svg")
}

// CombinedSkipper combines multiple skip functions
func CombinedSkipper(skippers ...func(c *fiber.Ctx) bool) func(c *fiber.Ctx) bool {
	return func(c *fiber.Ctx) bool {
		for _, skipper := range skippers {
			if skipper(c) {
				return true
			}
		}
		return false
	}
}
