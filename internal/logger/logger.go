package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zoomxml/config"
	"github.com/zoomxml/internal/models"
)

var Logger zerolog.Logger

// Initialize configures the global logger
func Initialize() {
	cfg := config.Get()

	// Configure output
	var output io.Writer = os.Stdout
	if cfg.IsDevelopment() {
		// Pretty logging for development
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global log level
	level := zerolog.InfoLevel
	switch cfg.Logger.Level {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(level)

	// Configure global logger
	Logger = zerolog.New(output).
		With().
		Timestamp().
		Str("service", cfg.App.Name).
		Str("version", cfg.App.Version).
		Logger()

	// Set as global logger
	log.Logger = Logger
}

// CredentialOperation represents the type of operation performed on credentials
type CredentialOperation string

const (
	OpCreate CredentialOperation = "create"
	OpRead   CredentialOperation = "read"
	OpUpdate CredentialOperation = "update"
	OpDelete CredentialOperation = "delete"
	OpList   CredentialOperation = "list"
)

// LogCredentialOperation logs credential-related operations for audit purposes
func LogCredentialOperation(ctx context.Context, user *models.User, operation CredentialOperation, companyID int64, credentialID *int64, success bool, errorMsg string) {
	event := Logger.Info()
	if !success {
		event = Logger.Error()
	}

	event = event.
		Str("type", "credential_audit").
		Str("operation", string(operation)).
		Int64("company_id", companyID).
		Bool("success", success)

	if user != nil {
		event = event.
			Int64("user_id", user.ID).
			Str("user_email", user.Email)
	}

	if credentialID != nil {
		event = event.Int64("credential_id", *credentialID)
	}

	if errorMsg != "" {
		event = event.Str("error", errorMsg)
	}

	event.Msg("Credential operation performed")
}

// LogSecurityEvent logs security-related events
func LogSecurityEvent(ctx context.Context, user *models.User, event string, details string) {
	logEvent := Logger.Warn().
		Str("type", "security").
		Str("event", event)

	if user != nil {
		logEvent = logEvent.
			Int64("user_id", user.ID).
			Str("user_email", user.Email)
	}

	if details != "" {
		logEvent = logEvent.Str("details", details)
	}

	logEvent.Msg("Security event occurred")
}

// LogError logs error messages with context
func LogError(ctx context.Context, operation string, err error, details map[string]any) {
	event := Logger.Error().
		Str("operation", operation).
		Err(err)

	for key, value := range details {
		event = event.Interface(key, value)
	}

	event.Msg("Operation failed")
}

// LogInfo logs informational messages
func LogInfo(ctx context.Context, operation string, message string, details map[string]any) {
	event := Logger.Info().
		Str("operation", operation)

	for key, value := range details {
		event = event.Interface(key, value)
	}

	event.Msg(message)
}

// LogWarning logs warning messages
func LogWarning(ctx context.Context, operation string, message string, details map[string]any) {
	event := Logger.Warn().
		Str("operation", operation)

	for key, value := range details {
		event = event.Interface(key, value)
	}

	event.Msg(message)
}

// LogDebug logs debug messages (only in development)
func LogDebug(ctx context.Context, operation string, message string, details map[string]any) {
	event := Logger.Debug().
		Str("operation", operation)

	for key, value := range details {
		event = event.Interface(key, value)
	}

	event.Msg(message)
}

// LogDatabaseOperation logs database operations for debugging
func LogDatabaseOperation(ctx context.Context, operation string, table string, duration time.Duration, err error) {
	event := Logger.Debug().
		Str("type", "database").
		Str("operation", operation).
		Str("table", table).
		Dur("duration", duration)

	if err != nil {
		event = event.Err(err)
		event.Msg("Database operation failed")
	} else {
		event.Msg("Database operation completed")
	}
}

// LogAPIRequest logs API requests for monitoring
func LogAPIRequest(ctx context.Context, method string, path string, userID *int64, statusCode int, duration time.Duration) {
	event := Logger.Info().
		Str("type", "api_request").
		Str("method", method).
		Str("path", path).
		Int("status_code", statusCode).
		Dur("duration", duration)

	if userID != nil {
		event = event.Int64("user_id", *userID)
	}

	event.Msg("API request processed")
}

// LogEncryptionOperation logs encryption/decryption operations
func LogEncryptionOperation(ctx context.Context, operation string, success bool, errorMsg string) {
	event := Logger.Debug().
		Str("type", "crypto").
		Str("operation", operation).
		Bool("success", success)

	if errorMsg != "" {
		event = event.Str("error", errorMsg)
	}

	event.Msg("Encryption operation performed")
}

// LogPermissionCheck logs permission validation attempts
func LogPermissionCheck(ctx context.Context, user *models.User, resource string, action string, allowed bool, reason string) {
	event := Logger.Info().
		Str("type", "permission_check").
		Str("resource", resource).
		Str("action", action).
		Bool("allowed", allowed)

	if user != nil {
		event = event.
			Int64("user_id", user.ID).
			Str("user_email", user.Email)
	}

	if reason != "" {
		event = event.Str("reason", reason)
	}

	event.Msg("Permission check performed")
}

// Standard logging functions that can replace other loggers

// Print logs a message at info level (compatible with standard logger.Print)
func Print(v ...any) {
	Logger.Info().Msg(fmt.Sprint(v...))
}

// Printf logs a formatted message at info level (compatible with standard logger.Printf)
func Printf(format string, v ...any) {
	Logger.Info().Msgf(format, v...)
}

// Println logs a message at info level (compatible with standard logger.Println)
func Println(v ...any) {
	Logger.Info().Msg(fmt.Sprintln(v...))
}

// Fatal logs a message at fatal level and exits (compatible with standard logger.Fatal)
func Fatal(v ...any) {
	Logger.Fatal().Msg(fmt.Sprint(v...))
}

// Fatalf logs a formatted message at fatal level and exits (compatible with standard logger.Fatalf)
func Fatalf(format string, v ...any) {
	Logger.Fatal().Msgf(format, v...)
}

// Fatalln logs a message at fatal level and exits (compatible with standard logger.Fatalln)
func Fatalln(v ...any) {
	Logger.Fatal().Msg(fmt.Sprintln(v...))
}

// Panic logs a message at panic level and panics (compatible with standard logger.Panic)
func Panic(v ...any) {
	Logger.Panic().Msg(fmt.Sprint(v...))
}

// Panicf logs a formatted message at panic level and panics (compatible with standard logger.Panicf)
func Panicf(format string, v ...any) {
	Logger.Panic().Msgf(format, v...)
}

// Panicln logs a message at panic level and panics (compatible with standard logger.Panicln)
func Panicln(v ...any) {
	Logger.Panic().Msg(fmt.Sprintln(v...))
}

// Structured logging functions

// WithField creates a logger with a single field
func WithField(key string, value any) *zerolog.Event {
	return Logger.Info().Interface(key, value)
}

// WithFields creates a logger with multiple fields
func WithFields(fields map[string]any) *zerolog.Event {
	event := Logger.Info()
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	return event
}

// WithError creates a logger with an error field
func WithError(err error) *zerolog.Event {
	return Logger.Error().Err(err)
}

// WithContext creates a logger with context
func WithContext(ctx context.Context) *zerolog.Event {
	return Logger.Info().Ctx(ctx)
}

// Level-specific structured logging

// InfoWithFields logs an info message with fields
func InfoWithFields(message string, fields map[string]any) {
	event := Logger.Info()
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	event.Msg(message)
}

// ErrorWithFields logs an error message with fields
func ErrorWithFields(message string, err error, fields map[string]any) {
	event := Logger.Error().Err(err)
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	event.Msg(message)
}

// WarnWithFields logs a warning message with fields
func WarnWithFields(message string, fields map[string]any) {
	event := Logger.Warn()
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	event.Msg(message)
}

// DebugWithFields logs a debug message with fields
func DebugWithFields(message string, fields map[string]any) {
	event := Logger.Debug()
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	event.Msg(message)
}
