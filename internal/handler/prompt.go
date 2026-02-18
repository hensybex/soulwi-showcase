// internal/handler/prompt_handler.go

package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type PromptHandler struct {
	promptUC usecase.PromptUsecase
}

func NewPromptHandler(pu usecase.PromptUsecase) *PromptHandler {
	return &PromptHandler{promptUC: pu}
}

// GET /prompts
func (h *PromptHandler) ListPrompts(c *gin.Context) {
	prompts, err := h.promptUC.ListPrompts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list prompts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": prompts})
}

// GET /prompts/:id
func (h *PromptHandler) GetPrompt(c *gin.Context) {
	role := c.GetString("role") // Extract the role from the context
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompt ID"})
		return
	}

	var prompt interface{}
	if role == "admin" {
		// Admin: Full prompt details
		prompt, err = h.promptUC.GetPrompt(c.Request.Context(), uint(idUint64))
	} else if role == "user" {
		// User: Limited prompt details
		//prompt, err = h.promptUC.GetUserPrompt(c.Request.Context(), uint(idUint64))
		// COMMENTED! LIMIT DETAILS LATER!
		prompt, err = h.promptUC.GetPrompt(c.Request.Context(), uint(idUint64))
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prompt"})
		return
	}
	if prompt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": prompt})
}

// POST /prompts
func (h *PromptHandler) CreatePrompt(c *gin.Context) {
	var p model.Prompt
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Delegate to usecase for creation and validation logic
	createdPrompt, err := h.promptUC.CreatePrompt(c.Request.Context(), &p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": createdPrompt})
}

// PUT /prompts/:id
func (h *PromptHandler) UpdatePrompt(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompt ID"})
		return
	}

	var p model.Prompt
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	p.ID = uint(idUint64)

	if err := h.promptUC.UpdatePrompt(c.Request.Context(), &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prompt"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

// DELETE /prompts/:id
func (h *PromptHandler) DeletePrompt(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompt ID"})
		return
	}
	if err := h.promptUC.DeletePrompt(c.Request.Context(), uint(idUint64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete prompt"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Prompt deleted"})
}

// PATCH /prompts/:id/group
func (h *PromptHandler) UpdateSubGroup(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompt ID"})
		return
	}

	var payload struct {
		MainGroupID *uint `json:"main_group_id"` // Use IDs instead of names
		SubGroupID  *uint `json:"sub_group_id"`  // Use IDs instead of names
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Delegate to usecase for updating logic
	if err := h.promptUC.UpdateGroups(c.Request.Context(), uint(idUint64), payload.MainGroupID, payload.SubGroupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group information updated"})
}

// POST /prompt-main-groups
func (h *PromptHandler) CreateMainGroup(c *gin.Context) {
	var mg model.PromptMainGroup
	if err := c.ShouldBindJSON(&mg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	createdMainGroup, err := h.promptUC.CreateMainGroup(c.Request.Context(), &mg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": createdMainGroup})
}

// POST /prompt-sub-groups
func (h *PromptHandler) CreateSubGroup(c *gin.Context) {
	var sg model.PromptSubGroup
	if err := c.ShouldBindJSON(&sg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	createdSubGroup, err := h.promptUC.CreateSubGroup(c.Request.Context(), &sg)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			// This is a duplicate name error
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": createdSubGroup})
}

// GET /main-groups
func (h *PromptHandler) ListMainGroups(c *gin.Context) {
	mainGroups, err := h.promptUC.ListMainGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list main groups"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mainGroups})
}

// GET /sub-groups
func (h *PromptHandler) ListSubGroups(c *gin.Context) {
	subGroups, err := h.promptUC.ListSubGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list subgroups"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": subGroups})
}

// GET /main-groups/:id/sub-groups
func (h *PromptHandler) ListSubGroupsByMainGroup(c *gin.Context) {
	log.Println("IN ListSubGroupsByMainGroup")
	role := c.GetString("role") // Extract the role from the context
	log.Println(role)
	idStr := c.Param("id")
	log.Println(idStr)
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid main group ID"})
		return
	}

	var subGroups interface{}
	if role == "admin" {
		// Admin: Full subgroup details
		subGroups, err = h.promptUC.ListSubGroupsByMainGroup(c.Request.Context(), uint(idUint64))
	} else if role == "user" {
		// User: Limited subgroup details
		//subGroups, err = h.promptUC.ListUserSubGroupsByMainGroup(c.Request.Context(), uint(idUint64))
		// COMMENTED! LIMIT DETAILS LATER!
		subGroups, err = h.promptUC.ListSubGroupsByMainGroup(c.Request.Context(), uint(idUint64))
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list subgroups for the main group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subGroups})
}

// DELETE /sub-groups/:id
func (h *PromptHandler) DeleteSubGroup(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sub-group ID"})
		return
	}

	// Delegate to usecase for deletion logic
	if err := h.promptUC.DeleteSubGroup(c.Request.Context(), uint(idUint64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sub-group and related prompts deleted successfully"})
}

// GET /sub-groups/:id/prompts
func (h *PromptHandler) GetPromptsBySubGroup(c *gin.Context) {
	role := c.GetString("role") // Extract the role from the context
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sub-group ID"})
		return
	}

	var prompts interface{}
	if role == "admin" {
		// Admin: Full prompt details
		prompts, err = h.promptUC.GetPromptsBySubGroup(c.Request.Context(), uint(idUint64))
	} else if role == "user" {
		// User: Limited prompt details
		//prompts, err = h.promptUC.GetUserPromptsBySubGroup(c.Request.Context(), uint(idUint64))
		// COMMENTED! LIMIT DETAILS LATER!
		prompts, err = h.promptUC.GetPromptsBySubGroup(c.Request.Context(), uint(idUint64))
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": prompts})
}

// GET /prompts/base
func (h *PromptHandler) GetBasePrompts(c *gin.Context) {
	prompts, err := h.promptUC.GetBasePrompts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch base prompts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": prompts})
}

// PUT /sub-groups/:id/base-prompt
func (h *PromptHandler) UpdateBasePrompt(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sub-group ID"})
		return
	}

	var payload struct {
		BasePromptID uint `json:"base_prompt_id"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.promptUC.UpdateSubGroupBasePrompt(c.Request.Context(), uint(idUint64), payload.BasePromptID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Base prompt updated successfully"})
}

func (h *PromptHandler) UpdateAllBasePrompts(c *gin.Context) {
	var payload struct {
		BasePromptID uint `json:"base_prompt_id"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.promptUC.UpdateAllSubGroupsBasePrompt(c.Request.Context(), payload.BasePromptID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Base prompt updated for all sub-groups"})
}
