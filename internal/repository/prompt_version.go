package repository

import (
	"context"
	"fmt"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

// PromptVersionRepo handles reading/writing the single global prompt version.
type PromptVersionRepo interface {
	GetVersion(ctx context.Context) (int64, error)
	IncrementVersion(ctx context.Context) error
}

type promptVersionRepo struct {
	db *gorm.DB
}

func NewPromptVersionRepo(db *gorm.DB) PromptVersionRepo {
	return &promptVersionRepo{db: db}
}

func (r *promptVersionRepo) GetVersion(ctx context.Context) (int64, error) {
	var pv model.PromptVersion
	if err := r.db.WithContext(ctx).First(&pv, 1).Error; err != nil {
		return 0, fmt.Errorf("GetVersion failed: %w", err)
	}
	return pv.Version, nil
}

func (r *promptVersionRepo) IncrementVersion(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Model(&model.PromptVersion{}).
		Where("id = ?", 1).
		UpdateColumn("version", gorm.Expr("version + ?", 1))
	if result.Error != nil {
		return fmt.Errorf("IncrementVersion failed: %w", result.Error)
	}
	return nil
}
