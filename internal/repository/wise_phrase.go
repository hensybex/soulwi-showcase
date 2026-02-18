// internal/repository/wise_phrase.go
package repository

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

// Provide an interface for mocking, etc.
type WisePhraseRepository interface {
	CreateBatch(ctx context.Context, phrases []model.WisePhrase) error
	ListAll(ctx context.Context) ([]model.WisePhrase, error)
	ListPaged(ctx context.Context, limit, offset int) ([]model.WisePhrase, error)
	Count(ctx context.Context) (int64, error)
	IncrementLikeCount(ctx context.Context, phraseID uint) error
	DecrementLikeCount(ctx context.Context, phraseID uint) error
	IncrementShareCount(ctx context.Context, phraseID uint) error
	GetByID(ctx context.Context, id uint) (*model.WisePhrase, error)
	DeleteByID(ctx context.Context, id uint) error
}

type wisePhraseRepo struct {
	db *gorm.DB
}

func NewWisePhraseRepo(db *gorm.DB) WisePhraseRepository {
	return &wisePhraseRepo{db: db}
}

func (r *wisePhraseRepo) CreateBatch(ctx context.Context, phrases []model.WisePhrase) error {
	return r.db.WithContext(ctx).Create(&phrases).Error
}

func (r *wisePhraseRepo) ListAll(ctx context.Context) ([]model.WisePhrase, error) {
	var list []model.WisePhrase
	err := r.db.WithContext(ctx).Order("id DESC").Find(&list).Error
	return list, err
}

func (r *wisePhraseRepo) ListPaged(ctx context.Context, limit, offset int) ([]model.WisePhrase, error) {
	var list []model.WisePhrase
	query := r.db.WithContext(ctx).Order("id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *wisePhraseRepo) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.WisePhrase{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *wisePhraseRepo) IncrementLikeCount(ctx context.Context, phraseID uint) error {
	// You could also do a raw update for efficiency
	return r.db.WithContext(ctx).Model(&model.WisePhrase{}).
		Where("id = ?", phraseID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).
		Error
}

func (r *wisePhraseRepo) DecrementLikeCount(ctx context.Context, phraseID uint) error {
	return r.db.WithContext(ctx).Model(&model.WisePhrase{}).
		Where("id = ?", phraseID).
		UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).
		Error
}

func (r *wisePhraseRepo) GetByID(ctx context.Context, id uint) (*model.WisePhrase, error) {
	var wp model.WisePhrase
	if err := r.db.WithContext(ctx).First(&wp, id).Error; err != nil {
		return nil, err
	}
	return &wp, nil
}

func (r *wisePhraseRepo) IncrementShareCount(ctx context.Context, phraseID uint) error {
	return r.db.WithContext(ctx).Model(&model.WisePhrase{}).
		Where("id = ?", phraseID).
		UpdateColumn("share_count", gorm.Expr("share_count + ?", 1)).
		Error
}

func (r *wisePhraseRepo) DeleteByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.WisePhrase{}, id).Error
}
