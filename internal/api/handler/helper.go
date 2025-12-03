package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// TODO: This is a temporary solution for multi-project architecture.
// In the future, this should be replaced with ProjectContext middleware
// that extracts project_id from user preferences and sets it in gin.Context.
//
// See: MULTI_PROJECT_IMPLEMENTATION_PLAN.md for complete implementation guide.
//
// getUserDefaultProjectID gets the user's default/first project ID.
// This is a temporary workaround until ProjectContext middleware is implemented.
func getUserDefaultProjectID(c *gin.Context, db *gorm.DB) (uint, error) {
	// Get user ID from context (set by Auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, gorm.ErrRecordNotFound
	}

	// Try to get active project from user preferences
	var pref model.UserProjectPreference
	if err := db.Where("user_id = ?", userID).First(&pref).Error; err == nil {
		return pref.ProjectID, nil
	}

	// If no preference set, get user's first project
	var project model.Project
	if err := db.Where("user_id = ?", userID).First(&project).Error; err != nil {
		return 0, err
	}

	return project.ID, nil
}
