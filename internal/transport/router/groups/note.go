// File: internal/transport/router/groups/note.go

package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
)

func ConfigureNoteRoutes(r *gin.Engine, c *di.Container, authMiddleware gin.HandlerFunc) {
	notesGroup := r.Group("/notes")
	// old code had c.FirebaseAuth => user or admin
	notesGroup.Use(authMiddleware, middleware.RequireRoleMiddleware("admin", "user"))

	notesGroup.GET("/", c.NoteHandler.ListNotes)
	notesGroup.POST("/", c.NoteHandler.CreateNote)
	notesGroup.GET("/:id", c.NoteHandler.GetNote)
	notesGroup.PUT("/:id", c.NoteHandler.UpdateNote)
	notesGroup.DELETE("/:id", c.NoteHandler.DeleteNote)
}
