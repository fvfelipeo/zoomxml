package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Application settings
	App AppConfig `json:"app"`

	// Database configuration
	Database DatabaseConfig `json:"database"`

	// Storage configuration
	Storage StorageConfig `json:"storage"`

	// Authentication configuration
	Auth AuthConfig `json:"auth"`

	// Server configuration
	Server ServerConfig `json:"server"`

	// Scheduler configuration
	Scheduler SchedulerConfig `json:"scheduler"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `json:"rate_limit"`
}

// AppConfig holds application-specific settings
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`

	// Connection pool settings
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// StorageConfig holds MinIO/S3 storage settings
type StorageConfig struct {
	Endpoint   string `json:"endpoint"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	BucketName string `json:"bucket_name"`
	UseSSL     bool   `json:"use_ssl"`
	Region     string `json:"region"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	JWTSecret           string        `json:"jwt_secret"`
	JWTExpirationHours  int           `json:"jwt_expiration_hours"`
	RefreshTokenExpiry  time.Duration `json:"refresh_token_expiry"`
	PasswordMinLength   int           `json:"password_min_length"`
	EnableRefreshTokens bool          `json:"enable_refresh_tokens"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`

	// CORS settings
	EnableCORS     bool     `json:"enable_cors"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// SchedulerConfig holds scheduler settings
type SchedulerConfig struct {
	EnableAutoSync       bool          `json:"enable_auto_sync"`
	DefaultSyncInterval  time.Duration `json:"default_sync_interval"`
	JobProcessorInterval time.Duration `json:"job_processor_interval"`
	MaxRetries           int           `json:"max_retries"`
	RetryBackoffFactor   float64       `json:"retry_backoff_factor"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	EnableFile bool   `json:"enable_file"`
	FilePath   string `json:"file_path"`
	MaxSize    int    `json:"max_size"` // MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"` // days
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	EnableRateLimit    bool `json:"enable_rate_limit"`
	PublicRPM          int  `json:"public_rpm"`
	AuthenticatedRPM   int  `json:"authenticated_rpm"`
	HeavyOperationsRPM int  `json:"heavy_operations_rpm"`
	DownloadRPM        int  `json:"download_rpm"`
}

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't return error
		fmt.Printf("Warning: .env file not found or could not be loaded: %v\n", err)
	}

	config := &Config{
		App:       loadAppConfig(),
		Database:  loadDatabaseConfig(),
		Storage:   loadStorageConfig(),
		Auth:      loadAuthConfig(),
		Server:    loadServerConfig(),
		Scheduler: loadSchedulerConfig(),
		Logging:   loadLoggingConfig(),
		RateLimit: loadRateLimitConfig(),
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %v", err)
	}

	return config, nil
}

// loadAppConfig loads application configuration
func loadAppConfig() AppConfig {
	return AppConfig{
		Name:        getEnv("APP_NAME", "ZoomXML"),
		Version:     getEnv("APP_VERSION", "1.0.0"),
		Environment: getEnv("APP_ENV", "development"),
		Debug:       getEnvBool("APP_DEBUG", false),
	}
}

// loadDatabaseConfig loads database configuration
func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnvInt("DB_PORT", 5432),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "password"),
		DBName:          getEnv("DB_NAME", "nfse_metadata"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

// loadStorageConfig loads storage configuration
func loadStorageConfig() StorageConfig {
	return StorageConfig{
		Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:  getEnv("MINIO_ACCESS_KEY", "admin"),
		SecretKey:  getEnv("MINIO_SECRET_KEY", "password123"),
		BucketName: getEnv("MINIO_BUCKET", "nfse-storage"),
		UseSSL:     getEnvBool("MINIO_USE_SSL", false),
		Region:     getEnv("MINIO_REGION", "us-east-1"),
	}
}

// loadAuthConfig loads authentication configuration
func loadAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:           getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpirationHours:  getEnvInt("JWT_EXPIRATION_HOURS", 24),
		RefreshTokenExpiry:  getEnvDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		PasswordMinLength:   getEnvInt("PASSWORD_MIN_LENGTH", 8),
		EnableRefreshTokens: getEnvBool("ENABLE_REFRESH_TOKENS", true),
	}
}

// loadServerConfig loads server configuration
func loadServerConfig() ServerConfig {
	return ServerConfig{
		Host:         getEnv("SERVER_HOST", "0.0.0.0"),
		Port:         getEnvInt("PORT", 3000),
		ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),

		EnableCORS:     getEnvBool("ENABLE_CORS", true),
		AllowedOrigins: getEnvStringSlice("ALLOWED_ORIGINS", []string{"*"}),
		AllowedMethods: getEnvStringSlice("ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AllowedHeaders: getEnvStringSlice("ALLOWED_HEADERS", []string{"*"}),
	}
}

// loadSchedulerConfig loads scheduler configuration
func loadSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		EnableAutoSync:       getEnvBool("ENABLE_AUTO_SYNC", true),
		DefaultSyncInterval:  getEnvDuration("DEFAULT_SYNC_INTERVAL", 1*time.Hour),
		JobProcessorInterval: getEnvDuration("JOB_PROCESSOR_INTERVAL", 30*time.Second),
		MaxRetries:           getEnvInt("MAX_RETRIES", 5),
		RetryBackoffFactor:   getEnvFloat("RETRY_BACKOFF_FACTOR", 2.0),
	}
}

// loadLoggingConfig loads logging configuration
func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:      getEnv("LOG_LEVEL", "info"),
		Format:     getEnv("LOG_FORMAT", "json"),
		Output:     getEnv("LOG_OUTPUT", "stdout"),
		EnableFile: getEnvBool("LOG_ENABLE_FILE", false),
		FilePath:   getEnv("LOG_FILE_PATH", "logs/zoomxml.log"),
		MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
		MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
		MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
	}
}

// loadRateLimitConfig loads rate limiting configuration
func loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		EnableRateLimit:    getEnvBool("ENABLE_RATE_LIMIT", true),
		PublicRPM:          getEnvInt("PUBLIC_RPM", 100),
		AuthenticatedRPM:   getEnvInt("AUTHENTICATED_RPM", 1000),
		HeavyOperationsRPM: getEnvInt("HEAVY_OPERATIONS_RPM", 10),
		DownloadRPM:        getEnvInt("DOWNLOAD_RPM", 50),
	}
}

// Utility functions for reading environment variables

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as integer with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets environment variable as boolean with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvFloat gets environment variable as float64 with default value
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvDuration gets environment variable as duration with default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvStringSlice gets environment variable as string slice with default value
func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// validateConfig validates the loaded configuration
func validateConfig(config *Config) error {
	// Validate required fields
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Storage.Endpoint == "" {
		return fmt.Errorf("storage endpoint is required")
	}
	if config.Storage.AccessKey == "" {
		return fmt.Errorf("storage access key is required")
	}
	if config.Storage.SecretKey == "" {
		return fmt.Errorf("storage secret key is required")
	}
	if config.Storage.BucketName == "" {
		return fmt.Errorf("storage bucket name is required")
	}

	if config.Auth.JWTSecret == "" || config.Auth.JWTSecret == "your-secret-key-change-in-production" {
		return fmt.Errorf("JWT secret must be set and changed from default")
	}

	// Validate ranges
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}
	if config.Database.Port < 1 || config.Database.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}

	if config.Auth.JWTExpirationHours < 1 {
		return fmt.Errorf("JWT expiration hours must be at least 1")
	}
	if config.Auth.PasswordMinLength < 4 {
		return fmt.Errorf("password minimum length must be at least 4")
	}

	return nil
}
