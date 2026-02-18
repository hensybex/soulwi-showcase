package usecase

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
)

type SubgroupsAndPromptsUsecase interface {
	ListSubGroups(ctx context.Context) ([]model.PromptSubGroup, error)
	ListPrompts(ctx context.Context) ([]model.Prompt, error)
}

type subgroupsAndPromptsUsecase struct {
	promptRepo repository.PromptRepository
}

func NewSubgroupsAndPromptsUsecase(promptRepo repository.PromptRepository) SubgroupsAndPromptsUsecase {
	return &subgroupsAndPromptsUsecase{
		promptRepo: promptRepo,
	}
}

func (u *subgroupsAndPromptsUsecase) ListSubGroups(ctx context.Context) ([]model.PromptSubGroup, error) {
	return u.promptRepo.ListSubGroups(ctx)
}

func (u *subgroupsAndPromptsUsecase) ListPrompts(ctx context.Context) ([]model.Prompt, error) {
	return u.promptRepo.List(ctx)
}
