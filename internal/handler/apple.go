// internal/handler/apple_handler.go
package handler

import (
	"log"
	"net/http"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type AppleHandler struct {
	subscriptionUC usecase.SubscriptionUsecase
	FirebaseAuth   *firebaseAuth.Client
}

func NewAppleHandler(subscriptionUC usecase.SubscriptionUsecase, fbAuthClient *firebaseAuth.Client) *AppleHandler {
	return &AppleHandler{
		subscriptionUC: subscriptionUC,
		FirebaseAuth:   fbAuthClient,
	}
}

// DTO structures
type VerifyReceiptRequest struct {
	ReceiptData string `json:"receiptData" binding:"required"`
	UserID      string `json:"userId" binding:"required"`
}

type VerifyReceiptResponse struct {
	Success      bool        `json:"success"`
	Subscription interface{} `json:"subscription,omitempty"`
	Message      string      `json:"message,omitempty"`
}

type ServerNotificationRequest struct {
	SignedPayload string `json:"signedPayload" binding:"required"`
}

// POST /apple/verify-receipt
func (h *AppleHandler) VerifyReceipt(c *gin.Context) {
	var req VerifyReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Проверяем, что пользователь аутентифицирован и имеет доступ к этому userID
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

	// В простой реализации userId должен совпадать с firebase_uid
	// В более сложной системе может быть дополнительная проверка доступа
	if req.UserID != firebaseUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Валидируем чек
	subscription, err := h.subscriptionUC.ValidateAppleReceipt(c.Request.Context(), req.ReceiptData, req.UserID)
	if err != nil {
		log.Printf("Failed to validate Apple receipt: %v", err)
		c.JSON(http.StatusBadRequest, VerifyReceiptResponse{
			Success: false,
			Message: "Receipt validation failed",
		})
		return
	}

	c.JSON(http.StatusOK, VerifyReceiptResponse{
		Success:      true,
		Subscription: subscription,
		Message:      "Receipt validated successfully",
	})
}

func (h *AppleHandler) VerifyReceiptTest(c *gin.Context) {
	log.Println("--- [Handler:VerifyReceiptTest] Received new request ---")
	var req VerifyReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Логируем ошибку парсинга JSON
		log.Printf("[Handler:VerifyReceiptTest] 🔴 ERROR: Failed to bind JSON. Details: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	// Логируем, что пришло от клиента
	log.Printf("[Handler:VerifyReceiptTest] ✅ JSON body bound successfully. UserID: %s, ReceiptData length: %d", req.UserID, len(req.ReceiptData))

	firebaseUIDVal, _ := c.Get("firebase_uid")
	firebaseUID, _ := firebaseUIDVal.(string)
	log.Printf("[Handler:VerifyReceiptTest] ℹ️ Firebase UID from token: %s", firebaseUID)

	if req.UserID != firebaseUID {
		log.Printf("[Handler:VerifyReceiptTest] 🔴 FORBIDDEN: Request UserID (%s) does not match Firebase UID (%s)", req.UserID, firebaseUID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	log.Println("[Handler:VerifyReceiptTest] ✅ UserID matches Firebase UID. Calling use case...")

	// Вызываем use case
	subscription, err := h.subscriptionUC.ValidateAppleTestJWS(c.Request.Context(), req.ReceiptData, req.UserID)
	if err != nil {
		// Это самая вероятная точка отказа. Логируем ошибку из use case.
		log.Printf("[Handler:VerifyReceiptTest] 🔴 ERROR from use case: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test JWS validation failed", "details": err.Error()})
		return
	}

	log.Println("[Handler:VerifyReceiptTest] ✅ Use case finished successfully. Returning 200 OK.")
	c.JSON(http.StatusOK, gin.H{"success": true, "subscription": subscription})
}

// POST /apple/server-notifications
func (h *AppleHandler) HandleServerNotifications(c *gin.Context) {
	var req ServerNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid server notification request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Обрабатываем уведомление
	if err := h.subscriptionUC.ProcessAppleServerNotification(c.Request.Context(), req.SignedPayload); err != nil {
		log.Printf("Failed to process Apple server notification: %v", err)
		// Возвращаем 200, чтобы Apple не повторял запрос, но логируем ошибку
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Failed to process notification"})
		return
	}

	log.Printf("Successfully processed Apple server notification")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GET /apple/subscription-status (дополнительный endpoint для проверки статуса)
func (h *AppleHandler) GetSubscriptionStatus(c *gin.Context) {
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

	subscription, err := h.subscriptionUC.GetUserSubscription(c.Request.Context(), firebaseUID)
	if err != nil {
		log.Printf("Failed to get user subscription: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subscription})
}
