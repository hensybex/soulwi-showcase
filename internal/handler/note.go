// File: internal/handler/note.go

package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type NoteHandler struct {
	noteUC usecase.NoteUsecase
}

func NewNoteHandler(nu usecase.NoteUsecase) *NoteHandler {
	return &NoteHandler{noteUC: nu}
}

// GET /notes?user_uid=xxx[&search=...]
func (h *NoteHandler) ListNotes(c *gin.Context) {
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

	search := c.Query("search")
	if search == "" {
		// normal listing
		notes, err := h.noteUC.ListNotes(c, firebaseUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list notes"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": notes})
	} else {
		// search
		notes, err := h.noteUC.SearchNotes(c, firebaseUID, search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search notes"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": notes})
	}
}

// POST /notes?user_uid=xxx
func (h *NoteHandler) CreateNote(c *gin.Context) {
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

	var note model.Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	note.UserUID = firebaseUID

	if err := h.noteUC.CreateNote(c, &note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": note})
}

// GET /notes/:id?user_uid=xxx
func (h *NoteHandler) GetNote(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	note, err := h.noteUC.GetNote(c, uint(idUint64), firebaseUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": note})
}

// PUT /notes/:id?user_uid=xxx
func (h *NoteHandler) UpdateNote(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	var req model.Note
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.ID = uint(idUint64)
	req.UserUID = firebaseUID

	if err := h.noteUC.UpdateNote(c, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Note updated"})
}

// DELETE /notes/:id?user_uid=xxx
func (h *NoteHandler) DeleteNote(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	if err := h.noteUC.DeleteNote(c, uint(idUint64), firebaseUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
}
