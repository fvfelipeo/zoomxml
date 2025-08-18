package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zoomxml/internal/api/handlers"
	"github.com/zoomxml/internal/api/middleware"
)

// SetupRoutes configura todas as rotas da aplicação
func SetupRoutes(app *fiber.App) {
	// Criar handlers
	userHandler := handlers.NewUserHandler()
	companyHandler := handlers.NewCompanyHandler()
	credentialHandler := handlers.NewCredentialHandler()
	cnpjHandler := handlers.NewCNPJHandler()

	// Grupo API
	api := app.Group("/api")

	// Configurar rotas de usuários
	setupUserRoutes(api, userHandler)

	// Configurar rotas de empresas
	setupCompanyRoutes(api, companyHandler)

	// Configurar rotas de credenciais
	setupCredentialRoutes(api, credentialHandler)

	// Configurar rotas de CNPJ
	setupCNPJRoutes(api, cnpjHandler)

	// Configurar rotas de autenticação
	setupAuthRoutes(api)

	// Configurar rotas de estatísticas
	setupStatsRoutes(api)
}

// setupUserRoutes configura as rotas de gerenciamento de usuários
func setupUserRoutes(api fiber.Router, handler *handlers.UserHandler) {
	// Rotas de usuários (apenas admin com token especial)
	// Conforme especificação: apenas requisições com token admin podem criar/editar/excluir usuários
	users := api.Group("/users")
	users.Use(middleware.AdminTokenMiddleware()) // Token de admin definido no .env (ADMIN_TOKEN)

	users.Post("/", handler.CreateUser)      // POST /api/users - Criar usuário
	users.Get("/", handler.GetUsers)         // GET /api/users - Listar usuários
	users.Get("/:id", handler.GetUser)       // GET /api/users/:id - Obter usuário
	users.Patch("/:id", handler.UpdateUser)  // PATCH /api/users/:id - Editar usuário
	users.Delete("/:id", handler.DeleteUser) // DELETE /api/users/:id - Remover usuário
}

// setupCompanyRoutes configura as rotas de gerenciamento de empresas
func setupCompanyRoutes(api fiber.Router, handler *handlers.CompanyHandler) {
	companies := api.Group("/companies")

	// Aplicar autenticação opcional para todas as rotas de empresas
	// Isso permite que usuários não autenticados vejam empresas públicas
	companies.Use(middleware.OptionalAuthMiddleware())

	// CRUD de empresas
	companies.Post("/", middleware.AuthMiddleware(), handler.CreateCompany)                                        // Criar requer autenticação
	companies.Get("/", handler.GetCompanies)                                                                       // Listar (com regras de visibilidade)
	companies.Get("/:id", handler.GetCompany)                                                                      // Obter (com regras de visibilidade)
	companies.Patch("/:id", middleware.AuthMiddleware(), handler.UpdateCompany)                                    // Atualizar requer autenticação
	companies.Delete("/:id", middleware.AuthMiddleware(), middleware.AdminOnlyMiddleware(), handler.DeleteCompany) // Deletar apenas admin

	// Rotas para gerenciar membros de empresas restritas
	setupCompanyMemberRoutes(companies)

	// Rotas para gerenciar credenciais de empresas
	setupCompanyCredentialRoutes(companies)

	// Rotas para NFSe
	setupNFSeRoutes(companies)
}

// setupCompanyMemberRoutes configura as rotas de membros de empresas
func setupCompanyMemberRoutes(companies fiber.Router) {
	// Rotas para gerenciar membros (apenas para empresas restritas)
	members := companies.Group("/:companyId/members")
	members.Use(middleware.AuthMiddleware()) // Requer autenticação

	// TODO: Implementar handlers de membros
	// members.Post("/", memberHandler.AddMember)       // Adicionar membro
	// members.Get("/", memberHandler.GetMembers)       // Listar membros
	// members.Delete("/:userId", memberHandler.RemoveMember) // Remover membro
}

// setupCredentialRoutes configura as rotas de credenciais
func setupCredentialRoutes(api fiber.Router, handler *handlers.CredentialHandler) {
	// As rotas de credenciais são configuradas dentro das rotas de empresas
	// Esta função existe para manter a consistência com o padrão de setup
	// As rotas reais são configuradas em setupCompanyCredentialRoutes
}

// setupCompanyCredentialRoutes configura as rotas de credenciais de empresas
func setupCompanyCredentialRoutes(companies fiber.Router) {
	// Rotas para gerenciar credenciais
	credentials := companies.Group("/:company_id/credentials")
	credentials.Use(middleware.AuthMiddleware()) // Requer autenticação

	// Implementar handlers de credenciais
	credentialHandler := handlers.NewCredentialHandler()
	credentials.Post("/", credentialHandler.CreateCredential)      // Criar credencial
	credentials.Get("/", credentialHandler.GetCredentials)         // Listar credenciais
	credentials.Patch("/:id", credentialHandler.UpdateCredential)  // Atualizar credencial
	credentials.Delete("/:id", credentialHandler.DeleteCredential) // Deletar credencial
}

// setupNFSeRoutes configura as rotas de NFSe
func setupNFSeRoutes(companies fiber.Router) {
	// Rotas para NFSe
	nfse := companies.Group("/:company_id/nfse")
	nfse.Use(middleware.AuthMiddleware()) // Requer autenticação

	// Implementar handlers de NFSe
	nfseHandler := handlers.NewNFSeHandler()
	nfse.Post("/fetch", nfseHandler.FetchNFSeDocuments) // Buscar documentos NFSe
	nfse.Get("/", nfseHandler.GetNFSeDocuments)         // Listar documentos NFSe armazenados
}

// setupCNPJRoutes configura as rotas de consulta de CNPJ
func setupCNPJRoutes(api fiber.Router, handler *handlers.CNPJHandler) {
	// Rota para consultar CNPJ (requer autenticação)
	api.Get("/cnpj/:cnpj", middleware.AuthMiddleware(), handler.ConsultarCNPJ)
}

// setupAuthRoutes configura as rotas de autenticação
func setupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	authHandler := handlers.NewAuthHandler()

	// Rotas de autenticação
	auth.Post("/login", authHandler.Login)                                // Login de usuários
	auth.Post("/logout", middleware.AuthMiddleware(), authHandler.Logout) // Logout (requer autenticação)
	auth.Get("/me", middleware.AuthMiddleware(), authHandler.GetProfile)  // Perfil do usuário logado
}

// setupStatsRoutes configura as rotas de estatísticas
func setupStatsRoutes(api fiber.Router) {
	stats := api.Group("/stats")
	statsHandler := handlers.NewStatsHandler()

	// Rotas de estatísticas (requer autenticação)
	stats.Use(middleware.AuthMiddleware())
	stats.Get("/dashboard", statsHandler.GetDashboardStats)   // Estatísticas do dashboard
	stats.Get("/companies/:id", statsHandler.GetCompanyStats) // Estatísticas de empresa específica
}
