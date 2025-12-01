package repository

import (
	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// RepositoryRepo handles repository database operations
type RepositoryRepo struct {
	db *gorm.DB
}

// NewRepositoryRepo creates a new repository repo
func NewRepositoryRepo(db *gorm.DB) *RepositoryRepo {
	return &RepositoryRepo{db: db}
}

// List returns all repositories with pagination
func (r *RepositoryRepo) List(page, pageSize int) ([]model.Repository, int64, error) {
	var repos []model.Repository
	var total int64

	// Count total
	if err := r.db.Model(&model.Repository{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := r.db.Preload("Platform").Preload("LLMModel").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&repos).Error

	return repos, total, err
}

// Get retrieves a repository by ID
func (r *RepositoryRepo) Get(id uint) (*model.Repository, error) {
	var repo model.Repository
	err := r.db.Preload("Platform").Preload("LLMModel").First(&repo, id).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// GetByPlatformRepoID retrieves a repository by platform repo ID
func (r *RepositoryRepo) GetByPlatformRepoID(platformID uint, platformRepoID int64) (*model.Repository, error) {
	var repo model.Repository
	err := r.db.Where("platform_id = ? AND platform_repo_id = ?", platformID, platformRepoID).First(&repo).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// Create creates a new repository
func (r *RepositoryRepo) Create(repo *model.Repository) error {
	return r.db.Create(repo).Error
}

// BatchCreate creates multiple repositories
func (r *RepositoryRepo) BatchCreate(repos []model.Repository) error {
	return r.db.Create(&repos).Error
}

// Update updates a repository
func (r *RepositoryRepo) Update(repo *model.Repository) error {
	return r.db.Save(repo).Error
}

// Delete deletes a repository
func (r *RepositoryRepo) Delete(id uint) error {
	return r.db.Delete(&model.Repository{}, id).Error
}

// UpdateLLMModel updates the LLM model for a repository
func (r *RepositoryRepo) UpdateLLMModel(id uint, llmModelID *uint) error {
	return r.db.Model(&model.Repository{}).Where("id = ?", id).Update("llm_model_id", llmModelID).Error
}

// UpdateWebhook updates webhook information
func (r *RepositoryRepo) UpdateWebhook(id uint, webhookID *int64, webhookURL string) error {
	updates := map[string]interface{}{
		"webhook_id":  webhookID,
		"webhook_url": webhookURL,
	}
	return r.db.Model(&model.Repository{}).Where("id = ?", id).Updates(updates).Error
}
