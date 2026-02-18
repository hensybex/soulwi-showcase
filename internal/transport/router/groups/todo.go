// File: internal/transport/router/groups/todo.go

package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
)

func ConfigureTodoRoutes(r *gin.Engine, c *di.Container, authMiddleware gin.HandlerFunc) {
	todosGroup := r.Group("/todos")
	// old: c.FirebaseAuth.Middleware() => means user or admin
	todosGroup.Use(authMiddleware, middleware.RequireRoleMiddleware("admin", "user"))

	// Changed "/" to "" to match the path "/todos" instead of "/todos/"
	todosGroup.GET("", c.TodoHandler.ListTodos)
	todosGroup.POST("", c.TodoHandler.CreateTodo)
	todosGroup.GET("/:id", c.TodoHandler.GetTodo)
	todosGroup.PUT("/:id", c.TodoHandler.UpdateTodo)
	todosGroup.DELETE("/:id", c.TodoHandler.DeleteTodo)
	todosGroup.PATCH("/:id/complete", c.TodoHandler.CompleteTodo)
}
