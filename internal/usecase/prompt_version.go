package usecase

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/repository"
)

type PromptVersionUsecase interface {
	GetVersion(ctx context.Context) (int64, error)
	IncrementVersion(ctx context.Context) error
}

type promptVersionUsecase struct {
	repo repository.PromptVersionRepo
}

func NewPromptVersionUsecase(repo repository.PromptVersionRepo) PromptVersionUsecase {
	return &promptVersionUsecase{repo: repo}
}

func (u *promptVersionUsecase) GetVersion(ctx context.Context) (int64, error) {
	return u.repo.GetVersion(ctx)
}

func (u *promptVersionUsecase) IncrementVersion(ctx context.Context) error {
	return u.repo.IncrementVersion(ctx)
}
