// internal/usecase/message_usecase.go

package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
	"gorm.io/gorm"
)

type MessageUsecase interface {
	ListMessages(ctx context.Context, chatID uint) ([]model.Message, error)
	CreateMessage(ctx context.Context, msg *model.Message) error
	DeleteMessageBranch(ctx context.Context, messageID uint, chatID uint, userUID string) error
	GetMessageChain(ctx context.Context, leafMessageID uint, chatID uint, userUID string) ([]model.Message, error)
}

type messageUsecase struct {
	messageRepo repository.MessageRepository
	chatRepo    repository.ChatRepository
}

func NewMessageUsecase(mr repository.MessageRepository, cr repository.ChatRepository) MessageUsecase {
	return &messageUsecase{
		messageRepo: mr,
		chatRepo:    cr,
	}
}

func (uc *messageUsecase) ListMessages(ctx context.Context, chatID uint) ([]model.Message, error) {
	// This now correctly returns only the active message chain.
	return uc.messageRepo.ListActiveByChat(ctx, chatID)
}

func (uc *messageUsecase) CreateMessage(ctx context.Context, msg *model.Message) error {
	// If ParentID is not provided, find the last active message in the chat and set it as parent.
	if msg.ParentID == nil {
		lastMsg, err := uc.messageRepo.GetLastActiveMessage(ctx, msg.ChatID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err // Real error
		}
		if lastMsg != nil {
			msg.ParentID = &lastMsg.ID
		}
	} else {
		// Ensure the parent message exists and belongs to the same chat
		parentMsg, err := uc.messageRepo.GetByID(ctx, *msg.ParentID)
		if err != nil {
			return fmt.Errorf("invalid parent_id: %w", err)
		}
		if parentMsg.ChatID != msg.ChatID {
			return errors.New("parent message belongs to different chat")
		}
	}
	return uc.messageRepo.Create(ctx, msg)
}

// DeleteMessageBranch deactivates a message and its entire descendant tree.
func (uc *messageUsecase) DeleteMessageBranch(ctx context.Context, messageID uint, chatID uint, userUID string) error {
	// 1. Verify user has access to the chat.
	if _, err := uc.chatRepo.GetByID(ctx, chatID, userUID); err != nil {
		return fmt.Errorf("chat access denied or not found: %w", err)
	}

	// 2. Verify the message belongs to the chat.
	msg, err := uc.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}
	if msg.ChatID != chatID {
		return errors.New("message does not belong to the specified chat")
	}

	// 3. Use recursive CTE to deactivate entire branch in a single query
	return uc.messageRepo.DeactivateMessageBranch(ctx, messageID, chatID)
}

// GetMessageChain validates access and then fetches the message history.
func (uc *messageUsecase) GetMessageChain(ctx context.Context, leafMessageID uint, chatID uint, userUID string) ([]model.Message, error) {
	if _, err := uc.chatRepo.GetByID(ctx, chatID, userUID); err != nil {
		return nil, fmt.Errorf("chat access denied or not found: %w", err)
	}
	msg, err := uc.messageRepo.GetByID(ctx, leafMessageID)
	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}
	if msg.ChatID != chatID {
		return nil, errors.New("message does not belong to the specified chat")
	}

	return uc.messageRepo.GetMessageChain(ctx, leafMessageID)
}
