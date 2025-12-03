package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/model"
	"gorm.io/gorm"
)

// ProjectContext middleware extracts user's active project ID and sets it in context.
// This replaces the temporary getUserDefaultProjectID() function calls in handlers.
//
// Usage in router:
//   protected.Use(middleware.ProjectContext(db))
//
// Usage in handler:
//   projectID := c.GetUint("project_id")
//
func ProjectContext(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by Auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Try to get active project from user preferences
		var pref model.UserProjectPreference
		if err := db.Where("user_id = ?", userID).First(&pref).Error; err == nil {
			c.Set("project_id", pref.ProjectID)
			c.Next()
			return
		}

		// If no preference set, get user's first project
		var project model.Project
		if err := db.Where("user_id = ?", userID).First(&project).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No project found. Please create a project first.",
			})
			c.Abort()
			return
		}

		c.Set("project_id", project.ID)
		c.Next()
	}
}
