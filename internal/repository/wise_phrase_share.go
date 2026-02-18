package repository

import (
	"context"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type WisePhraseShareRepository interface {
	CreateShare(ctx context.Context, userUID string, phraseID uint) error
	DeleteAllByUserUID(ctx context.Context, userUID string) error // NEW
	ReassignUserUID(ctx context.Context, oldUID, newUID string) error
	DeleteByPhraseID(ctx context.Context, phraseID uint) error
}

type wisePhraseShareRepo struct {
	db *gorm.DB
}

func NewWisePhraseShareRepo(db *gorm.DB) WisePhraseShareRepository {
	return &wisePhraseShareRepo{db: db}
}

func (r *wisePhraseShareRepo) CreateShare(ctx context.Context, userUID string, phraseID uint) error {
	rec := &model.WisePhraseShare{
		UserUID:      userUID,
		WisePhraseID: phraseID,
	}
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *wisePhraseShareRepo) DeleteAllByUserUID(ctx context.Context, userUID string) error {
	return r.db.WithContext(ctx).
		Where("user_uid = ?", userUID).
		Delete(&model.WisePhraseShare{}).Error
}

func (r *wisePhraseShareRepo) DeleteByPhraseID(ctx context.Context, phraseID uint) error {
	return r.db.WithContext(ctx).
		Where("wise_phrase_id = ?", phraseID).
		Delete(&model.WisePhraseShare{}).Error
}

func (r *wisePhraseShareRepo) ReassignUserUID(ctx context.Context, oldUID, newUID string) error {
	if oldUID == "" || oldUID == newUID {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.WisePhraseShare{}).
		Where("user_uid = ?", oldUID).
		Update("user_uid", newUID).Error
}
