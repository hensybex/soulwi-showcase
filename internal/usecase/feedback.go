// internal/usecase/feedback.go

package usecase

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
)

type FeedbackUsecase interface {
	CreateFeedback(ctx context.Context, userID string, text string) error
}

type feedbackUsecase struct {
	repo repository.FeedbackRepository
}

func NewFeedbackUsecase(repo repository.FeedbackRepository) FeedbackUsecase {
	return &feedbackUsecase{repo: repo}
}

func (uc *feedbackUsecase) CreateFeedback(ctx context.Context, userID string, text string) error {
	fb := &model.Feedback{
		UserID: userID,
		Text:   text,
	}
	return uc.repo.Create(ctx, fb)
}
