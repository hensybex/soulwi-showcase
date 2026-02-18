// internal/repository/chat_repository.go

package repository

import (
	"context"
	"log"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type ChatRepository interface {
	ListByUser(ctx context.Context, userUID string) ([]model.Chat, error)
	GetByID(ctx context.Context, id uint, userUID string) (*model.Chat, error)
	Create(ctx context.Context, chat *model.Chat) error
	Delete(ctx context.Context, id uint, userUID string) error
	UpdateName(ctx context.Context, id uint, userUID, newName string) error
	DeleteAllByUserUID(ctx context.Context, userUID string) error // NEW
	ReassignUserUID(ctx context.Context, oldUID, newUID string) error
}

type chatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) ChatRepository {
	return &chatRepo{db: db}
}

func (r *chatRepo) ListByUser(ctx context.Context, userUID string) ([]model.Chat, error) {
	var chats []model.Chat
	if err := r.db.WithContext(ctx).
		Where("user_uid = ?", userUID).
		Order("created_at DESC").
		Find(&chats).Error; err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *chatRepo) GetByID(ctx context.Context, id uint, userUID string) (*model.Chat, error) {
	var c model.Chat
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_uid = ?", id, userUID).
		First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *chatRepo) Create(ctx context.Context, chat *model.Chat) error {
	return r.db.WithContext(ctx).Create(chat).Error
}

func (r *chatRepo) Delete(ctx context.Context, id uint, userUID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_uid = ?", id, userUID).
		Delete(&model.Chat{}).Error
}

func (r *chatRepo) UpdateName(ctx context.Context, id uint, userUID, newName string) error {
	log.Println("IM HERE")
	log.Println(id)
	log.Println(userUID)
	log.Println(newName)
	return r.db.WithContext(ctx).
		Model(&model.Chat{}).
		Where("id = ? AND user_uid = ?", id, userUID).
		Update("name", newName).Error
}

func (r *chatRepo) DeleteAllByUserUID(ctx context.Context, userUID string) error {
	return r.db.WithContext(ctx).
		Where("user_uid = ?", userUID).
		Delete(&model.Chat{}).Error
}

func (r *chatRepo) ReassignUserUID(ctx context.Context, oldUID, newUID string) error {
	if oldUID == "" || oldUID == newUID {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Chat{}).
		Where("user_uid = ?", oldUID).
		Update("user_uid", newUID).Error
}
