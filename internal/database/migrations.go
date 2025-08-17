package database

import (
	"context"
	"fmt"
	"github.com/zoomxml/internal/logger"

	"github.com/uptrace/bun"
)

// Migration represents a database migration
type Migration struct {
	bun.BaseModel `bun:"table:migrations,alias:m"`

	ID        int64  `bun:"id,pk,autoincrement"`
	Name      string `bun:"name,unique,notnull"`
	AppliedAt string `bun:"applied_at,notnull,default:current_timestamp"`
}

// MigrationFunc represents a migration function
type MigrationFunc func(ctx context.Context, db *bun.DB) error

// MigrationItem represents a migration with its function
type MigrationItem struct {
	Name string
	Up   MigrationFunc
}

// GetMigrations returns all available migrations
func GetMigrations() []MigrationItem {
	return []MigrationItem{
		{
			Name: "001_create_users_table",
			Up:   createUsersTable,
		},
		{
			Name: "002_create_companies_table",
			Up:   createCompaniesTable,
		},
		{
			Name: "003_create_company_members_table",
			Up:   createCompanyMembersTable,
		},
		{
			Name: "004_create_company_credentials_table",
			Up:   createCompanyCredentialsTable,
		},
		{
			Name: "005_create_documents_table",
			Up:   createDocumentsTable,
		},
		{
			Name: "006_create_audit_logs_table",
			Up:   createAuditLogsTable,
		},
		{
			Name: "007_create_indexes",
			Up:   createIndexes,
		},
	}
}

// RunMigrations executes all pending migrations
func RunMigrations(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(ctx, DB); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations := GetMigrations()

	for _, migration := range migrations {
		// Check if migration has already been applied
		exists, err := DB.NewSelect().
			Model((*Migration)(nil)).
			Where("name = ?", migration.Name).
			Exists(ctx)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", migration.Name, err)
		}

		if exists {
			logger.Printf("Migration %s already applied, skipping", migration.Name)
			continue
		}

		// Run migration
		logger.Printf("Running migration: %s", migration.Name)
		if err := migration.Up(ctx, DB); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Name, err)
		}

		// Record migration as applied
		migrationRecord := &Migration{Name: migration.Name}
		if _, err := DB.NewInsert().Model(migrationRecord).Exec(ctx); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Name, err)
		}

		logger.Printf("Migration %s completed successfully", migration.Name)
	}

	logger.Println("All migrations completed successfully")
	return nil
}

// createMigrationsTable creates the migrations tracking table
func createMigrationsTable(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// Migration functions
func createUsersTable(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func createCompaniesTable(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS companies (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			cnpj VARCHAR(18) UNIQUE NOT NULL,
			address TEXT,
			city VARCHAR(255),
			state VARCHAR(2),
			zip_code VARCHAR(10),
			phone VARCHAR(20),
			email VARCHAR(255),
			restricted BOOLEAN NOT NULL DEFAULT false,
			auto_fetch BOOLEAN NOT NULL DEFAULT false,
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func createCompanyMembersTable(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS company_members (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			company_id INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, company_id)
		)
	`)
	return err
}

func createCompanyCredentialsTable(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS company_credentials (
			id SERIAL PRIMARY KEY,
			company_id INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
			type VARCHAR(100) NOT NULL,
			name VARCHAR(255) NOT NULL,
			login VARCHAR(255),
			encrypted_secret TEXT,
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func createDocumentsTable(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS documents (
			id SERIAL PRIMARY KEY,
			company_id INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			key VARCHAR(255),
			number VARCHAR(100),
			series VARCHAR(50),
			issue_date TIMESTAMP,
			due_date TIMESTAMP,
			amount DECIMAL(15,2),
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			storage_key VARCHAR(500),
			hash VARCHAR(255),
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func createAuditLogsTable(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS audit_logs (
			id SERIAL PRIMARY KEY,
			actor_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			action VARCHAR(50) NOT NULL,
			entity VARCHAR(100) NOT NULL,
			entity_id INTEGER,
			details JSONB,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func createIndexes(ctx context.Context, db *bun.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)",
		"CREATE INDEX IF NOT EXISTS idx_users_active ON users(active)",
		"CREATE INDEX IF NOT EXISTS idx_companies_cnpj ON companies(cnpj)",
		"CREATE INDEX IF NOT EXISTS idx_companies_restricted ON companies(restricted)",
		"CREATE INDEX IF NOT EXISTS idx_companies_active ON companies(active)",
		"CREATE INDEX IF NOT EXISTS idx_company_members_user_id ON company_members(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_company_members_company_id ON company_members(company_id)",
		"CREATE INDEX IF NOT EXISTS idx_company_credentials_company_id ON company_credentials(company_id)",
		"CREATE INDEX IF NOT EXISTS idx_company_credentials_type ON company_credentials(type)",
		"CREATE INDEX IF NOT EXISTS idx_documents_company_id ON documents(company_id)",
		"CREATE INDEX IF NOT EXISTS idx_documents_type ON documents(type)",
		"CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status)",
		"CREATE INDEX IF NOT EXISTS idx_documents_key ON documents(key)",
		"CREATE INDEX IF NOT EXISTS idx_documents_issue_date ON documents(issue_date)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return err
		}
	}

	return nil
}
