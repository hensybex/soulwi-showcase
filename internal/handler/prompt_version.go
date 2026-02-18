package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type PromptVersionHandler struct {
	versionUC usecase.PromptVersionUsecase
}

func NewPromptVersionHandler(uc usecase.PromptVersionUsecase) *PromptVersionHandler {
	return &PromptVersionHandler{versionUC: uc}
}

// GetPromptVersion returns the current version number as JSON.
func (h *PromptVersionHandler) GetPromptVersion(c *gin.Context) {
	version, err := h.versionUC.GetVersion(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"version": version})
}
