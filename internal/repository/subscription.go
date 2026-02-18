// internal/repository/subscription_repository.go
package repository

import (
	"context"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *model.Subscription) error
	GetByOriginalTransactionID(ctx context.Context, originalTransactionID string) (*model.Subscription, error)
	GetByUserID(ctx context.Context, userID uint) (*model.Subscription, error)
	GetActiveByUserID(ctx context.Context, userID uint) (*model.Subscription, error)
	Update(ctx context.Context, subscription *model.Subscription) error
	Delete(ctx context.Context, id uint) error
	GetExpiredSubscriptions(ctx context.Context, beforeTime time.Time) ([]model.Subscription, error)
	MarkAllAsDeletedByUserID(ctx context.Context, userID uint, now time.Time) error
	ReassignUserID(ctx context.Context, oldUserID, newUserID uint) error
}

type subscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepo(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, subscription *model.Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *subscriptionRepo) GetByOriginalTransactionID(ctx context.Context, originalTransactionID string) (*model.Subscription, error) {
	var subscription model.Subscription
	err := r.db.WithContext(ctx).
		Where("original_transaction_id = ?", originalTransactionID).
		Preload("User").
		First(&subscription).Error

	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepo) GetByUserID(ctx context.Context, userID uint) (*model.Subscription, error) {
	var subscription model.Subscription
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("User").
		Order("created_at DESC").
		First(&subscription).Error

	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepo) GetActiveByUserID(ctx context.Context, userID uint) (*model.Subscription, error) {
	var subscription model.Subscription
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status IN ? AND expires_at > ?",
			userID,
			[]string{"active", "trial", "grace_period"},
			time.Now()).
		Preload("User").
		Order("created_at DESC").
		First(&subscription).Error

	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepo) Update(ctx context.Context, subscription *model.Subscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

func (r *subscriptionRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Subscription{}, id).Error
}

func (r *subscriptionRepo) GetExpiredSubscriptions(ctx context.Context, beforeTime time.Time) ([]model.Subscription, error) {
	var subscriptions []model.Subscription
	err := r.db.WithContext(ctx).
		Where("expires_at < ? AND status IN ?", beforeTime, []string{"active", "trial", "grace_period"}).
		Preload("User").
		Find(&subscriptions).Error

	return subscriptions, err
}

func (r *subscriptionRepo) MarkAllAsDeletedByUserID(ctx context.Context, userID uint, now time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.Subscription{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"status":     "deleted_user",
			"expires_at": now,
		}).Error
}

func (r *subscriptionRepo) ReassignUserID(ctx context.Context, oldUserID, newUserID uint) error {
	if oldUserID == 0 || oldUserID == newUserID {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Subscription{}).
		Where("user_id = ?", oldUserID).
		Update("user_id", newUserID).Error
}
