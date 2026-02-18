// internal/handler/feedback.go

package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type FeedbackHandler struct {
	uc usecase.FeedbackUsecase
}

func NewFeedbackHandler(uc usecase.FeedbackUsecase) *FeedbackHandler {
	return &FeedbackHandler{uc: uc}
}

// POST /feedback
func (h *FeedbackHandler) CreateFeedback(c *gin.Context) {
	firebaseUIDVal, ok := c.Get("firebase_uid")
	if !ok {
		log.Println("[ERROR] firebase_uid missing in JWT context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid user JWT"})
		return
	}

	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		log.Println("[ERROR] Invalid firebase_uid format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase UID in context"})
		return
	}

	// Используй firebaseUID (string) вместо numeric userID
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := h.uc.CreateFeedback(c.Request.Context(), firebaseUID, req.Text); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save feedback"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Feedback saved"})
}
