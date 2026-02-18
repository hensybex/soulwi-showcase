// internal/repository/wise_phrase_like.go
package repository

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type WisePhraseLikeRepository interface {
	CreateLike(ctx context.Context, userUID string, phraseID uint) error
	IsAlreadyLiked(ctx context.Context, userUID string, phraseID uint) (bool, error)
	DeleteLike(ctx context.Context, userUID string, phraseID uint) error
	ListLikedPhrases(ctx context.Context, userUID string) ([]model.LikedPhraseResponse, error)
	DeleteAllByUserUID(ctx context.Context, userUID string) error // NEW
	ReassignUserUID(ctx context.Context, oldUID, newUID string) error
	DeleteByPhraseID(ctx context.Context, phraseID uint) error
}

type wisePhraseLikeRepo struct {
	db *gorm.DB
}

func NewWisePhraseLikeRepo(db *gorm.DB) WisePhraseLikeRepository {
	return &wisePhraseLikeRepo{db: db}
}

func (r *wisePhraseLikeRepo) CreateLike(ctx context.Context, userUID string, phraseID uint) error {
	rec := &model.WisePhraseLike{
		UserUID:      userUID,
		WisePhraseID: phraseID,
	}
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *wisePhraseLikeRepo) IsAlreadyLiked(ctx context.Context, userUID string, phraseID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.WisePhraseLike{}).
		Where("user_uid = ? AND wise_phrase_id = ?", userUID, phraseID).
		Count(&count).Error
	return count > 0, err
}

func (r *wisePhraseLikeRepo) DeleteLike(ctx context.Context, userUID string, phraseID uint) error {
	return r.db.WithContext(ctx).
		Where("user_uid = ? AND wise_phrase_id = ?", userUID, phraseID).
		Delete(&model.WisePhraseLike{}).Error
}

// For listing them, we can do a join
func (r *wisePhraseLikeRepo) ListLikedPhrases(ctx context.Context, userUID string) ([]model.LikedPhraseResponse, error) {
	var results []model.LikedPhraseResponse

	err := r.db.WithContext(ctx).
		Model(&model.WisePhraseLike{}).
		// Select fields from both tables, aliasing the like's timestamp to 'liked_at'
		Select("wise_phrases.id, wise_phrases.created_at, wise_phrases.text, wise_phrases.like_count, wise_phrase_likes.created_at as liked_at").
		Joins("JOIN wise_phrases ON wise_phrases.id = wise_phrase_likes.wise_phrase_id").
		Where("wise_phrase_likes.user_uid = ?", userUID).
		// Order by the like timestamp in ascending order (oldest likes first)
		Order("wise_phrase_likes.created_at ASC").
		Scan(&results).Error

	return results, err
}

func (r *wisePhraseLikeRepo) DeleteAllByUserUID(ctx context.Context, userUID string) error {
	return r.db.WithContext(ctx).
		Where("user_uid = ?", userUID).
		Delete(&model.WisePhraseLike{}).Error
}

func (r *wisePhraseLikeRepo) DeleteByPhraseID(ctx context.Context, phraseID uint) error {
	return r.db.WithContext(ctx).
		Where("wise_phrase_id = ?", phraseID).
		Delete(&model.WisePhraseLike{}).Error
}

func (r *wisePhraseLikeRepo) ReassignUserUID(ctx context.Context, oldUID, newUID string) error {
	if oldUID == "" || oldUID == newUID {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.WisePhraseLike{}).
		Where("user_uid = ?", oldUID).
		Update("user_uid", newUID).Error
}
