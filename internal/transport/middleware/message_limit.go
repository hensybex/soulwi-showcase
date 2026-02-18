// internal/transport/middleware/message_limit.go (обновленная версия)
package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
	"gorm.io/gorm"
)

type MessageLimitMiddleware struct {
	userRepo repository.UserRepository
}

func NewMessageLimitMiddleware(userRepo repository.UserRepository) *MessageLimitMiddleware {
	return &MessageLimitMiddleware{
		userRepo: userRepo,
	}
}

// CheckMessageLimit проверяет, может ли пользователь отправить сообщение.
// Этот middleware должен применяться к роуту отправки сообщения.
func (m *MessageLimitMiddleware) CheckMessageLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		firebaseUIDVal, exists := c.Get("firebase_uid")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
			return
		}

		firebaseUID, ok := firebaseUIDVal.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
			return
		}

		user, err := m.userRepo.GetByFirebaseUID(c.Request.Context(), firebaseUID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("[MESSAGE_LIMIT] User not found for UID=%s", firebaseUID)
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{
					"error":   "user_not_synced",
					"message": "User record missing. Please retry syncing your account.",
				})
			} else {
				log.Printf("[MESSAGE_LIMIT] Failed to get user for UID=%s: %v", firebaseUID, err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user status"})
			}
			return
		}

		// Проверяем, может ли пользователь отправить сообщение.
		// Эта функция МОДИФИЦИРУЕТ `user`, если началась новая неделя.
		canSend := user.CanSendMessage()

		// Важно: так как CanSendMessage мог изменить состояние user (сбросить счетчик),
		// мы должны сохранить эти изменения в базе данных.
		if err := m.userRepo.Update(c.Request.Context(), user); err != nil {
			log.Printf("[MESSAGE_LIMIT] Failed to update user after CanSendMessage for user %s: %v", firebaseUID, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
			return
		}

		if !canSend {
			log.Printf("[MESSAGE_LIMIT] Weekly limit reached for UID=%s count=%d", firebaseUID, user.WeeklyMessageCount)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":        "Message limit reached",
				"message":      "You have reached your weekly limit of 15 messages. Upgrade to premium for unlimited messages.",
				"weekly_count": user.WeeklyMessageCount,
				"limit":        model.WeeklyMessageLimit,
			})
			return
		}

		// Сохраняем обновленного пользователя в контексте для дальнейшего использования в хендлере.
		c.Set("user", user)

		c.Next()
	}
}

// IncrementMessageCount увеличивает счетчик сообщений ПОСЛЕ успешной отправки.
// Вызывается из SSE хендлера.
func (m *MessageLimitMiddleware) IncrementMessageCount(ctx context.Context, firebaseUID string) error {
	user, err := m.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return err
	}

	// Увеличиваем счетчик
	user.IncrementMessageCount()

	// Сохраняем финальное состояние
	return m.userRepo.Update(ctx, user)
}
