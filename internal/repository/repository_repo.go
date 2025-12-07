package repository

import (
	"time"

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

// withRelations preloads common relationships for repository queries
func (r *RepositoryRepo) withRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("Platform").Preload("LLMProvider")
}

// List returns all repositories with pagination for a specific project
func (r *RepositoryRepo) List(projectID uint, page, pageSize int) ([]model.Repository, int64, error) {
	var repos []model.Repository
	var total int64

	// Count total
	if err := r.db.Model(&model.Repository{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := r.withRelations(r.db).Where("project_id = ?", projectID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&repos).Error

	return repos, total, err
}

// FindByProjectID retrieves all repositories for a project (no pagination)
func (r *RepositoryRepo) FindByProjectID(projectID uint) ([]model.Repository, error) {
	var repos []model.Repository
	err := r.withRelations(r.db).Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&repos).Error
	return repos, err
}

// Get retrieves a repository by ID with project validation
func (r *RepositoryRepo) Get(id uint, projectID uint) (*model.Repository, error) {
	var repo model.Repository
	err := r.withRelations(r.db).Where("id = ? AND project_id = ?", id, projectID).
		First(&repo).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// GetByID retrieves a repository by ID WITHOUT project validation.
//
// ⚠️ SECURITY WARNING: This method bypasses project isolation!
// Only use for system-level operations where project context is not available:
// - Webhook handlers (incoming events from Git platforms)
// - Background jobs (scheduled tasks)
// - System migrations
//
// ✅ For user requests, use Get(id, projectID) instead to enforce project isolation.
//
// Example valid usage:
//   func HandleWebhook(repoID uint) {
//       repo, _ := repoRepo.GetByID(repoID)  // OK: webhook context has no user
//   }
//
// Example INVALID usage:
//   func (h *Handler) GetRepository(c *gin.Context) {
//       repo, _ := h.repo.GetByID(id)  // ❌ WRONG: exposes other users' repos
//   }
func (r *RepositoryRepo) GetByID(id uint) (*model.Repository, error) {
	var repo model.Repository
	err := r.withRelations(r.db).First(&repo, id).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}


// GetByPlatformRepoID retrieves a repository by platform repo ID with project scope
func (r *RepositoryRepo) GetByPlatformRepoID(projectID uint, platformID uint, platformRepoID int64) (*model.Repository, error) {
	var repo model.Repository
	err := r.db.Where("project_id = ? AND platform_id = ? AND platform_repo_id = ?", 
		projectID, platformID, platformRepoID).First(&repo).Error
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

// SetWebhookStatus is the centralized function for updating webhook status
// All webhook status changes should go through this function to maintain consistency
func (r *RepositoryRepo) SetWebhookStatus(id uint, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"webhook_status":       status,
		"last_webhook_test_at": time.Now(),
	}
	if errorMsg != "" {
		updates["last_webhook_test_error"] = errorMsg
	} else {
		updates["last_webhook_test_error"] = ""
	}
	return r.db.Model(&model.Repository{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateWebhook updates webhook information after successful creation
func (r *RepositoryRepo) UpdateWebhook(id uint, webhookID int64, webhookURL string) error {
	updates := map[string]interface{}{
		"webhook_id":               &webhookID,
		"webhook_url":              webhookURL,
		"webhook_status":           model.WebhookStatusActive,
		"last_webhook_test_status": model.WebhookTestResultSuccess,
		"last_webhook_test_at":     time.Now(),
		"last_webhook_test_error":  "",
	}
	return r.db.Model(&model.Repository{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateWebhookTestStatus updates webhook status based on test result
// Deprecated: Use SetWebhookStatus instead for better clarity
func (r *RepositoryRepo) UpdateWebhookTestStatus(id uint, status string, errorMsg string) error {
	webhookStatus := model.WebhookStatusInactive // Default to inactive
	if status == model.WebhookTestResultSuccess {
		webhookStatus = model.WebhookStatusActive
	} else if status == model.WebhookTestResultFailed {
		webhookStatus = model.WebhookStatusInactive
	}

	updates := map[string]interface{}{
		"webhook_status":           webhookStatus,
		"last_webhook_test_status": status,
		"last_webhook_test_at":     time.Now(),
		"last_webhook_test_error":  errorMsg,
	}
	return r.db.Model(&model.Repository{}).Where("id = ?", id).Updates(updates).Error
}

