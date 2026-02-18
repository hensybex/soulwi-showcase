// internal/handler/message_handler.go

package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
	"gorm.io/gorm"
)

type MessageHandler struct {
	messageUC usecase.MessageUsecase
	chatUC    usecase.ChatUsecase
}

func NewMessageHandler(mu usecase.MessageUsecase, cu usecase.ChatUsecase) *MessageHandler {
	return &MessageHandler{messageUC: mu, chatUC: cu}
}

// GET /chats/:chat_id/messages?user_uid=xxx
func (h *MessageHandler) ListMessages(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat_id"})
		return
	}
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}

	// confirm chat belongs to user
	if _, err2 := h.chatUC.GetChat(c.Request.Context(), uint(chatID), firebaseUID); err2 != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found or not yours"})
		return
	}

	messages, err3 := h.messageUC.ListMessages(c.Request.Context(), uint(chatID))
	if err3 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list messages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// DELETE /chats/:chat_id/messages/:message_id
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat_id"})
		return
	}

	messageIDStr := c.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message_id"})
		return
	}

	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}

	err = h.messageUC.DeleteMessageBranch(c.Request.Context(), uint(messageID), uint(chatID), firebaseUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Message or chat not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message branch", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message and subsequent messages deleted successfully"})
}
