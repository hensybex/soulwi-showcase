// internal/repository/base_prompt.go

package repository

import (
	"context"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type BasePromptRepository interface {
	// Existing methods...

	// BasePrompt methods
	ListBasePrompts(ctx context.Context) ([]model.BasePrompt, error)
	GetBasePromptByID(ctx context.Context, id uint) (*model.BasePrompt, error)
	CreateBasePrompt(ctx context.Context, bp *model.BasePrompt) error
	UpdateBasePrompt(ctx context.Context, bp *model.BasePrompt) error
	DeleteBasePrompt(ctx context.Context, id uint) error
}

// Implement the new methods in basePromptRepo

type basePromptRepo struct {
	db *gorm.DB
}

func NewBasePromptRepo(db *gorm.DB) BasePromptRepository {
	return &basePromptRepo{db: db}
}

func (r *basePromptRepo) ListBasePrompts(ctx context.Context) ([]model.BasePrompt, error) {
	var basePrompts []model.BasePrompt
	if err := r.db.WithContext(ctx).Find(&basePrompts).Error; err != nil {
		return nil, err
	}
	return basePrompts, nil
}

func (r *basePromptRepo) GetBasePromptByID(ctx context.Context, id uint) (*model.BasePrompt, error) {
	var basePrompt model.BasePrompt
	if err := r.db.WithContext(ctx).First(&basePrompt, id).Error; err != nil {
		return nil, err
	}
	return &basePrompt, nil
}

func (r *basePromptRepo) CreateBasePrompt(ctx context.Context, bp *model.BasePrompt) error {
	return r.db.WithContext(ctx).Create(bp).Error
}

func (r *basePromptRepo) UpdateBasePrompt(ctx context.Context, bp *model.BasePrompt) error {
	return r.db.WithContext(ctx).Save(bp).Error
}

func (r *basePromptRepo) DeleteBasePrompt(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.BasePrompt{}, id).Error
}
