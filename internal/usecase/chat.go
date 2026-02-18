// internal/usecase/chat_usecase.go

package usecase

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
)

type ChatUsecase interface {
	ListChats(ctx context.Context, userUID string) ([]model.Chat, error)
	GetChat(ctx context.Context, chatID uint, userUID string) (*model.Chat, error)
	CreateChat(ctx context.Context, chat *model.Chat) (*model.Chat, error)
	DeleteChat(ctx context.Context, chatID uint, userUID string) error
	RenameChat(ctx context.Context, chatID uint, userUID, newName string) error
}

type chatUsecase struct {
	chatRepo   repository.ChatRepository
	promptRepo repository.PromptRepository
}

func NewChatUsecase(
	cr repository.ChatRepository,
	pr repository.PromptRepository,
) ChatUsecase {
	return &chatUsecase{
		chatRepo:   cr,
		promptRepo: pr,
	}
}

func (uc *chatUsecase) ListChats(ctx context.Context, userUID string) ([]model.Chat, error) {
	return uc.chatRepo.ListByUser(ctx, userUID)
}

func (uc *chatUsecase) GetChat(ctx context.Context, chatID uint, userUID string) (*model.Chat, error) {
	return uc.chatRepo.GetByID(ctx, chatID, userUID)
}

func (uc *chatUsecase) CreateChat(ctx context.Context, chat *model.Chat) (*model.Chat, error) {
	prompt, err := uc.promptRepo.GetByID(ctx, *chat.PromptID)
	if err != nil {
		return chat, err
	}
	newChat := &model.Chat{
		UserUID:  chat.UserUID,
		Name:     prompt.Name,
		PromptID: chat.PromptID,
	}
	err = uc.chatRepo.Create(ctx, newChat)
	return newChat, err
}

func (uc *chatUsecase) DeleteChat(ctx context.Context, chatID uint, userUID string) error {
	return uc.chatRepo.Delete(ctx, chatID, userUID)
}

func (uc *chatUsecase) RenameChat(ctx context.Context, chatID uint, userUID, newName string) error {
	return uc.chatRepo.UpdateName(ctx, chatID, userUID, newName)
}
