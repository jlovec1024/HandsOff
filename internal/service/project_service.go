package service

import (
	"context"
	"fmt"

	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/repository"
	"github.com/handsoff/handsoff/pkg/logger"
	"gorm.io/gorm"
)

// ProjectService handles project business logic
type ProjectService struct {
	projectRepo *repository.ProjectRepository
	prefRepo    *repository.UserPreferenceRepository
	log         *logger.Logger
}

// NewProjectService creates a new project service
func NewProjectService(
	projectRepo *repository.ProjectRepository,
	prefRepo *repository.UserPreferenceRepository,
	log *logger.Logger,
) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		prefRepo:    prefRepo,
		log:         log,
	}
}

// CreateProject creates a new project for a user
func (s *ProjectService) CreateProject(ctx context.Context, userID uint, name, description string) (*model.Project, error) {
	// Validate project name
	if name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	// Check if project name already exists for this user
	existing, err := s.projectRepo.FindByUserIDAndName(ctx, userID, name)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing project: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("project name '%s' already exists", name)
	}

	// Create project
	project := &model.Project{
		Name:        name,
		Description: description,
		UserID:      userID,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	s.log.Info("Project created",
		"project_id", project.ID,
		"user_id", userID,
		"name", name)

	return project, nil
}

// GetUserProjects retrieves all projects for a user
func (s *ProjectService) GetUserProjects(ctx context.Context, userID uint) ([]model.Project, error) {
	projects, err := s.projectRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}
	return projects, nil
}

// GetProject retrieves a specific project with ownership validation
func (s *ProjectService) GetProject(ctx context.Context, id uint, userID uint) (*model.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Validate ownership
	if project.UserID != userID {
		return nil, fmt.Errorf("access denied: you don't own this project")
	}

	return project, nil
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(ctx context.Context, id uint, userID uint, name, description string) error {
	// Get existing project with ownership validation
	project, err := s.GetProject(ctx, id, userID)
	if err != nil {
		return err
	}

	// Check if new name conflicts with another project
	if name != "" && name != project.Name {
		existing, err := s.projectRepo.FindByUserIDAndName(ctx, userID, name)
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing project: %w", err)
		}
		if existing != nil && existing.ID != id {
			return fmt.Errorf("project name '%s' already exists", name)
		}
		project.Name = name
	}

	if description != "" {
		project.Description = description
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	s.log.Info("Project updated",
		"project_id", id,
		"user_id", userID)

	return nil
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(ctx context.Context, id uint, userID uint) error {
	// Get existing project with ownership validation
	project, err := s.GetProject(ctx, id, userID)
	if err != nil {
		return err
	}

	// Check if this is the user's only project
	count, err := s.projectRepo.Count(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to count projects: %w", err)
	}
	if count <= 1 {
		return fmt.Errorf("cannot delete your only project. Create another project first")
	}

	// Check if this is the active project
	activeProject, err := s.prefRepo.GetActiveProject(ctx, userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to get active project: %w", err)
	}

	if activeProject != nil && activeProject.ID == id {
		return fmt.Errorf("cannot delete the active project. Switch to another project first")
	}

	// Delete project
	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	s.log.Info("Project deleted",
		"project_id", id,
		"user_id", userID,
		"name", project.Name)

	return nil
}

// SwitchActiveProject sets a project as the active project for a user
func (s *ProjectService) SwitchActiveProject(ctx context.Context, userID uint, projectID uint) error {
	// Validate project ownership
	project, err := s.GetProject(ctx, projectID, userID)
	if err != nil {
		return err
	}

	// Set as active project
	if err := s.prefRepo.SetActiveProject(ctx, userID, projectID); err != nil {
		return fmt.Errorf("failed to set active project: %w", err)
	}

	s.log.Info("Active project switched",
		"user_id", userID,
		"project_id", projectID,
		"name", project.Name)

	return nil
}

// GetActiveProject retrieves the active project for a user
func (s *ProjectService) GetActiveProject(ctx context.Context, userID uint) (*model.Project, error) {
	project, err := s.prefRepo.GetActiveProject(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active project selected. Please select a project")
		}
		return nil, fmt.Errorf("failed to get active project: %w", err)
	}

	// Double-check ownership
	if project.UserID != userID {
		return nil, fmt.Errorf("active project ownership mismatch")
	}

	return project, nil
}

// GetActiveProjectID retrieves only the active project ID for a user (performance optimized)
func (s *ProjectService) GetActiveProjectID(ctx context.Context, userID uint) (uint, error) {
	projectID, err := s.prefRepo.GetActiveProjectID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("no active project selected")
		}
		return 0, fmt.Errorf("failed to get active project ID: %w", err)
	}
	return projectID, nil
}
