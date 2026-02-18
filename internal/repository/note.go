// File: internal/repository/note.go

package repository

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type NoteRepository interface {
	Create(ctx context.Context, note *model.Note) error
	GetByID(ctx context.Context, id uint, userUID string) (*model.Note, error)
	Update(ctx context.Context, note *model.Note) error
	Delete(ctx context.Context, id uint, userUID string) error
	ListByUser(ctx context.Context, userUID string) ([]model.Note, error)
	SearchByUser(ctx context.Context, userUID, query string) ([]model.Note, error)
	DeleteAllByUserUID(ctx context.Context, userUID string) error // NEW
	ReassignUserUID(ctx context.Context, oldUID, newUID string) error
}

type noteRepo struct {
	db *gorm.DB
}

func NewNoteRepo(db *gorm.DB) NoteRepository {
	return &noteRepo{db: db}
}

func (r *noteRepo) Create(ctx context.Context, note *model.Note) error {
	return r.db.WithContext(ctx).Create(note).Error
}

func (r *noteRepo) GetByID(ctx context.Context, id uint, userUID string) (*model.Note, error) {
	var n model.Note
	// Ensure the note belongs to the user
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_uid = ?", id, userUID).
		First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *noteRepo) Update(ctx context.Context, note *model.Note) error {
	return r.db.WithContext(ctx).Save(note).Error
}

func (r *noteRepo) Delete(ctx context.Context, id uint, userUID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_uid = ?", id, userUID).
		Delete(&model.Note{}).Error
}

func (r *noteRepo) ListByUser(ctx context.Context, userUID string) ([]model.Note, error) {
	var notes []model.Note
	if err := r.db.WithContext(ctx).
		Where("user_uid = ?", userUID).
		Order("created_at DESC").
		Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

// Search by name OR text using ILIKE or LIKE (depending on Postgres)
func (r *noteRepo) SearchByUser(ctx context.Context, userUID, query string) ([]model.Note, error) {
	var notes []model.Note
	pattern := "%" + query + "%"
	if err := r.db.WithContext(ctx).
		Where("user_uid = ? AND (name ILIKE ? OR text ILIKE ?)", userUID, pattern, pattern).
		Order("created_at DESC").
		Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *noteRepo) DeleteAllByUserUID(ctx context.Context, userUID string) error {
	return r.db.WithContext(ctx).
		Where("user_uid = ?", userUID).
		Delete(&model.Note{}).Error
}

func (r *noteRepo) ReassignUserUID(ctx context.Context, oldUID, newUID string) error {
	if oldUID == "" || oldUID == newUID {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Note{}).
		Where("user_uid = ?", oldUID).
		Update("user_uid", newUID).Error
}
