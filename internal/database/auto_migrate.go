package database

import (
	"context"
	"github.com/zoomxml/internal/logger"

	"github.com/zoomxml/internal/models"
)

// AutoMigrate executa migração automática usando os modelos do Bun
func AutoMigrate(ctx context.Context) error {
	logger.Println("Starting auto-migration...")

	// Registrar modelos
	models.RegisterModels(DB)

	// Criar tabelas automaticamente
	allModels := models.GetAllModels()
	for _, model := range allModels {
		if _, err := DB.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			return err
		}
	}

	logger.Println("Auto-migration completed successfully")
	return nil
}

// DropAllTables remove todas as tabelas (usar apenas em desenvolvimento/testes)
func DropAllTables(ctx context.Context) error {
	logger.Println("Dropping all tables...")

	allModels := models.GetAllModels()
	// Reverter ordem para evitar problemas de foreign key
	for i := len(allModels) - 1; i >= 0; i-- {
		model := allModels[i]
		if _, err := DB.NewDropTable().Model(model).IfExists().Cascade().Exec(ctx); err != nil {
			return err
		}
	}

	logger.Println("All tables dropped successfully")
	return nil
}

// ResetDatabase remove e recria todas as tabelas (usar apenas em desenvolvimento/testes)
func ResetDatabase(ctx context.Context) error {
	if err := DropAllTables(ctx); err != nil {
		return err
	}
	return AutoMigrate(ctx)
}
