package repository

import (
	"context"

	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// ProjectRepository handles project database operations
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// FindByID retrieves a project by ID
func (r *ProjectRepository) FindByID(ctx context.Context, id uint) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// FindByUserID retrieves all projects for a user
func (r *ProjectRepository) FindByUserID(ctx context.Context, userID uint) ([]model.Project, error) {
	var projects []model.Project
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&projects).Error
	return projects, err
}

// FindByUserIDAndName retrieves a project by user ID and name
func (r *ProjectRepository) FindByUserIDAndName(ctx context.Context, userID uint, name string) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND name = ?", userID, name).
		First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// Update updates a project
func (r *ProjectRepository) Update(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

// Delete soft deletes a project
func (r *ProjectRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, id).Error
}

// Count counts projects for a user
func (r *ProjectRepository) Count(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Project{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
