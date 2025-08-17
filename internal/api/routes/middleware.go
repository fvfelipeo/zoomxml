package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/zoomxml/internal/models"
)

// RouteMiddleware provides route-specific middleware configurations
type RouteMiddleware struct{}

// NewRouteMiddleware creates a new route middleware instance
func NewRouteMiddleware() *RouteMiddleware {
	return &RouteMiddleware{}
}

// RateLimitConfig defines rate limiting configuration for different route groups
type RateLimitConfig struct {
	// Public routes (auth, health)
	PublicLimit int
	
	// Authenticated routes (general API)
	AuthenticatedLimit int
	
	// Heavy operations (sync, bulk operations)
	HeavyOperationsLimit int
	
	// XML download routes
	DownloadLimit int
}

// GetDefaultRateLimitConfig returns default rate limiting configuration
func GetDefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		PublicLimit:          100,  // 100 requests per minute
		AuthenticatedLimit:   1000, // 1000 requests per minute
		HeavyOperationsLimit: 10,   // 10 requests per minute
		DownloadLimit:        50,   // 50 downloads per minute
	}
}

// ApplyPublicRateLimit applies rate limiting for public routes
func (rm *RouteMiddleware) ApplyPublicRateLimit() fiber.Handler {
	config := GetDefaultRateLimitConfig()
	
	return limiter.New(limiter.Config{
		Max:        config.PublicLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(models.APIResponse{
				Success: false,
				Error:   "Rate limit exceeded for public routes",
			})
		},
	})
}

// ApplyAuthenticatedRateLimit applies rate limiting for authenticated routes
func (rm *RouteMiddleware) ApplyAuthenticatedRateLimit() fiber.Handler {
	config := GetDefaultRateLimitConfig()
	
	return limiter.New(limiter.Config{
		Max:        config.AuthenticatedLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use user ID from context if available, fallback to IP
			userID := c.Locals("empresa_id")
			if userID != nil {
				return userID.(string)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(models.APIResponse{
				Success: false,
				Error:   "Rate limit exceeded for authenticated routes",
			})
		},
	})
}

// ApplyHeavyOperationsRateLimit applies rate limiting for heavy operations
func (rm *RouteMiddleware) ApplyHeavyOperationsRateLimit() fiber.Handler {
	config := GetDefaultRateLimitConfig()
	
	return limiter.New(limiter.Config{
		Max:        config.HeavyOperationsLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("empresa_id")
			if userID != nil {
				return "heavy_" + userID.(string)
			}
			return "heavy_" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(models.APIResponse{
				Success: false,
				Error:   "Rate limit exceeded for heavy operations. Please wait before retrying.",
			})
		},
	})
}

// ApplyDownloadRateLimit applies rate limiting for download operations
func (rm *RouteMiddleware) ApplyDownloadRateLimit() fiber.Handler {
	config := GetDefaultRateLimitConfig()
	
	return limiter.New(limiter.Config{
		Max:        config.DownloadLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("empresa_id")
			if userID != nil {
				return "download_" + userID.(string)
			}
			return "download_" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(models.APIResponse{
				Success: false,
				Error:   "Download rate limit exceeded. Please wait before downloading more files.",
			})
		},
	})
}

// ApplyTimeoutMiddleware applies timeout for long-running operations
func (rm *RouteMiddleware) ApplyTimeoutMiddleware(duration time.Duration) fiber.Handler {
	return timeout.New(func(c *fiber.Ctx) error {
		return c.Next()
	}, duration)
}

// ApplyValidationMiddleware applies request validation
func (rm *RouteMiddleware) ApplyValidationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add request validation logic here
		// For example, validate content-type, request size, etc.
		
		// Check content length for POST/PUT requests
		if c.Method() == "POST" || c.Method() == "PUT" {
			contentLength := len(c.Body())
			maxSize := 10 * 1024 * 1024 // 10MB
			
			if contentLength > maxSize {
				return c.Status(413).JSON(models.APIResponse{
					Success: false,
					Error:   "Request body too large",
				})
			}
		}
		
		return c.Next()
	}
}

// ApplySecurityHeaders applies security headers
func (rm *RouteMiddleware) ApplySecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// API specific headers
		c.Set("X-API-Version", "1.0.0")
		c.Set("X-Service", "ZoomXML")
		
		return c.Next()
	}
}

// ApplyCompressionMiddleware applies response compression
func (rm *RouteMiddleware) ApplyCompressionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Enable compression for JSON responses
		if c.Get("Accept-Encoding") != "" {
			c.Set("Content-Encoding", "gzip")
		}
		
		return c.Next()
	}
}
