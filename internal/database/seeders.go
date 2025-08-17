package database

import (
	"context"

	"github.com/zoomxml/config"
	"github.com/zoomxml/internal/logger"
	"github.com/zoomxml/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// SeedAdminUser cria o usuário admin usando as configurações do .env
func SeedAdminUser(ctx context.Context) error {
	cfg := config.Get()
	
	// Verificar se já existe um usuário admin
	exists, err := DB.NewSelect().
		Model((*models.User)(nil)).
		Where("role = 'admin'").
		Exists(ctx)

	if err != nil {
		return err
	}

	if exists {
		logger.InfoWithFields("Admin user already exists, skipping seed", map[string]any{
			"operation": "seed_admin_user",
		})
		return nil
	}

	// Hash da senha padrão
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Criar usuário admin usando o token do .env
	adminUser := &models.User{
		Name:     "Admin User",
		Email:    "admin@zoomxml.com",
		Password: string(hashedPassword),
		Token:    cfg.Auth.AdminToken, // Usar o token do .env
		Role:     "admin",
		Active:   true,
	}

	_, err = DB.NewInsert().Model(adminUser).Exec(ctx)
	if err != nil {
		return err
	}

	logger.InfoWithFields("Admin user created successfully", map[string]any{
		"operation": "seed_admin_user",
		"email":     adminUser.Email,
		"token":     adminUser.Token,
	})

	return nil
}

// SeedDevelopmentData cria dados iniciais para desenvolvimento
func SeedDevelopmentData(ctx context.Context) error {
	cfg := config.Get()
	
	// Só executar em ambiente de desenvolvimento
	if cfg.App.Env != "development" {
		return nil
	}

	logger.InfoWithFields("Seeding development data", map[string]any{
		"operation": "seed_development",
	})

	// Seed do usuário admin
	if err := SeedAdminUser(ctx); err != nil {
		return err
	}

	// Verificar se já existe uma empresa exemplo
	exists, err := DB.NewSelect().
		Model((*models.Company)(nil)).
		Where("cnpj = '00.000.000/0001-00'").
		Exists(ctx)

	if err != nil {
		return err
	}

	if !exists {
		// Criar empresa exemplo
		exampleCompany := &models.Company{
			Name:               "Empresa Exemplo LTDA",
			CNPJ:               "00.000.000/0001-00",
			TradeName:          "Empresa Exemplo",
			Address:            "Rua Exemplo",
			Number:             "123",
			District:           "Centro",
			City:               "São Paulo",
			State:              "SP",
			ZipCode:            "01000-000",
			Phone:              "(11) 1234-5678",
			Email:              "contato@exemplo.com.br",
			CompanySize:        "ME",
			MainActivity:       "Atividade exemplo",
			LegalNature:        "Sociedade Limitada",
			RegistrationStatus: "ATIVA",
			Restricted:         false,
			AutoFetch:          true,
			Active:             true,
		}

		_, err = DB.NewInsert().Model(exampleCompany).Exec(ctx)
		if err != nil {
			return err
		}

		logger.InfoWithFields("Example company created", map[string]any{
			"operation": "seed_development",
			"cnpj":      exampleCompany.CNPJ,
			"name":      exampleCompany.Name,
		})
	}

	logger.InfoWithFields("Development data seeding completed", map[string]any{
		"operation": "seed_development",
	})

	return nil
}

// RunSeeders executa todos os seeders necessários
func RunSeeders(ctx context.Context) error {
	cfg := config.Get()

	// Sempre criar o usuário admin (em qualquer ambiente)
	if err := SeedAdminUser(ctx); err != nil {
		logger.ErrorWithFields("Failed to seed admin user", err, map[string]any{
			"operation": "run_seeders",
		})
		return err
	}

	// Dados de desenvolvimento apenas em dev
	if cfg.App.Env == "development" {
		if err := SeedDevelopmentData(ctx); err != nil {
			logger.ErrorWithFields("Failed to seed development data", err, map[string]any{
				"operation": "run_seeders",
			})
			return err
		}
	}

	return nil
}
