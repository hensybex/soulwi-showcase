// internal/handler/user_handler.go

package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type UserHandler struct {
	userUC       usecase.UserUsecase
	FirebaseAuth *firebaseAuth.Client
}

func NewUserHandler(userUC usecase.UserUsecase, fbAuthClient *firebaseAuth.Client) *UserHandler {
	return &UserHandler{
		userUC:       userUC,
		FirebaseAuth: fbAuthClient,
	}
}

// POST /users
func (h *UserHandler) RegisterOrUpdateUser(c *gin.Context) {
	if h.FirebaseAuth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Firebase auth is not configured"})
		return
	}

	log.Println("INFO: Received request for RegisterOrUpdateUser")
	var user model.User

	// Make a copy of the request body to be able to log it in case of an error
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore the body

	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v. Request Body: %s", err, string(bodyBytes))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	log.Printf("INFO: Successfully bound JSON body to user struct: %+v", user)

	// Must have a valid Firebase token in the context
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		log.Println("ERROR: Missing firebase_uid in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok || firebaseUID == "" {
		log.Printf("ERROR: Invalid firebase_uid format in context. Value: %v", firebaseUIDVal)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}
	log.Printf("INFO: Extracted firebase_uid: %s", firebaseUID)

	// We always overwrite with the token's UID to ensure integrity
	user.FirebaseUID = firebaseUID
	log.Printf("[SYNC] Start uid=%s prev=%s loginType=%s email_set=%t name_set=%t tz=%d",
		user.FirebaseUID,
		user.PreviousFirebaseUID,
		user.LoginType,
		user.Email != "",
		user.Name != "",
		user.TimezoneOffset)

	// 1) Get the full user record from Firebase to check status (e.g., anonymous)
	firebaseRecord, err := h.FirebaseAuth.GetUser(c.Request.Context(), firebaseUID)
	if err != nil {
		log.Printf("ERROR: Failed to get user record from Firebase with UID %s: %v", firebaseUID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user with Firebase"})
		return
	}

	// Check if the user is anonymous
	// Note: After linking, an account is no longer anonymous, even if it started that way.
	isAnonymous := firebaseRecord.Email == "" && len(firebaseRecord.ProviderUserInfo) == 0
	log.Printf("INFO: User check complete. UID: %s, IsAnonymous: %t", firebaseUID, isAnonymous)

	if isAnonymous {
		user.LoginType = "anonymous"
		user.Email = "" // Explicitly clear email for anonymous users
	} else {
		// For regular users, ensure email is from the verified token
		user.LoginType = "email" // Or determine from ProviderUserInfo if needed
		user.Email = firebaseRecord.Email

		// Fallback: If name is not in the request body, use DisplayName from Firebase token
		if user.Name == "" && firebaseRecord.DisplayName != "" {
			user.Name = firebaseRecord.DisplayName
			log.Printf("INFO: Using DisplayName from Firebase token as user name: %s", user.Name)
		}
	}

	log.Printf("INFO: Final user object before upserting to DB: %+v", user)

	// 2) Now, perform the upsert logic with the verified and consolidated user data
	registeredUser, err := h.userUC.RegisterOrUpdateUser(c.Request.Context(), &user)
	if err != nil {
		log.Printf("ERROR: Failed during user upsert in use case: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register or update user"})
		return
	}

	log.Printf("[SYNC] Done uid=%s id=%d loginType=%s email_set=%t tz=%d",
		registeredUser.FirebaseUID,
		registeredUser.ID,
		registeredUser.LoginType,
		registeredUser.Email != "",
		registeredUser.TimezoneOffset)
	log.Printf("SUCCESS: User %s successfully registered/updated.", registeredUser.FirebaseUID)
	c.JSON(http.StatusOK, gin.H{"data": registeredUser})
}

// GET /users/me
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// Get Firebase UID from the authenticated user
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		log.Println("[ERROR] firebase_uid missing in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}

	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		log.Println("[ERROR] Invalid firebase_uid format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}

	user, err := h.userUC.GetUserByFirebaseUID(c.Request.Context(), firebaseUID)
	if err != nil {
		log.Printf("[ERROR] User not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// --- НОВЫЙ МЕТОД ---
type fcmTokenRequest struct {
	FCMToken string `json:"fcm_token" binding:"required"`
}

// POST /users/fcm-token
func (h *UserHandler) UpdateFCMToken(c *gin.Context) {
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format in context"})
		return
	}

	var req fcmTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error})
		return
	}

	err := h.userUC.UpdateFCMToken(c.Request.Context(), firebaseUID, req.FCMToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update FCM token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "FCM token updated successfully"})
}

// DELETE /users/me
func (h *UserHandler) DeleteCurrentUser(c *gin.Context) {
	uidVal, ok := c.Get("firebase_uid")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	uid, _ := uidVal.(string)

	// Purge user data, then soft-delete the user row.
	if err := h.userUC.DeleteUser(c.Request.Context(), uid); err != nil {
		log.Printf("[ERROR] Delete user failed for %s: %v", uid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
