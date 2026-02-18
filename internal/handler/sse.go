// internal/handler/sse.go

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type SSEHandler struct {
	aiUsecase              usecase.AIUsecase
	messageUC              usecase.MessageUsecase
	messageRepo            repository.MessageRepository
	chatUC                 usecase.ChatUsecase
	messageLimitMiddleware *middleware.MessageLimitMiddleware
}

func NewSSEHandler(
	aiUsecase usecase.AIUsecase,
	messageUC usecase.MessageUsecase,
	messageRepo repository.MessageRepository,
	chatUC usecase.ChatUsecase,
	messageLimitMiddleware *middleware.MessageLimitMiddleware,
) *SSEHandler {
	return &SSEHandler{
		aiUsecase:              aiUsecase,
		messageUC:              messageUC,
		messageRepo:            messageRepo,
		chatUC:                 chatUC,
		messageLimitMiddleware: messageLimitMiddleware,
	}
}

func (h *SSEHandler) DeleteMessageBranch(c *gin.Context) {
	firebaseUIDVal, _ := c.Get("firebase_uid")
	firebaseUID, _ := firebaseUIDVal.(string)

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

	err = h.messageUC.DeleteMessageBranch(c.Request.Context(), uint(messageID), uint(chatID), firebaseUID)
	if err != nil {
		log.Printf("Failed to deactivate message branch for message %d: %v", messageID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message branch"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "branch deactivated"})
}

// *** THIS FUNCTION IS THE FIX ***
// It now correctly formats multi-line data for SSE streams.
func (h *SSEHandler) writeSSEChunk(c *gin.Context, content string) {
	// Replace every newline in the content with a newline followed by a "data: " prefix.
	// This correctly formats the data for the SSE spec.
	escapedData := strings.ReplaceAll(content, "\n", "\ndata: ")

	ssePayload := fmt.Sprintf("data: %s\n\n", escapedData)

	if _, err := io.WriteString(c.Writer, ssePayload); err != nil {
		// Client likely disconnected, just return
		return
	}
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (h *SSEHandler) writeSSEEvent(c *gin.Context, eventName string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal SSE event data: %v", err)
		return
	}
	ssePayload := fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, string(jsonData))
	if _, err := io.WriteString(c.Writer, ssePayload); err != nil {
		// Client likely disconnected, just return
		return
	}
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (h *SSEHandler) ChatStreamSSE(c *gin.Context) {
	firebaseUIDVal, _ := c.Get("firebase_uid")
	firebaseUID, _ := firebaseUIDVal.(string)
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat_id"})
		return
	}

	chat, err := h.chatUC.GetChat(c.Request.Context(), uint(chatID), firebaseUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found or not yours"})
		return
	}

	existingMessages, err := h.messageUC.ListMessages(c.Request.Context(), uint(chatID))
	if err != nil {
		// Log the error but don't block the chat. The rename is a non-critical enhancement.
		log.Printf("Error fetching messages for chat %d, continuing without rename check: %v", chatID, err)
	}
	shouldRenameChat := chat.PromptID != nil && *chat.PromptID == 1 && len(existingMessages) == 0

	var reqBody struct {
		Content  string `json:"content"`
		ParentID *uint  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var newUserMessage model.Message
	if reqBody.Content != "" {
		newUserMessage = model.Message{
			ChatID:   uint(chatID),
			Role:     "user",
			Content:  reqBody.Content,
			ParentID: reqBody.ParentID,
			IsActive: true,
		}
		if err := h.messageUC.CreateMessage(c.Request.Context(), &newUserMessage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new message"})
			return
		}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")
	flusher, _ := c.Writer.(http.Flusher)
	flusher.Flush()

	var parentMessageIDForAI *uint
	if newUserMessage.ID != 0 {
		parentMessageIDForAI = &newUserMessage.ID
	} else if reqBody.ParentID != nil {
		parentMessageIDForAI = reqBody.ParentID
	}

	outChan, errChan := h.aiUsecase.StartChatStream(c.Request.Context(), uint(chatID), firebaseUID, newUserMessage.Content, parentMessageIDForAI)

	var streamErr error
	clientClosed := c.Request.Context().Done()
	var fullResponse strings.Builder

	for {
		select {
		case partial, ok := <-outChan:
			if !ok {
				outChan = nil
			} else {
				// This now calls the corrected helper function
				h.writeSSEChunk(c, partial)
				fullResponse.WriteString(partial)
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else if err != nil {
				streamErr = err
				// ADD THIS LOGIC TO NOTIFY THE CLIENT
				log.Printf("Error from AI Usecase, sending to client: %v", err)
				h.writeSSEEvent(c, "error", gin.H{"message": "An error occurred while communicating with the AI model."})
				goto end_loop // Exit the loop immediately on error
			}
		case <-clientClosed:
			goto end_loop
		}
		if outChan == nil && errChan == nil {
			break
		}
	}
end_loop:

	if streamErr == nil && c.Request.Context().Err() == nil {
		if newUserMessage.ID != 0 {
			h.writeSSEEvent(c, "user_message_created", newUserMessage)
		}

		aiContent := fullResponse.String()
		if aiContent != "" {
			aiMessage := model.Message{
				ChatID:   uint(chatID),
				Role:     "assistant",
				Content:  aiContent,
				ParentID: parentMessageIDForAI,
				IsActive: true,
			}
			if err := h.messageUC.CreateMessage(c.Request.Context(), &aiMessage); err == nil {
				h.writeSSEEvent(c, "assistant_message_created", aiMessage)

				// If this is the first message in a default chat, generate a name for it.
				if shouldRenameChat {
					newName, err := h.aiUsecase.GenerateChatName(c.Request.Context(), newUserMessage.Content, aiContent)
					if err != nil {
						log.Printf("Failed to generate chat name for chat %d: %v", chatID, err)
					} else {
						if err := h.chatUC.RenameChat(c.Request.Context(), uint(chatID), firebaseUID, newName); err != nil {
							log.Printf("Failed to rename chat %d: %v", chatID, err)
						} else {
							log.Printf("Successfully renamed chat %d to '%s'", chatID, newName)
							h.writeSSEEvent(c, "chat_renamed", gin.H{"chat_id": chatID, "new_name": newName})
						}
					}
				}
			} else {
				log.Printf("Failed to save final assistant message for chat %d: %v", chatID, err)
			}
		}
	}

	if streamErr == nil && c.Request.Context().Err() == nil && reqBody.Content != "" {
		err := h.messageLimitMiddleware.IncrementMessageCount(context.Background(), firebaseUID)
		if err != nil {
			log.Printf("CRITICAL: Failed to increment message count for user %s: %v", firebaseUID, err)
		}
	}
}

// MockChatStreamSSE mimics the real SSE handler, including DB operations.
func (h *SSEHandler) MockChatStreamSSE(c *gin.Context) {
	// --- 1. Authentication and Authorization (Mimics the real endpoint) ---
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, _ := firebaseUIDVal.(string)

	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat_id"})
		return
	}

	if _, err := h.chatUC.GetChat(c.Request.Context(), uint(chatID), firebaseUID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found or not yours"})
		return
	}

	// --- 2. Parse Request Body (Mimics the real endpoint) ---
	var reqBody struct {
		Content  string `json:"content"`
		ParentID *uint  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// --- 3. Save User Message to DB (Mimics the real endpoint) ---
	var newUserMessage model.Message
	if reqBody.Content != "" {
		newUserMessage = model.Message{
			ChatID:   uint(chatID),
			Role:     "user",
			Content:  reqBody.Content,
			ParentID: reqBody.ParentID,
			IsActive: true,
		}
		if err := h.messageUC.CreateMessage(c.Request.Context(), &newUserMessage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Mock failed to save new message"})
			return
		}
	}

	// --- 4. Prepare for SSE Stream ---
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		log.Println("Streaming unsupported!")
		return
	}
	flusher.Flush()

	// --- 5. Generate, Save, and Stream Mock Response ---
	mockMessageStr := fmt.Sprintf("This is a mock database-saved response to your message: '%s'. I am a test endpoint proving the full request/response/save cycle.", reqBody.Content)

	// Save the mock assistant response to the DB
	var parentIDForMock *uint
	if newUserMessage.ID != 0 {
		parentIDForMock = &newUserMessage.ID
	} else if reqBody.ParentID != nil {
		parentIDForMock = reqBody.ParentID
	}

	mockAssistantMessage := model.Message{
		ChatID:   uint(chatID),
		Role:     "assistant",
		Content:  mockMessageStr,
		ParentID: parentIDForMock, // Link response to the user's message or original parent
		IsActive: true,
	}

	// Confirm user message creation first, if applicable
	if newUserMessage.ID != 0 {
		h.writeSSEEvent(c, "user_message_created", newUserMessage)
	}

	if reqBody.Content != "" || reqBody.ParentID != nil { // Save assistant message if there was a user message OR it's a regen
		if err := h.messageUC.CreateMessage(c.Request.Context(), &mockAssistantMessage); err != nil {
			// Log the error but don't stop the stream, the client might still want the text.
			log.Printf("ERROR: Failed to save mock assistant message to DB: %v", err)
		}
	}

	// Stream the response chunk by chunk
	words := strings.Fields(mockMessageStr)
	for i := 0; i < len(words); i += 3 {
		end := i + 3
		if end > len(words) {
			end = len(words)
		}
		prefix := " "
		if i == 0 {
			prefix = ""
		}
		chunk := prefix + strings.Join(words[i:end], " ")

		h.writeSSEChunk(c, chunk)
		time.Sleep(100 * time.Millisecond)
	}

	// **FIX: The mock must also send the final created message event, just like the real handler.**
	if mockAssistantMessage.ID != 0 { // Check if it was saved
		h.writeSSEEvent(c, "assistant_message_created", mockAssistantMessage)
	}

	// Also increment the message count, just like the real handler
	if reqBody.Content != "" {
		err := h.messageLimitMiddleware.IncrementMessageCount(context.Background(), firebaseUID)
		if err != nil {
			log.Printf("CRITICAL: Mock handler failed to increment message count for user %s: %v", firebaseUID, err)
		}
	}
}
