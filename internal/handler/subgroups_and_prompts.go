package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type SubgroupsAndPromptsHandler struct {
	usecase usecase.SubgroupsAndPromptsUsecase
}

func NewSubgroupsAndPromptsHandler(uc usecase.SubgroupsAndPromptsUsecase) *SubgroupsAndPromptsHandler {
	return &SubgroupsAndPromptsHandler{usecase: uc}
}

func (h *SubgroupsAndPromptsHandler) GetAllSubgroupsAndPrompts(c *gin.Context) {
	subgroups, err := h.usecase.ListSubGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	prompts, err := h.usecase.ListPrompts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"subgroups": subgroups,
		"prompts":   prompts,
	})
}
