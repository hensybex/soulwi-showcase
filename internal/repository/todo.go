// File: internal/repository/todo_repository.go

package repository

import (
	"context"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type TodoRepository interface {
	Create(ctx context.Context, todo *model.Todo) error
	GetByID(ctx context.Context, id uint, userUID string) (*model.Todo, error)
	Update(ctx context.Context, todo *model.Todo) error
	Delete(ctx context.Context, id uint, userUID string) error
	ListByUserAndDay(ctx context.Context, userUID string, day time.Time) ([]model.Todo, error)
	SetDone(ctx context.Context, id uint, userUID string, done bool) error
	DeleteAllByUserUID(ctx context.Context, userUID string) error // NEW
	ReassignUserUID(ctx context.Context, oldUID, newUID string) error
}

type todoRepo struct {
	db *gorm.DB
}

func NewTodoRepo(db *gorm.DB) TodoRepository {
	return &todoRepo{db: db}
}

func (r *todoRepo) Create(ctx context.Context, todo *model.Todo) error {
	return r.db.WithContext(ctx).Create(todo).Error
}

func (r *todoRepo) GetByID(ctx context.Context, id uint, userUID string) (*model.Todo, error) {
	var t model.Todo
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_uid = ?", id, userUID).
		First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *todoRepo) Update(ctx context.Context, todo *model.Todo) error {
	return r.db.WithContext(ctx).Save(todo).Error
}

func (r *todoRepo) Delete(ctx context.Context, id uint, userUID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_uid = ?", id, userUID).
		Delete(&model.Todo{}).Error
}

func (r *todoRepo) ListByUserAndDay(ctx context.Context, userUID string, day time.Time) ([]model.Todo, error) {
	var todos []model.Todo
	startOfDay := day
	endOfDay := day.Add(24 * time.Hour)

	if err := r.db.WithContext(ctx).
		Where("user_uid = ? AND target_day >= ? AND target_day < ?", userUID, startOfDay, endOfDay).
		Order("created_at ASC").
		Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *todoRepo) SetDone(ctx context.Context, id uint, userUID string, done bool) error {
	return r.db.WithContext(ctx).
		Model(&model.Todo{}).
		Where("id = ? AND user_uid = ?", id, userUID).
		Update("is_done", done).Error
}

func (r *todoRepo) DeleteAllByUserUID(ctx context.Context, userUID string) error {
	return r.db.WithContext(ctx).
		Where("user_uid = ?", userUID).
		Delete(&model.Todo{}).Error
}

func (r *todoRepo) ReassignUserUID(ctx context.Context, oldUID, newUID string) error {
	if oldUID == "" || oldUID == newUID {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Todo{}).
		Where("user_uid = ?", oldUID).
		Update("user_uid", newUID).Error
}
