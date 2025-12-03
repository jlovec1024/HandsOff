package initializer

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/service"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/database"
	"github.com/handsoff/handsoff/pkg/logger"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// Initialize performs all startup initialization tasks
// 1. Database table migration (idempotent)
// 2. Default admin user creation (idempotent)
func Initialize(db *gorm.DB, cfg *config.Config, log *logger.Logger) error {
	// Step 1: Auto-migrate database schema
	if err := database.AutoMigrate(db); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}
	log.Info("Database schema migration completed successfully")

	// Step 2: Create default admin user if not exists
	if err := ensureAdminUser(db, cfg, log); err != nil {
		return fmt.Errorf("admin user initialization failed: %w", err)
	}

	return nil
}

// ensureAdminUser creates default admin user if not exists (idempotent)
func ensureAdminUser(db *gorm.DB, cfg *config.Config, log *logger.Logger) error {
	// Check if admin already exists
	var existingUser model.User
	err := db.Where("username = ?", "admin").First(&existingUser).Error

	if err == nil {
		// Admin already exists, skip creation
		log.Info("Admin user already exists, skipping creation",
			"username", existingUser.Username,
			"email", existingUser.Email)
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		// Database error
		return fmt.Errorf("failed to check admin user: %w", err)
	}

	// Admin doesn't exist, create it
	userService := service.NewUserService(db)
	admin := &model.User{
		Username: "admin",
		Password: cfg.Admin.InitialPassword, // Will be hashed automatically by BeforeCreate hook
		Email:    cfg.Admin.Email,
		IsActive: true,
	}

	if err := userService.CreateUser(admin); err != nil {
		// Check if error is due to duplicate entry using database-specific error types
		if isDuplicateKeyError(err) {
			log.Info("Admin user already exists (created by another instance), skipping creation")
			return nil
		}
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Info("Default admin user created successfully",
		"username", admin.Username,
		"email", admin.Email)
	log.Warn("⚠️  SECURITY: Please change the default password immediately after first login!")

	return nil
}

// isDuplicateKeyError checks if the error is a duplicate key/unique constraint violation
// Supports MySQL, SQLite, and PostgreSQL
func isDuplicateKeyError(err error) bool {
	// MySQL error 1062: Duplicate entry
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}

	// SQLite error: UNIQUE constraint failed
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
		return true
	}

	// PostgreSQL and other databases: use GORM's ErrDuplicatedKey
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
