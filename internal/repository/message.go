package repository

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type MessageRepository interface {
	ListActiveByChat(ctx context.Context, chatID uint) ([]model.Message, error)
	GetByID(ctx context.Context, messageID uint) (*model.Message, error)
	Create(ctx context.Context, msg *model.Message) error
	UpdateActiveStatus(ctx context.Context, messageIDs []uint, isActive bool) error
	GetLastActiveMessage(ctx context.Context, chatID uint) (*model.Message, error)
	GetMessageChain(ctx context.Context, leafMessageID uint) ([]model.Message, error)
	DeactivateMessageBranch(ctx context.Context, rootID uint, chatID uint) error
	// --- Missing method signature added here ---
	GetLastAssistantMessageByChatID(ctx context.Context, chatID uint) (*model.Message, error)
	DeleteAllByUserUID(ctx context.Context, userUID string) error // NEW
}

type messageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) MessageRepository {
	return &messageRepo{db: db}
}

func (r *messageRepo) ListActiveByChat(ctx context.Context, chatID uint) ([]model.Message, error) {
	var msgs []model.Message
	if err := r.db.WithContext(ctx).
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Order("created_at ASC").
		Find(&msgs).Error; err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *messageRepo) GetByID(ctx context.Context, messageID uint) (*model.Message, error) {
	var msg model.Message
	if err := r.db.WithContext(ctx).First(&msg, messageID).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepo) Create(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *messageRepo) UpdateActiveStatus(ctx context.Context, messageIDs []uint, isActive bool) error {
	if len(messageIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("id IN ?", messageIDs).
		Update("is_active", isActive).Error
}

func (r *messageRepo) GetLastActiveMessage(ctx context.Context, chatID uint) (*model.Message, error) {
	var msg model.Message
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Order("created_at DESC").
		First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetMessageChain traverses up the parent links to get the full conversation history for a given message.
// The returned slice is ordered from the root of the conversation to the leafMessage.
func (r *messageRepo) GetMessageChain(ctx context.Context, leafMessageID uint) ([]model.Message, error) {
	var chain []model.Message
	currentID := &leafMessageID

	for currentID != nil {
		var msg model.Message
		if err := r.db.WithContext(ctx).First(&msg, *currentID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return nil, err
		}
		chain = append(chain, msg)
		currentID = msg.ParentID
	}

	// Reverse the slice to get chronological order
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain, nil
}

func (r *messageRepo) DeactivateMessageBranch(ctx context.Context, rootID uint, chatID uint) error {
	return r.db.WithContext(ctx).Exec(`
    WITH RECURSIVE descendants AS (
      SELECT id 
      FROM messages 
      WHERE id = ? 
      UNION ALL
      SELECT m.id 
      FROM messages m 
      INNER JOIN descendants d ON m.parent_id = d.id
    ) 
    UPDATE messages 
    SET is_active = false 
    WHERE id IN (SELECT id FROM descendants) AND chat_id = ?
  `, rootID, chatID).Error
}

// --- NEW METHOD IMPLEMENTATION ---
// GetLastAssistantMessageByChatID fetches the most recent, active message from the 'assistant'.
func (r *messageRepo) GetLastAssistantMessageByChatID(ctx context.Context, chatID uint) (*model.Message, error) {
	var msg model.Message
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND is_active = ? AND role = ?", chatID, true, "assistant").
		Order("created_at DESC").
		First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepo) DeleteAllByUserUID(ctx context.Context, userUID string) error {
	// delete messages where chat belongs to user
	return r.db.WithContext(ctx).Exec(`
		DELETE FROM messages
		WHERE chat_id IN (SELECT id FROM chats WHERE user_uid = ?)`, userUID).Error
}
