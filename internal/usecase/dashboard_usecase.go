// internal/usecase/dashboard_usecase.go

package usecase

import (
	"github.com/hensybex/soulwi_go_back/internal/dto"
	"github.com/hensybex/soulwi_go_back/internal/repository"
)

type DashboardUsecase interface {
	GetDashboardData() (*dto.DashboardDataDTO, error)
	GetInfographicData() (*dto.InfographicDataDTO, error)
}

type dashboardUsecase struct {
	dashboardRepo repository.DashboardRepository
}

func NewDashboardUsecase(dashboardRepo repository.DashboardRepository) DashboardUsecase {
	return &dashboardUsecase{dashboardRepo}
}

func (uc *dashboardUsecase) GetDashboardData() (*dto.DashboardDataDTO, error) {
	feedbacks, err := uc.dashboardRepo.GetFeedbacks(100) // Limit to 100 for now
	if err != nil {
		return nil, err
	}

	promptUsage, err := uc.dashboardRepo.GetPromptUsageStats()
	if err != nil {
		return nil, err
	}

	wisePhraseStats, err := uc.dashboardRepo.GetWisePhraseStats()
	if err != nil {
		return nil, err
	}

	dashboardData := &dto.DashboardDataDTO{
		Feedbacks:       feedbacks,
		PromptUsage:     promptUsage,
		WisePhraseStats: wisePhraseStats,
	}

	return dashboardData, nil
}

func (uc *dashboardUsecase) GetInfographicData() (*dto.InfographicDataDTO, error) {
	return uc.dashboardRepo.GetInfographicData()
}
