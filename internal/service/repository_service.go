package service

import (
	"fmt"

	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/repository"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/crypto"
	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

// RepositoryService handles repository business logic
type RepositoryService struct {
	repo         *repository.RepositoryRepo
	platformRepo *repository.PlatformRepository
	encryptor    *crypto.Encryptor
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(
	repo *repository.RepositoryRepo,
	platformRepo *repository.PlatformRepository,
	cfg *config.Config,
) (*RepositoryService, error) {
	encryptor, err := crypto.NewEncryptor(cfg.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	return &RepositoryService{
		repo:         repo,
		platformRepo: platformRepo,
		encryptor:    encryptor,
	}, nil
}

// GitLabRepository represents a GitLab repository
type GitLabRepository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullPath      string `json:"full_path"`
	HTTPURL       string `json:"http_url"`
	SSHURL        string `json:"ssh_url"`
	DefaultBranch string `json:"default_branch"`
	Description   string `json:"description"`
}

// ListFromGitLab fetches repositories from GitLab
func (s *RepositoryService) ListFromGitLab(page, perPage int) ([]GitLabRepository, int, error) {
	// Get platform config
	platformConfig, err := s.platformRepo.GetConfig()
	if err != nil {
		return nil, 0, fmt.Errorf("platform not configured: %w", err)
	}

	// Decrypt token
	token, err := s.encryptor.Decrypt(platformConfig.AccessToken)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Create GitLab client
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(platformConfig.BaseURL))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// List projects
	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
		Membership: gitlab.Bool(true),
		Archived:   gitlab.Bool(false),
	}

	projects, resp, err := git.Projects.ListProjects(opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list GitLab projects: %w", err)
	}

	// Convert to our format
	var repos []GitLabRepository
	for _, p := range projects {
		repos = append(repos, GitLabRepository{
			ID:            int64(p.ID),
			Name:          p.Name,
			FullPath:      p.PathWithNamespace,
			HTTPURL:       p.HTTPURLToRepo,
			SSHURL:        p.SSHURLToRepo,
			DefaultBranch: p.DefaultBranch,
			Description:   p.Description,
		})
	}

	return repos, resp.TotalPages, nil
}

// List returns all repositories
func (s *RepositoryService) List(page, pageSize int) ([]model.Repository, int64, error) {
	return s.repo.List(page, pageSize)
}

// Get retrieves a repository
func (s *RepositoryService) Get(id uint) (*model.Repository, error) {
	return s.repo.Get(id)
}

// BatchImport imports multiple repositories from GitLab
func (s *RepositoryService) BatchImport(platformRepoIDs []int64, webhookCallbackURL string) error {
	// Get platform config
	platformConfig, err := s.platformRepo.GetConfig()
	if err != nil {
		return fmt.Errorf("platform not configured: %w", err)
	}

	// Decrypt token
	token, err := s.encryptor.Decrypt(platformConfig.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Create GitLab client
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(platformConfig.BaseURL))
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	var repos []model.Repository

	for _, platformRepoID := range platformRepoIDs {
		// Check if already imported
		existing, err := s.repo.GetByPlatformRepoID(platformConfig.ID, platformRepoID)
		if err == nil && existing != nil {
			continue // Already imported
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing repo: %w", err)
		}

		// Get project from GitLab
		project, _, err := git.Projects.GetProject(int(platformRepoID), nil)
		if err != nil {
			return fmt.Errorf("failed to get project %d: %w", platformRepoID, err)
		}

		// Create webhook
		webhookID, webhookURL, err := s.createWebhook(git, int(platformRepoID), webhookCallbackURL)
		if err != nil {
			return fmt.Errorf("failed to create webhook for project %d: %w", platformRepoID, err)
		}

		// Create repository record
		repo := model.Repository{
			PlatformID:     platformConfig.ID,
			PlatformRepoID: int64(project.ID),
			Name:           project.Name,
			FullPath:       project.PathWithNamespace,
			HTTPURL:        project.HTTPURLToRepo,
			SSHURL:         project.SSHURLToRepo,
			DefaultBranch:  project.DefaultBranch,
			WebhookID:      &webhookID,
			WebhookURL:     webhookURL,
			IsActive:       true,
		}

		repos = append(repos, repo)
	}

	if len(repos) > 0 {
		return s.repo.BatchCreate(repos)
	}

	return nil
}

// createWebhook creates a webhook for a GitLab project
func (s *RepositoryService) createWebhook(git *gitlab.Client, projectID int, callbackURL string) (int64, string, error) {
	opts := &gitlab.AddProjectHookOptions{
		URL:                   gitlab.String(callbackURL),
		MergeRequestsEvents:   gitlab.Bool(true),
		PushEvents:            gitlab.Bool(false),
		EnableSSLVerification: gitlab.Bool(false),
	}

	hook, _, err := git.Projects.AddProjectHook(projectID, opts)
	if err != nil {
		return 0, "", err
	}

	return int64(hook.ID), hook.URL, nil
}

// UpdateLLMModel updates the LLM model for a repository
func (s *RepositoryService) UpdateLLMModel(id uint, llmModelID *uint) error {
	return s.repo.UpdateLLMModel(id, llmModelID)
}

// Delete deletes a repository and removes webhook from GitLab
func (s *RepositoryService) Delete(id uint) error {
	// Get repository
	repo, err := s.repo.Get(id)
	if err != nil {
		return fmt.Errorf("repository not found: %w", err)
	}

	// Get platform config
	platformConfig, err := s.platformRepo.GetConfig()
	if err != nil {
		return fmt.Errorf("platform not configured: %w", err)
	}

	// Decrypt token
	token, err := s.encryptor.Decrypt(platformConfig.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Create GitLab client
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(platformConfig.BaseURL))
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Delete webhook if exists
	if repo.WebhookID != nil {
		_, err = git.Projects.DeleteProjectHook(int(repo.PlatformRepoID), int(*repo.WebhookID))
		// Ignore error if webhook already deleted
	}

	// Delete repository record
	return s.repo.Delete(id)
}
