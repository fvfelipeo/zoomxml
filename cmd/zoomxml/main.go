package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/zoomxml/config"
	"github.com/zoomxml/internal/api/middleware"
	"github.com/zoomxml/internal/api/routes"
	"github.com/zoomxml/internal/database"
	"github.com/zoomxml/internal/logger"
	"github.com/zoomxml/internal/services"
	"github.com/zoomxml/internal/storage"

	_ "github.com/zoomxml/docs" // Swagger docs
)

// @title ZoomXML API
// @version 1.0
// @description Sistema de Gerenciamento Multi-Empresarial de Documentos Fiscais
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.zoomxml.com/support
// @contact.email support@zoomxml.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8000
// @BasePath /api

// @securityDefinitions.apikey AdminToken
// @in header
// @name token
// @description Token de administrador para operações de usuários (ex: admin-secret-token)

// @securityDefinitions.apikey UserToken
// @in header
// @name token
// @description Token de usuário para autenticação (ex: U6HGHy4SDK)

func main() {
	// Carregar configuração
	cfg := config.Load()

	// Inicializar logger
	logger.Initialize()
	logger.Printf("Starting %s v%s in %s mode", cfg.App.Name, cfg.App.Version, cfg.App.Env)

	// Conectar ao banco de dados
	if err := database.Connect(); err != nil {
		logger.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Executar migrações automáticas
	ctx := context.Background()
	if err := database.AutoMigrate(ctx); err != nil {
		logger.Fatal("Failed to run migrations:", err)
	}

	// Executar seeders (criar usuário admin automaticamente)
	if err := database.RunSeeders(ctx); err != nil {
		logger.Fatal("Failed to run seeders:", err)
	}

	// Inicializar storage (MinIO)
	if err := storage.InitializeStorage(); err != nil {
		logger.Fatal("Failed to initialize storage:", err)
	}

	// Inicializar e iniciar o scheduler NFSe
	nfseScheduler := services.NewNFSeScheduler()
	if err := nfseScheduler.Start(); err != nil {
		logger.Fatal("Failed to start NFSe scheduler:", err)
	}

	// Graceful shutdown do scheduler
	defer nfseScheduler.Stop()

	// Criar aplicação Fiber
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: errorHandler,
	})

	// Middleware global
	setupMiddleware(app, cfg)

	// Configurar rotas
	routes.SetupRoutes(app)

	// Configurar graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Iniciar servidor
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Printf("Server starting on %s", addr)

	if err := app.Listen(addr); err != nil {
		logger.Fatal("Failed to start server:", err)
	}

	logger.Println("Server stopped")
}

// setupMiddleware configura os middlewares globais
func setupMiddleware(app *fiber.App, cfg *config.Config) {
	// Recover middleware
	app.Use(recover.New())

	// Logger middleware - usando nosso logger customizado
	app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skip: middleware.CombinedSkipper(
			middleware.HealthCheckSkipper,
			middleware.StaticFileSkipper,
		),
	}))

	// CORS middleware
	if cfg.Server.EnableCORS {
		allowOrigins := strings.Join(cfg.Server.AllowedOrigins, ",")
		allowCredentials := allowOrigins != "*" // Não permitir credentials com wildcard

		app.Use(cors.New(cors.Config{
			AllowOrigins:     allowOrigins,
			AllowMethods:     strings.Join(cfg.Server.AllowedMethods, ","),
			AllowHeaders:     strings.Join(cfg.Server.AllowedHeaders, ","),
			AllowCredentials: allowCredentials,
		}))
	}

	// Health check endpoint
	// @Summary Health Check
	// @Description Verifica o status da aplicação
	// @Tags health
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Status da aplicação"
	// @Router /health [get]
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"version":   cfg.App.Version,
		})
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)
}

// errorHandler manipula erros globais
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
		"code":  code,
	})
}
