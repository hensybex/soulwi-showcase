// internal/handler/dashboard_handler.go

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type DashboardHandler struct {
	dashboardUsecase usecase.DashboardUsecase
}

func NewDashboardHandler(dashboardUsecase usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{dashboardUsecase}
}

func (h *DashboardHandler) GetDashboardData(c *gin.Context) {
	dashboardData, err := h.dashboardUsecase.GetDashboardData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboardData)
}

func (h *DashboardHandler) GetInfographicData(c *gin.Context) {
	infographicData, err := h.dashboardUsecase.GetInfographicData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, infographicData)
}
