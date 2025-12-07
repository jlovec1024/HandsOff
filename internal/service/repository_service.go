package service

import (
	"fmt"

	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/repository"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/crypto"
	"github.com/xanzy/go-gitlab"
)

// RepositoryService handles repository business logic
type RepositoryService struct {
	repo              *repository.RepositoryRepo
	platformRepo      *repository.PlatformRepository
	systemConfigSvc   *SystemConfigService
	encryptor         *crypto.Encryptor
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(
	repo *repository.RepositoryRepo,
	platformRepo *repository.PlatformRepository,
	systemConfigSvc *SystemConfigService,
	cfg *config.Config,
) (*RepositoryService, error) {
	encryptor, err := crypto.NewEncryptor(cfg.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	return &RepositoryService{
		repo:            repo,
		platformRepo:    platformRepo,
		systemConfigSvc: systemConfigSvc,
		encryptor:       encryptor,
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
func (s *RepositoryService) ListFromGitLab(projectID uint, page, perPage int) ([]GitLabRepository, int, error) {
	// Get platform config
	platformConfig, err := s.platformRepo.GetConfig(projectID)
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
func (s *RepositoryService) List(projectID uint, page, pageSize int) ([]model.Repository, int64, error) {
	return s.repo.List(projectID, page, pageSize)
}

// Get retrieves a repository
func (s *RepositoryService) Get(id uint, projectID uint) (*model.Repository, error) {
	return s.repo.Get(id, projectID)
}

// BatchImportResult represents the result of batch import operation
type BatchImportResult struct {
	Succeeded []int64              `json:"succeeded"` // Successfully imported repository IDs
	Failed    []BatchImportFailure `json:"failed"`    // Failed imports
}

// BatchImportFailure represents a single failed import
type BatchImportFailure struct {
	RepositoryID int64  `json:"repository_id"`
	Error        string `json:"error"`
}

// BatchImport imports multiple repositories from GitLab
// Returns partial success - some repositories may succeed while others fail
func (s *RepositoryService) BatchImport(projectID uint, platformRepoIDs []int64, webhookCallbackURL string) error {
	// Get webhook URL from system config if not provided
	webhookCallbackURL, err := s.getWebhookURL(projectID, webhookCallbackURL)
	if err != nil {
		return err
	}

	// Create GitLab client
	git, err := s.createGitLabClient(projectID)
	if err != nil {
		return err
	}

	// Import repositories (partial success allowed)
	platformConfig, _ := s.platformRepo.GetConfig(projectID)
	for _, platformRepoID := range platformRepoIDs {
		// Import each repository independently
		// Errors are logged but don't stop the batch process
		_ = s.importSingleRepository(git, projectID, platformConfig.ID, platformRepoID, webhookCallbackURL)
	}

	return nil
}

// getWebhookURL retrieves webhook URL from parameter or system config
func (s *RepositoryService) getWebhookURL(projectID uint, webhookCallbackURL string) (string, error) {
	if webhookCallbackURL != "" {
		return webhookCallbackURL, nil
	}

	webhookConfig, err := s.systemConfigSvc.GetWebhookConfig(projectID)
	if err != nil {
		return "", fmt.Errorf("webhook URL not configured: %w", err)
	}

	return webhookConfig.WebhookCallbackURL, nil
}

// createGitLabClient creates an authenticated GitLab client
func (s *RepositoryService) createGitLabClient(projectID uint) (*gitlab.Client, error) {
	// Get platform config
	platformConfig, err := s.platformRepo.GetConfig(projectID)
	if err != nil {
		return nil, fmt.Errorf("platform not configured: %w", err)
	}

	// Decrypt token
	token, err := s.encryptor.Decrypt(platformConfig.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Create GitLab client
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(platformConfig.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return git, nil
}

// importSingleRepository imports a single repository with webhook creation
func (s *RepositoryService) importSingleRepository(
	git *gitlab.Client,
	projectID uint,
	platformID uint,
	platformRepoID int64,
	webhookCallbackURL string,
) error {
	// Step 1: Check if already imported
	if s.alreadyImported(projectID, platformID, platformRepoID) {
		return nil
	}

	// Step 2: Fetch GitLab project information
	project, err := s.fetchGitLabProject(git, platformRepoID)
	if err != nil {
		return err
	}

	// Step 3: Ensure webhook exists (find existing or create new)
	webhookID, webhookURL, err := s.ensureWebhook(git, platformRepoID, webhookCallbackURL)
	if err != nil {
		return err
	}

	// Step 4: Create repository record in database
	repoID, err := s.createRepositoryRecord(projectID, platformID, project, webhookID, webhookURL)
	if err != nil {
		return err
	}

	// Step 5: Test webhook immediately after import (async, don't block import)
	go func() {
		_ = s.TestWebhook(repoID, projectID)
	}()

	return nil
}

// alreadyImported checks if repository already exists in local database
func (s *RepositoryService) alreadyImported(projectID uint, platformID uint, platformRepoID int64) bool {
	existing, err := s.repo.GetByPlatformRepoID(projectID, platformID, platformRepoID)
	if err == nil && existing != nil {
		return true // Already imported
	}
	return false
}

// fetchGitLabProject retrieves project information from GitLab
func (s *RepositoryService) fetchGitLabProject(git *gitlab.Client, platformRepoID int64) (*gitlab.Project, error) {
	project, _, err := git.Projects.GetProject(int(platformRepoID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %d: %w", platformRepoID, err)
	}
	return project, nil
}

// ensureWebhook ensures webhook exists (finds existing or creates new)
func (s *RepositoryService) ensureWebhook(git *gitlab.Client, platformRepoID int64, callbackURL string) (int64, string, error) {
	// Try to find existing webhook first
	webhookID, webhookURL, err := s.findExistingWebhook(git, int(platformRepoID), callbackURL)
	if err != nil {
		return 0, "", fmt.Errorf("failed to check existing webhook: %w", err)
	}

	// If found, return it
	if webhookID != 0 {
		return webhookID, webhookURL, nil
	}

	// Otherwise, create new webhook
	webhookID, webhookURL, err = s.createWebhook(git, int(platformRepoID), callbackURL)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create webhook for project %d: %w", platformRepoID, err)
	}

	return webhookID, webhookURL, nil
}

// createRepositoryRecord creates repository record in database
func (s *RepositoryService) createRepositoryRecord(
	projectID uint,
	platformID uint,
	project *gitlab.Project,
	webhookID int64,
	webhookURL string,
) (uint, error) {
	repo := &model.Repository{
		ProjectID:      projectID,
		PlatformID:     platformID,
		PlatformRepoID: int64(project.ID),
		Name:           project.Name,
		FullPath:       project.PathWithNamespace,
		HTTPURL:        project.HTTPURLToRepo,
		SSHURL:         project.SSHURLToRepo,
		DefaultBranch:  project.DefaultBranch,
		WebhookID:      &webhookID,
		WebhookURL:     webhookURL,
		WebhookStatus:  model.WebhookStatusNotConfigured, // Will be tested immediately after creation
		IsActive:       true,
	}

	if err := s.repo.Create(repo); err != nil {
		return 0, fmt.Errorf("failed to create repository: %w", err)
	}

	return repo.ID, nil
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

// findExistingWebhook searches for an existing webhook with the same callback URL
// Returns webhook ID and URL if found, or (0, "", nil) if not found
func (s *RepositoryService) findExistingWebhook(git *gitlab.Client, projectID int, callbackURL string) (int64, string, error) {
	// List all webhooks for this project
	hooks, _, err := git.Projects.ListProjectHooks(projectID, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to list project hooks: %w", err)
	}

	// Search for webhook with matching URL
	for _, hook := range hooks {
		if hook.URL == callbackURL {
			return int64(hook.ID), hook.URL, nil
		}
	}

	// Not found
	return 0, "", nil
}

// UpdateLLMModel updates the LLM model for a repository
func (s *RepositoryService) UpdateLLMModel(id uint, llmModelID *uint) error {
	return s.repo.UpdateLLMModel(id, llmModelID)
}

// Delete deletes a repository and removes webhook from GitLab
func (s *RepositoryService) Delete(id uint, projectID uint) error {
	// Get repository
	repo, err := s.repo.Get(id, projectID)
	if err != nil {
		return fmt.Errorf("repository not found: %w", err)
	}

	// Get platform config
	platformConfig, err := s.platformRepo.GetConfig(projectID)
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

// TestWebhook tests if webhook exists on GitLab and updates status accordingly
func (s *RepositoryService) TestWebhook(id uint, projectID uint) error {
	// Step 1: Get repository and check local configuration
	repo, err := s.repo.Get(id, projectID)
	if err != nil {
		return fmt.Errorf("repository not found: %w", err)
	}

	if repo.WebhookID == nil {
		// Update status to not_configured
		_ = s.repo.SetWebhookStatus(id, model.WebhookStatusNotConfigured, "webhook ID is null")
		return fmt.Errorf("webhook not configured")
	}

	// Step 2: Create GitLab client
	platformConfig, err := s.platformRepo.GetConfig(projectID)
	if err != nil {
		return fmt.Errorf("platform not configured: %w", err)
	}

	token, err := s.encryptor.Decrypt(platformConfig.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(platformConfig.BaseURL))
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Step 3: Test webhook by fetching it from GitLab
	_, _, err = git.Projects.GetProjectHook(int(repo.PlatformRepoID), int(*repo.WebhookID))
	if err != nil {
		// Webhook not found on GitLab - update status to inactive
		if updateErr := s.repo.SetWebhookStatus(id, model.WebhookStatusInactive, err.Error()); updateErr != nil {
			// TODO: Use structured logger instead of fmt.Printf
			// logger.Error("Failed to update webhook status", "repository_id", id, "error", updateErr)
			fmt.Printf("[ERROR] Failed to update webhook status for repository %d: %v\n", id, updateErr)
		}
		return fmt.Errorf("webhook not found on GitLab: %w", err)
	}

	// Step 4: Update repository status to active
	if updateErr := s.repo.SetWebhookStatus(id, model.WebhookStatusActive, ""); updateErr != nil {
		// TODO: Use structured logger instead of fmt.Printf
		// logger.Error("Failed to update webhook status", "repository_id", id, "error", updateErr)
		fmt.Printf("[ERROR] Failed to update webhook status for repository %d: %v\n", id, updateErr)
	}

	return nil
}


// RecreateWebhook recreates webhook for a repository
func (s *RepositoryService) RecreateWebhook(id uint, projectID uint) error {
	// Get repository
	repo, err := s.repo.Get(id, projectID)
	if err != nil {
		return fmt.Errorf("repository not found: %w", err)
	}

	// Get webhook URL from system config
	webhookConfig, err := s.systemConfigSvc.GetWebhookConfig(projectID)
	if err != nil {
		return fmt.Errorf("webhook URL not configured: %w", err)
	}

	// Get platform config
	platformConfig, err := s.platformRepo.GetConfig(projectID)
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

	// Delete old webhook if exists
	if repo.WebhookID != nil {
		_, err = git.Projects.DeleteProjectHook(int(repo.PlatformRepoID), int(*repo.WebhookID))
		// Ignore error if webhook already deleted
	}

	// Create new webhook
	webhookID, webhookURL, err := s.createWebhook(git, int(repo.PlatformRepoID), webhookConfig.WebhookCallbackURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	// Update repository
	return s.repo.UpdateWebhook(id, webhookID, webhookURL)
}
