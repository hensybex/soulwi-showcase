// internal/transport/router/groups/prompt.go

package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
)

func ConfigurePromptRoutes(r *gin.Engine, c *di.Container, authMiddleware gin.HandlerFunc) {
	promptGroup := r.Group("/prompts")

	// --- Define PUBLIC prompt routes FIRST ---
	// These routes do not require any authentication or role checks.
	// Moved from below and removed middleware.RequireRoleMiddleware.

	// Public: Example version endpoint
	promptGroup.GET("/version", c.PromptVersionHandler.GetPromptVersion) // No middleware

	// Public: Example fetch all subgroups + prompts
	promptGroup.GET("/all-details", c.SubgroupsAndPromptsHandler.GetAllSubgroupsAndPrompts) // No middleware

	// --- Apply AUTHENTICATION middleware to the rest of the group ---
	// All routes defined below this line will require authentication (JWT or Firebase).
	promptGroup.Use(authMiddleware)

	// --- Authenticated GET endpoints (require user or admin role) ---
	promptGroup.GET("/main-groups", middleware.RequireRoleMiddleware("admin", "user"), c.PromptHandler.ListMainGroups)
	promptGroup.GET("/main-groups/:id/sub-groups", middleware.RequireRoleMiddleware("admin", "user"), c.PromptHandler.ListSubGroupsByMainGroup)
	promptGroup.GET("/sub-groups/:id/prompts", middleware.RequireRoleMiddleware("admin", "user"), c.PromptHandler.GetPromptsBySubGroup)
	promptGroup.GET("/base", middleware.RequireRoleMiddleware("admin", "user"), c.PromptHandler.GetBasePrompts)

	// Base prompt GET routes (user or admin)
	promptGroup.GET("/base-prompts", middleware.RequireRoleMiddleware("admin", "user"), c.BasePromptHandler.ListBasePrompts)
	promptGroup.GET("/base-prompts/:id", middleware.RequireRoleMiddleware("admin", "user"), c.BasePromptHandler.GetBasePrompt)

	// --- Authenticated Non-GET endpoints (admin only) ---
	promptGroup.POST("/sub-groups", middleware.RequireRoleMiddleware("admin"), c.PromptHandler.CreateSubGroup)
	promptGroup.POST("/", middleware.RequireRoleMiddleware("admin"), c.PromptHandler.CreatePrompt)
	promptGroup.PUT("/:id", middleware.RequireRoleMiddleware("admin"), c.PromptHandler.UpdatePrompt)
	promptGroup.DELETE("/sub-groups/:id", middleware.RequireRoleMiddleware("admin"), c.PromptHandler.DeleteSubGroup)
	promptGroup.DELETE("/:id", middleware.RequireRoleMiddleware("admin"), c.PromptHandler.DeletePrompt)

	// Base prompt admin routes
	promptGroup.POST("/base-prompts", middleware.RequireRoleMiddleware("admin"), c.BasePromptHandler.CreateBasePrompt)
	promptGroup.PUT("/base-prompts/:id", middleware.RequireRoleMiddleware("admin"), c.BasePromptHandler.UpdateBasePrompt)
	promptGroup.DELETE("/base-prompts/:id", middleware.RequireRoleMiddleware("admin"), c.BasePromptHandler.DeleteBasePrompt)

	// Put base prompt for sub‐groups (admin only)
	promptGroup.PUT("/sub-groups/:subGroupId/base-prompt", middleware.RequireRoleMiddleware("admin"), c.PromptHandler.UpdateBasePrompt)
	promptGroup.PUT("/sub-groups/base-prompt", middleware.RequireRoleMiddleware("admin"), c.PromptHandler.UpdateAllBasePrompts)

	// NOTE: The original GET /version and GET /all-details routes were moved above promptGroup.Use(authMiddleware)
}
