// internal/handler/wise_phrase.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type WisePhraseHandler struct {
	wpUC usecase.WisePhraseUsecase
}

func NewWisePhraseHandler(uc usecase.WisePhraseUsecase) *WisePhraseHandler {
	return &WisePhraseHandler{wpUC: uc}
}

// POST /wise-phrases/generate?prompt=xxx&count=20
func (h *WisePhraseHandler) GenerateBatch(c *gin.Context) {
	// Use a default count (10) if none is provided.
	countStr := c.Query("count")
	count := 10
	if countStr != "" {
		if val, err := strconv.Atoi(countStr); err == nil {
			count = val
		}
	}

	// Always use the "Wise Words Generator" prompt.
	err := h.wpUC.GenerateBatch(c, "Wise Words Generator", count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Batch generation complete"})
}

// GET /wise-phrases/admin?page=1&page_size=20
func (h *WisePhraseHandler) ListAdmin(c *gin.Context) {
	page := 1
	pageSize := 20

	if pStr := c.DefaultQuery("page", "1"); pStr != "" {
		if val, err := strconv.Atoi(pStr); err == nil && val > 0 {
			page = val
		}
	}

	if psStr := c.DefaultQuery("page_size", "20"); psStr != "" {
		if val, err := strconv.Atoi(psStr); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	offset := (page - 1) * pageSize

	phrases, total, err := h.wpUC.ListAdminPhrases(c, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": phrases,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// DELETE /wise-phrases/:id
func (h *WisePhraseHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	phraseID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phrase ID"})
		return
	}

	if err := h.wpUC.DeleteWisePhrase(c, uint(phraseID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /wise-phrases/random?user_uid=xxx
func (h *WisePhraseHandler) GetRandom(c *gin.Context) {
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
	wp, err := h.wpUC.GetRandomPhrase(c, firebaseUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wp})
}

// POST /wise-phrases/:id/like?user_uid=xxx
func (h *WisePhraseHandler) ToggleLikePhrase(c *gin.Context) {
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
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phrase ID"})
		return
	}
	err = h.wpUC.ToggleLikePhrase(c, firebaseUID, uint(idUint64))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Liked"})
}

// GET /wise-phrases/liked?user_uid=xxx
func (h *WisePhraseHandler) ListLikes(c *gin.Context) {
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
	phrases, err := h.wpUC.ListUserLikes(c, firebaseUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": phrases})
}

func (h *WisePhraseHandler) RecordShare(c *gin.Context) {
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
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phrase ID"})
		return
	}

	err = h.wpUC.RecordShare(c, firebaseUID, uint(idUint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Share recorded"})
}
