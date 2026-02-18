// File: internal/handler/todo.go

package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

// RequestLogger is a middleware function that logs detailed information about each incoming request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Log request details BEFORE processing
		log.Printf(
			"[INCOMING_REQUEST] Method: %s | Path: %s | RawPath: %s | Host: %s | From: %s",
			c.Request.Method,
			c.Request.URL.Path, // This is the most important field for us
			c.Request.URL.RawPath,
			c.Request.Host,
			c.Request.RemoteAddr,
		)

		c.Next() // Process the request

		// Log result AFTER processing
		latency := time.Since(start)
		log.Printf(
			"[REQUEST_COMPLETE] Status: %d | Latency: %s",
			c.Writer.Status(),
			latency,
		)
	}
}

type TodoHandler struct {
	todoUC usecase.TodoUsecase
}

func NewTodoHandler(tu usecase.TodoUsecase) *TodoHandler {
	return &TodoHandler{todoUC: tu}
}

// GET /todos?user_uid=xxx&day=yyyy-mm-dd
func (h *TodoHandler) ListTodos(c *gin.Context) {
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
	dayStr := c.Query("day")
	if dayStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing day"})
		return
	}
	parsedDay, err := time.Parse("2006-01-02", dayStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day format, use yyyy-mm-dd"})
		return
	}

	todos, err := h.todoUC.ListTodos(c.Request.Context(), firebaseUID, parsedDay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list todos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": todos})
}

// POST /todos?user_uid=xxx
func (h *TodoHandler) CreateTodo(c *gin.Context) {
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
	var t model.Todo
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	t.UserUID = firebaseUID

	if err := h.todoUC.CreateTodo(c.Request.Context(), &t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": t})
}

// GET /todos/:id?user_uid=xxx
func (h *TodoHandler) GetTodo(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	todo, err := h.todoUC.GetTodo(c.Request.Context(), uint(idUint64), firebaseUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": todo})
}

// PUT /todos/:id?user_uid=xxx
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	var t model.Todo
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	t.ID = uint(idUint64)
	t.UserUID = firebaseUID

	if err := h.todoUC.UpdateTodo(c.Request.Context(), &t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Todo updated"})
}

// DELETE /todos/:id?user_uid=xxx
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}
	if err := h.todoUC.DeleteTodo(c.Request.Context(), uint(idUint64), firebaseUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
}

// PATCH /todos/:id/complete?user_uid=xxx
func (h *TodoHandler) CompleteTodo(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}
	if err := h.todoUC.CompleteTodo(c.Request.Context(), uint(idUint64), firebaseUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set todo completed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Todo marked as done"})
}
