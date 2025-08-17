package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/api/handlers"
	"github.com/zoomxml/internal/api/middleware"
	"github.com/zoomxml/internal/models"
	"github.com/zoomxml/internal/services"
)

// RouteConfig holds configuration for route setup
type RouteConfig struct {
	Handlers    *handlers.HandlerContainer
	AuthService *services.AuthService
}

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, config RouteConfig) {
	// Initialize route middleware
	routeMiddleware := NewRouteMiddleware()

	// Apply global middleware
	app.Use(routeMiddleware.ApplySecurityHeaders())
	app.Use(routeMiddleware.ApplyValidationMiddleware())

	// Health check (with public rate limiting)
	app.Get("/health", routeMiddleware.ApplyPublicRateLimit(), healthCheck)

	// API routes
	api := app.Group("/api/v1")

	// Setup auth routes (public)
	setupAuthRoutes(api, config, routeMiddleware)

	// Setup protected routes
	setupProtectedRoutes(api, config, routeMiddleware)
}

// healthCheck handles the health check endpoint
func healthCheck(c *fiber.Ctx) error {
	return c.JSON(models.APIResponse{
		Success: true,
		Message: "ZoomXML Service is healthy",
		Data: map[string]interface{}{
			"status":    "ok",
			"version":   "1.0.0",
			"timestamp": time.Now(),
			"services": map[string]string{
				"database":  "connected",
				"storage":   "connected",
				"api":       "running",
				"scheduler": "running",
			},
		},
	})
}

// setupAuthRoutes configures authentication routes (public)
func setupAuthRoutes(api fiber.Router, config RouteConfig, routeMiddleware *RouteMiddleware) {
	auth := api.Group("/auth")

	// Apply public rate limiting to auth routes
	auth.Use(routeMiddleware.ApplyPublicRateLimit())

	// Public auth endpoints
	auth.Post("/login", config.Handlers.Auth.Login)
	auth.Post("/logout", config.Handlers.Auth.Logout)
	auth.Post("/refresh", config.Handlers.Auth.RefreshToken)
}

// setupProtectedRoutes configures protected routes
func setupProtectedRoutes(api fiber.Router, config RouteConfig, routeMiddleware *RouteMiddleware) {
	// Apply authentication middleware
	protected := api.Use(middleware.AuthMiddleware(config.AuthService))

	// Apply authenticated rate limiting
	protected.Use(routeMiddleware.ApplyAuthenticatedRateLimit())

	// Setup auth protected routes
	setupAuthProtectedRoutes(protected, config, routeMiddleware)

	// Setup empresa routes
	setupEmpresaRoutes(protected, config, routeMiddleware)

	// Setup NFS-e routes
	setupNFSeRoutes(protected, config, routeMiddleware)
}

// setupAuthProtectedRoutes configures protected auth routes
func setupAuthProtectedRoutes(protected fiber.Router, config RouteConfig, routeMiddleware *RouteMiddleware) {
	protected.Get("/auth/me", config.Handlers.Auth.Me)
}

// setupEmpresaRoutes configures empresa management routes
func setupEmpresaRoutes(protected fiber.Router, config RouteConfig, routeMiddleware *RouteMiddleware) {
	empresas := protected.Group("/empresas")

	empresas.Post("/", config.Handlers.Empresa.Create)
	empresas.Get("/", config.Handlers.Empresa.List)
	empresas.Get("/:id", config.Handlers.Empresa.GetByID)
	empresas.Put("/:id", config.Handlers.Empresa.Update)
	empresas.Delete("/:id", config.Handlers.Empresa.Delete)
}

// setupNFSeRoutes configures NFS-e related routes
func setupNFSeRoutes(protected fiber.Router, config RouteConfig, routeMiddleware *RouteMiddleware) {
	nfse := protected.Group("/nfse")

	// Job management routes
	setupJobRoutes(nfse, config, routeMiddleware)

	// XML consumption routes
	setupXMLRoutes(nfse, config, routeMiddleware)

	// Statistics routes
	setupStatsRoutes(nfse, config, routeMiddleware)
}

// setupJobRoutes configures job management routes
func setupJobRoutes(nfse fiber.Router, config RouteConfig, routeMiddleware *RouteMiddleware) {
	// Manual sync trigger (heavy operation)
	nfse.Post("/sync", routeMiddleware.ApplyHeavyOperationsRateLimit(), config.Handlers.NFSe.HandleManualSync)

	// Job listing and monitoring
	nfse.Get("/jobs", config.Handlers.NFSe.HandleListJobs)

	// Advanced job routes
	jobs := nfse.Group("/jobs")
	jobs.Get("/:id", config.Handlers.Job.HandleGetJob)
	jobs.Delete("/:id", config.Handlers.Job.HandleCancelJob)
	jobs.Post("/:id/retry", config.Handlers.Job.HandleRetryJob)
	jobs.Get("/status/:status", config.Handlers.Job.HandleListJobsByStatus)
}

// setupXMLRoutes configures XML consumption routes
func setupXMLRoutes(nfse fiber.Router, config RouteConfig, routeMiddleware *RouteMiddleware) {
	// XML listing routes
	nfse.Get("/xmls", config.Handlers.NFSe.HandleListStoredXMLs)
	nfse.Get("/xmls/:competencia", config.Handlers.NFSe.HandleListXMLsByCompetencia)

	// XML retrieval routes
	nfse.Get("/xml/:competencia/:numero", config.Handlers.NFSe.HandleGetStoredXML)

	// XML download routes (with download rate limiting)
	nfse.Get("/xml/:competencia/:numero/download",
		routeMiddleware.ApplyDownloadRateLimit(),
		config.Handlers.NFSe.HandleDownloadXML)
}

// setupStatsRoutes configures statistics routes
func setupStatsRoutes(nfse fiber.Router, config RouteConfig, routeMiddleware *RouteMiddleware) {
	// General stats
	nfse.Get("/stats", config.Handlers.Stats.HandleGetGeneralStats)

	// Advanced stats routes
	stats := nfse.Group("/stats")
	stats.Get("/competencia/:competencia", config.Handlers.Stats.HandleGetStatsByCompetencia)
	stats.Get("/summary", config.Handlers.Stats.HandleGetSummaryStats)
	stats.Get("/performance", config.Handlers.Stats.HandleGetPerformanceMetrics)
	stats.Get("/health", config.Handlers.Stats.HandleGetHealthMetrics)
}
