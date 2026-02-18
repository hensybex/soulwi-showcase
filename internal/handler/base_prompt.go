// internal/handler/base_prompt_handler.go

package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type BasePromptHandler struct {
	basePromptUC usecase.BasePromptUsecase
}

func NewBasePromptHandler(pu usecase.BasePromptUsecase) *BasePromptHandler {
	return &BasePromptHandler{basePromptUC: pu}
}

// GET /base-prompts
func (h *BasePromptHandler) ListBasePrompts(c *gin.Context) {
	basePrompts, err := h.basePromptUC.ListBasePrompts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list base prompts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": basePrompts})
}

// GET /base-prompts/:id
func (h *BasePromptHandler) GetBasePrompt(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base prompt ID"})
		return
	}

	basePrompt, err := h.basePromptUC.GetBasePrompt(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Base prompt not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": basePrompt})
}

// POST /base-prompts
func (h *BasePromptHandler) CreateBasePrompt(c *gin.Context) {
	var bp model.BasePrompt
	if err := c.ShouldBindJSON(&bp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdBP, err := h.basePromptUC.CreateBasePrompt(c.Request.Context(), &bp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create base prompt"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": createdBP})
}

// PUT /base-prompts/:id
func (h *BasePromptHandler) UpdateBasePrompt(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base prompt ID"})
		return
	}

	var bp model.BasePrompt
	if err := c.ShouldBindJSON(&bp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bp.ID = uint(id)
	if err := h.basePromptUC.UpdateBasePrompt(c.Request.Context(), &bp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update base prompt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bp})
}

// DELETE /base-prompts/:id
func (h *BasePromptHandler) DeleteBasePrompt(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base prompt ID"})
		return
	}

	if err := h.basePromptUC.DeleteBasePrompt(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete base prompt"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
