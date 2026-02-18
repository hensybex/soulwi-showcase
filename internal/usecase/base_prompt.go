// internal/usecase/base_prompt.go

package usecase

import (
	"context"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
)

type BasePromptUsecase interface {

	// BasePrompt methods
	ListBasePrompts(ctx context.Context) ([]model.BasePrompt, error)
	GetBasePrompt(ctx context.Context, id uint) (*model.BasePrompt, error)
	CreateBasePrompt(ctx context.Context, bp *model.BasePrompt) (*model.BasePrompt, error)
	UpdateBasePrompt(ctx context.Context, bp *model.BasePrompt) error
	DeleteBasePrompt(ctx context.Context, id uint) error
}

type basePromptUsecase struct {
	basePromptRepo repository.BasePromptRepository
}

func NewBasePromptUsecase(pr repository.BasePromptRepository) BasePromptUsecase {
	return &basePromptUsecase{basePromptRepo: pr}
}

func (u *basePromptUsecase) ListBasePrompts(ctx context.Context) ([]model.BasePrompt, error) {
	return u.basePromptRepo.ListBasePrompts(ctx)
}

func (u *basePromptUsecase) GetBasePrompt(ctx context.Context, id uint) (*model.BasePrompt, error) {
	return u.basePromptRepo.GetBasePromptByID(ctx, id)
}

func (u *basePromptUsecase) CreateBasePrompt(ctx context.Context, bp *model.BasePrompt) (*model.BasePrompt, error) {
	if err := u.basePromptRepo.CreateBasePrompt(ctx, bp); err != nil {
		return nil, err
	}
	return bp, nil
}

func (u *basePromptUsecase) UpdateBasePrompt(ctx context.Context, bp *model.BasePrompt) error {
	return u.basePromptRepo.UpdateBasePrompt(ctx, bp)
}

func (u *basePromptUsecase) DeleteBasePrompt(ctx context.Context, id uint) error {
	return u.basePromptRepo.DeleteBasePrompt(ctx, id)
}
