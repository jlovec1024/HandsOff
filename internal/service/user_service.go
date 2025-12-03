package service

import (
	"fmt"

	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// UserService handles user-related business logic
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// CreateUser creates a new user with default project using explicit transaction.
//
// ✅ TRANSACTION GUARANTEE: This method ensures atomicity using GORM transactions.
// If any step fails (user creation, project creation, or preference setting),
// ALL changes are rolled back automatically - no orphaned data is left behind.
//
// Steps performed in transaction:
//   1. Create user record (password hashed by BeforeCreate hook)
//   2. Create default project (only if user.IsActive == true)
//   3. Create user preference to set project as active
//
// Error handling:
//   - Returns wrapped error with context if any step fails
//   - Database automatically rolls back on error
//   - Safe to call multiple times (idempotent if username unique constraint honored)
//
// Example usage:
//   user := &model.User{Username: "alice", Password: "secret", IsActive: true}
//   if err := userService.CreateUser(user); err != nil {
//       // Either all steps succeeded, or none did (rollback)
//   }
func (s *UserService) CreateUser(user *model.User) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Step 1: Create user (password will be hashed by BeforeCreate hook)
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Step 2: Create default project only if user is active
		if user.IsActive {
			if err := s.createDefaultProject(tx, user); err != nil {
				return err
			}
		}

		return nil
	})
}

// CreateUserWithProject creates user and custom project in a single transaction
func (s *UserService) CreateUserWithProject(user *model.User, projectName, projectDesc string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Step 1: Create user
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Step 2: Create custom project
		project := &model.Project{
			Name:        projectName,
			Description: projectDesc,
			UserID:      user.ID,
		}

		if err := tx.Create(project).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		// Step 3: Set as default project
		preference := &model.UserProjectPreference{
			UserID:    user.ID,
			ProjectID: project.ID,
		}

		if err := tx.Create(preference).Error; err != nil {
			return fmt.Errorf("failed to set default project: %w", err)
		}

		return nil
	})
}

// createDefaultProject creates a default project and preference for a user
// This is extracted from the original AfterCreate hook for better testability
func (s *UserService) createDefaultProject(tx *gorm.DB, user *model.User) error {
	// Create default project
	project := &model.Project{
		Name:        fmt.Sprintf("%s-project", user.Username),
		Description: fmt.Sprintf("Default project for %s", user.Username),
		UserID:      user.ID,
	}

	if err := tx.Create(project).Error; err != nil {
		return fmt.Errorf("failed to create default project: %w", err)
	}

	// Create user preference to set this as active project
	preference := &model.UserProjectPreference{
		UserID:    user.ID,
		ProjectID: project.ID,
	}

	if err := tx.Create(preference).Error; err != nil {
		return fmt.Errorf("failed to set active project: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id uint) (*model.User, error) {
	var user model.User
	if err := s.db.Preload("Projects").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(user *model.User) error {
	return s.db.Save(user).Error
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(id uint) error {
	return s.db.Delete(&model.User{}, id).Error
}

// ActivateUser activates a user and creates default project if not exists
func (s *UserService) ActivateUser(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get user
		var user model.User
		if err := tx.First(&user, id).Error; err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}

		// Already active, nothing to do
		if user.IsActive {
			return nil
		}

		// Activate user
		user.IsActive = true
		if err := tx.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to activate user: %w", err)
		}

		// Check if user has any projects
		var projectCount int64
		if err := tx.Model(&model.Project{}).Where("user_id = ?", id).Count(&projectCount).Error; err != nil {
			return fmt.Errorf("failed to count projects: %w", err)
		}

		// If no projects exist, create default project
		if projectCount == 0 {
			if err := s.createDefaultProject(tx, &user); err != nil {
				return err
			}
		}

		return nil
	})
}

// DeactivateUser deactivates a user
func (s *UserService) DeactivateUser(id uint) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).Update("is_active", false).Error
}

// ListUsers returns all users with pagination
func (s *UserService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// Count total
	if err := s.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := s.db.Preload("Projects").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUserDefaultProject returns the user's default/active project
func (s *UserService) GetUserDefaultProject(userID uint) (*model.Project, error) {
	// Try to get active project from preferences
	var pref model.UserProjectPreference
	if err := s.db.Where("user_id = ?", userID).First(&pref).Error; err == nil {
		var project model.Project
		if err := s.db.First(&project, pref.ProjectID).Error; err != nil {
			return nil, fmt.Errorf("failed to find default project: %w", err)
		}
		return &project, nil
	}

	// If no preference set, get user's first project
	var project model.Project
	if err := s.db.Where("user_id = ?", userID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("no projects found for user: %w", err)
	}

	return &project, nil
}

// SetUserDefaultProject sets a project as user's default
func (s *UserService) SetUserDefaultProject(userID, projectID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify project belongs to user
		var project model.Project
		if err := tx.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
			return fmt.Errorf("project not found or access denied: %w", err)
		}

		// Update or create preference
		preference := model.UserProjectPreference{
			UserID:    userID,
			ProjectID: projectID,
		}

		// Use Save to update if exists, create if not
		if err := tx.Save(&preference).Error; err != nil {
			return fmt.Errorf("failed to set default project: %w", err)
		}

		return nil
	})
}
