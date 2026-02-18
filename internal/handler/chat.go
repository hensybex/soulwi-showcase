// internal/handler/chat_handler.go

package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type ChatHandler struct {
	chatUC usecase.ChatUsecase
}

func NewChatHandler(chatUC usecase.ChatUsecase) *ChatHandler {
	return &ChatHandler{chatUC: chatUC}
}

// GET /chats?user_uid=xxx
func (h *ChatHandler) ListChats(c *gin.Context) {
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
	chats, err := h.chatUC.ListChats(c.Request.Context(), firebaseUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list chats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": chats})
}

// POST /chats?user_uid=xxx
func (h *ChatHandler) CreateChat(c *gin.Context) {
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
	var chat model.Chat
	if err := c.ShouldBindJSON(&chat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	chat.UserUID = firebaseUID

	newChat, err := h.chatUC.CreateChat(c.Request.Context(), &chat)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": newChat})
}

// GET /chats/:chat_id
func (h *ChatHandler) GetChat(c *gin.Context) {
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
	chat, err2 := h.chatUC.GetChat(c.Request.Context(), uint(chatID), firebaseUID)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat"})
		return
	}
	if chat == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": chat})
}

// DELETE /chats/:chat_id?user_uid=xxx
func (h *ChatHandler) DeleteChat(c *gin.Context) {
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
	if err := h.chatUC.DeleteChat(c.Request.Context(), uint(chatID), firebaseUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete chat"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Chat deleted"})
}

// PATCH /chats/:chat_id/rename?user_uid=xxx
func (h *ChatHandler) RenameChat(c *gin.Context) {
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

	var payload struct {
		NewName string `json:"new_name"`
	}
	if err2 := c.ShouldBindJSON(&payload); err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err3 := h.chatUC.RenameChat(c.Request.Context(), uint(chatID), firebaseUID, payload.NewName); err3 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rename chat"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Chat renamed"})
}
