// internal/repository/feedback.go

package repository

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type FeedbackRepository interface {
	Create(ctx context.Context, feedback *model.Feedback) error
}

type feedbackRepo struct {
	db *gorm.DB
}

func NewFeedbackRepo(db *gorm.DB) FeedbackRepository {
	return &feedbackRepo{db: db}
}

func (r *feedbackRepo) Create(ctx context.Context, feedback *model.Feedback) error {
	return r.db.WithContext(ctx).Create(feedback).Error
}
