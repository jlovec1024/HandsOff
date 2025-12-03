package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Worker   WorkerConfig
	Security SecurityConfig
	Admin    AdminConfig
	Log      LogConfig
	Git      GitConfig
	CORS     CORSConfig
}

// AppConfig contains application-level settings
type AppConfig struct {
	Name string
	Env  string // development, production
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port int
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Type string // sqlite, mysql, postgres
	DSN  string // Data Source Name
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// WorkerConfig contains async worker settings
type WorkerConfig struct {
	Concurrency int
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	JWTSecret     string
	JWTExpiry     time.Duration
	EncryptionKey string
}

// AdminConfig contains default admin user settings
type AdminConfig struct {
	InitialPassword string
	Email           string
}

// LogConfig contains logging settings
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, console
	File   string
}

// GitConfig contains Git operation settings
type GitConfig struct {
	CloneTimeout time.Duration
	TempDir      string
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	AllowedOrigins []string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Read .env file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// .env file not found, continue with environment variables only
	}

	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "HandsOff"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Server: ServerConfig{
			Port: getEnvInt("API_PORT", 8080),
		},
		Database: DatabaseConfig{
			Type: getEnv("DB_TYPE", "sqlite"),
			DSN:  getEnv("DB_DSN", "data/app.db"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Worker: WorkerConfig{
			Concurrency: getEnvInt("WORKER_CONCURRENCY", 10),
		},
		Security: SecurityConfig{
			JWTSecret:     getEnv("JWT_SECRET", "change_this_to_a_random_secret_key"),
			JWTExpiry:     getEnvDuration("JWT_EXPIRY", 24*time.Hour),
			EncryptionKey: getEnv("ENCRYPTION_KEY", "CHANGE_THIS_TO_BASE64_ENCODED_32_BYTES_KEY"),
		},
		Admin: AdminConfig{
			InitialPassword: getEnv("ADMIN_INITIAL_PASSWORD", "admin123"),
			Email:           getEnv("ADMIN_EMAIL", "admin@jlovec.net"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "console"),
			File:   getEnv("LOG_FILE", "logs/app.log"),
		},
		Git: GitConfig{
			CloneTimeout: getEnvDuration("GIT_CLONE_TIMEOUT", 300*time.Second),
			TempDir:      getEnv("GIT_TEMP_DIR", "./temp/git"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Security.JWTSecret == "change_this_to_a_random_secret_key" {
		return fmt.Errorf("JWT_SECRET must be changed from default value")
	}

	if c.Security.EncryptionKey == "CHANGE_THIS_TO_BASE64_ENCODED_32_BYTES_KEY" {
		return fmt.Errorf("ENCRYPTION_KEY must be changed from default value")
	}

	if c.Database.Type != "sqlite" && c.Database.Type != "mysql" && c.Database.Type != "postgres" {
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}

	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := viper.GetString(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if viper.IsSet(key) {
		return viper.GetInt(key)
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if viper.IsSet(key) {
		return viper.GetDuration(key)
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
