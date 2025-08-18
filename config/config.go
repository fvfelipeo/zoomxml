package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	App           AppConfig
	Database      DatabaseConfig
	Storage       StorageConfig
	Auth          AuthConfig
	Server        ServerConfig
	Logger        LoggerConfig
	RateLimit     RateLimitConfig
	NFSeScheduler NFSeSchedulerConfig
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name    string
	Version string
	Env     string
	Debug   bool
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// StorageConfig holds MinIO/S3 storage configuration
type StorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret           string
	JWTExpirationHours  int
	RefreshTokenExpiry  time.Duration
	PasswordMinLength   int
	EnableRefreshTokens bool
	AdminToken          string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host           string
	Port           int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	EnableCORS     bool
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level      string
	Format     string
	Output     string
	EnableFile bool
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enable             bool
	PublicRPM          int
	AuthenticatedRPM   int
	HeavyOperationsRPM int
	DownloadRPM        int
}

// NFSeSchedulerConfig holds NFSe scheduler configuration
type NFSeSchedulerConfig struct {
	Enabled         bool
	Interval        string
	FetchDaysBack   int
	MaxPagesPerRun  int
	APIDelaySeconds int
}

var appConfig *Config

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		App: AppConfig{
			Name:    getEnv("APP_NAME", "ZoomXML"),
			Version: getEnv("APP_VERSION", "1.0.0"),
			Env:     getEnv("APP_ENV", "development"),
			Debug:   getEnvBool("APP_DEBUG", false),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Database:        getEnv("DB_NAME", "nfse_metadata"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Storage: StorageConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "admin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "password123"),
			Bucket:    getEnv("MINIO_BUCKET", "nfse-storage"),
			UseSSL:    getEnvBool("MINIO_USE_SSL", false),
			Region:    getEnv("MINIO_REGION", "us-east-1"),
		},
		Auth: AuthConfig{
			JWTSecret:           getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			JWTExpirationHours:  getEnvInt("JWT_EXPIRATION_HOURS", 24),
			RefreshTokenExpiry:  getEnvDuration("REFRESH_TOKEN_EXPIRY", 168*time.Hour),
			PasswordMinLength:   getEnvInt("PASSWORD_MIN_LENGTH", 8),
			EnableRefreshTokens: getEnvBool("ENABLE_REFRESH_TOKENS", true),
			AdminToken:          getEnv("ADMIN_TOKEN", "admin-secret-token"),
		},
		Server: ServerConfig{
			Host:           getEnv("SERVER_HOST", "0.0.0.0"),
			Port:           getEnvInt("PORT", 3000),
			ReadTimeout:    getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:   getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:    getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			EnableCORS:     getEnvBool("ENABLE_CORS", true),
			AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods: getEnvSlice("ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvSlice("ALLOWED_HEADERS", []string{"*"}),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			EnableFile: getEnvBool("LOG_ENABLE_FILE", false),
			FilePath:   getEnv("LOG_FILE_PATH", "logs/zoomxml.log"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
		},
		RateLimit: RateLimitConfig{
			Enable:             getEnvBool("ENABLE_RATE_LIMIT", true),
			PublicRPM:          getEnvInt("PUBLIC_RPM", 100),
			AuthenticatedRPM:   getEnvInt("AUTHENTICATED_RPM", 1000),
			HeavyOperationsRPM: getEnvInt("HEAVY_OPERATIONS_RPM", 10),
			DownloadRPM:        getEnvInt("DOWNLOAD_RPM", 50),
		},
		NFSeScheduler: NFSeSchedulerConfig{
			Enabled:         getEnvBool("NFSE_SCHEDULER_ENABLED", true),
			Interval:        getEnv("NFSE_SCHEDULER_INTERVAL", "24h"),
			FetchDaysBack:   getEnvInt("NFSE_FETCH_DAYS_BACK", 90),
			MaxPagesPerRun:  getEnvInt("NFSE_MAX_PAGES_PER_RUN", 10),
			APIDelaySeconds: getEnvInt("NFSE_API_DELAY_SECONDS", 2),
		},
	}

	appConfig = config
	return config
}

// Get returns the global configuration instance
func Get() *Config {
	if appConfig == nil {
		return Load()
	}
	return appConfig
}

// Helper functions for environment variable parsing
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return fallback
}

// IsDevelopment returns true if the app is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if the app is running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
