package repository

import (
	"context"

	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// UserPreferenceRepository handles user project preference operations
type UserPreferenceRepository struct {
	db *gorm.DB
}

// NewUserPreferenceRepository creates a new user preference repository
func NewUserPreferenceRepository(db *gorm.DB) *UserPreferenceRepository {
	return &UserPreferenceRepository{db: db}
}

// GetActiveProject retrieves the active project for a user
func (r *UserPreferenceRepository) GetActiveProject(ctx context.Context, userID uint) (*model.Project, error) {
	var pref model.UserProjectPreference
	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("user_id = ?", userID).
		First(&pref).Error

	if err != nil {
		return nil, err
	}

	return &pref.Project, nil
}

// GetActiveProjectID retrieves only the active project ID for a user
func (r *UserPreferenceRepository) GetActiveProjectID(ctx context.Context, userID uint) (uint, error) {
	var pref model.UserProjectPreference
	err := r.db.WithContext(ctx).
		Select("project_id").
		Where("user_id = ?", userID).
		First(&pref).Error

	if err != nil {
		return 0, err
	}

	return pref.ProjectID, nil
}

// SetActiveProject sets the active project for a user
func (r *UserPreferenceRepository) SetActiveProject(ctx context.Context, userID uint, projectID uint) error {
	// Use upsert (ON CONFLICT DO UPDATE)
	pref := model.UserProjectPreference{
		UserID:    userID,
		ProjectID: projectID,
	}

	return r.db.WithContext(ctx).
		Save(&pref).Error
}

// GetPreference retrieves the preference record for a user
func (r *UserPreferenceRepository) GetPreference(ctx context.Context, userID uint) (*model.UserProjectPreference, error) {
	var pref model.UserProjectPreference
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&pref).Error

	if err != nil {
		return nil, err
	}

	return &pref, nil
}
