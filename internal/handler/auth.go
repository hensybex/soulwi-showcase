// internal/handler/auth_handler.go (исправленная версия)

package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
	"github.com/hensybex/soulwi_go_back/internal/utils"
	"gorm.io/gorm"
)

type AuthHandler struct {
	JWTSecret    string
	FirebaseAuth *auth.Client
	UserRepo     repository.UserRepository
}

func NewAuthHandler(
	secret string,
	firebaseAuth *auth.Client,
	userRepo repository.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		JWTSecret:    secret,
		FirebaseAuth: firebaseAuth,
		UserRepo:     userRepo,
	}
}

// GET /auth/test-token
func (h *AuthHandler) GetTestToken(c *gin.Context) {
	token, err := utils.GenerateJWT(1, "admin", h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Alias for backward compatibility
func (h *AuthHandler) GenerateTestToken(c *gin.Context) {
	h.GetTestToken(c)
}

// POST /auth/register-test-user
func (h *AuthHandler) RegisterTestUser(c *gin.Context) {
	if h.FirebaseAuth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Firebase auth is not configured"})
		return
	}

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Create user in Firebase
	user, err := h.FirebaseAuth.CreateUser(c.Request.Context(), (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test user created successfully", "uid": user.UID})
}

// POST /auth/get-fb-token
func (h *AuthHandler) GetFirebaseToken(c *gin.Context) {
	if h.FirebaseAuth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Firebase auth is not configured"})
		return
	}

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password"` // Optional if you want to simulate login with password
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Fetch the user by email
	user, err := h.FirebaseAuth.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in Firebase", "details": err.Error()})
		return
	}

	// Generate a custom token for the user
	customToken, err := h.FirebaseAuth.CustomToken(c.Request.Context(), user.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate custom token", "details": err.Error()})
		return
	}

	// Exchange the custom token for an ID token using Firebase REST API.
	firebaseAPIKey := os.Getenv("FIREBASE_WEB_API_KEY")
	if firebaseAPIKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "FIREBASE_WEB_API_KEY is not configured"})
		return
	}
	apiURL := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=" + firebaseAPIKey

	payload := map[string]string{
		"token":             customToken,
		"returnSecureToken": "true",
	}

	resp, err := utils.MakeJSONRequest("POST", apiURL, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange custom token for ID token", "details": err.Error()})
		return
	}

	idToken, ok := resp["idToken"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve ID token", "details": resp})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ID token generated successfully", "id_token": idToken})
}

// Структуры для Apple Sign In
type AppleSignInRequest struct {
	IdentityToken string `json:"identityToken" binding:"required"`
}

type AppleSignInResponse struct {
	Success      bool   `json:"success"`
	SessionToken string `json:"sessionToken,omitempty"`
	Message      string `json:"message,omitempty"`
}

// POST /auth/apple-signin
func (h *AuthHandler) AppleSignIn(c *gin.Context) {
	if h.FirebaseAuth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Firebase auth is not configured"})
		return
	}

	var req AppleSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Верифицируем Apple Identity Token через Firebase Admin SDK
	token, err := h.FirebaseAuth.VerifyIDToken(context.Background(), req.IdentityToken)
	if err != nil {
		// Если Firebase не может верифицировать токен напрямую,
		// пробуем создать пользователя через Firebase Auth
		customToken, err2 := h.createFirebaseUserFromAppleToken(req.IdentityToken)
		if err2 != nil {
			c.JSON(http.StatusUnauthorized, AppleSignInResponse{
				Success: false,
				Message: "Failed to verify Apple identity token",
			})
			return
		}

		// Обменяем custom token на ID token
		// В реальном приложении клиент должен сделать это сам
		c.JSON(http.StatusOK, AppleSignInResponse{
			Success:      true,
			SessionToken: customToken,
			Message:      "Apple sign-in successful",
		})
		return
	}

	// Если токен уже валиден, создаем сессионный токен
	sessionToken, err := h.createSessionToken(token.UID, "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, AppleSignInResponse{
			Success: false,
			Message: "Failed to create session token",
		})
		return
	}

	c.JSON(http.StatusOK, AppleSignInResponse{
		Success:      true,
		SessionToken: sessionToken,
		Message:      "Apple sign-in successful",
	})
}

// Helper method для создания пользователя Firebase из Apple token
func (h *AuthHandler) createFirebaseUserFromAppleToken(identityToken string) (string, error) {
	if h.FirebaseAuth == nil {
		return "", fmt.Errorf("firebase auth is not configured")
	}

	// Парсим Apple identity token без валидации (для получения claims)
	// В production нужно валидировать подпись Apple
	token, _, err := new(jwt.Parser).ParseUnverified(identityToken, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	// Извлекаем информацию из Apple token
	sub, _ := claims["sub"].(string) // Apple User ID
	email, _ := claims["email"].(string)
	emailVerified, _ := claims["email_verified"].(bool)

	if sub == "" {
		return "", fmt.Errorf("missing sub claim")
	}

	// Создаем или обновляем пользователя в Firebase
	userToCreate := (&auth.UserToCreate{}).UID(sub)

	if email != "" {
		userToCreate = userToCreate.Email(email).EmailVerified(emailVerified)
	}

	// Пытаемся создать пользователя
	userRecord, err := h.FirebaseAuth.CreateUser(context.Background(), userToCreate)
	if err != nil {
		// Если пользователь уже существует, получаем его
		userRecord, err = h.FirebaseAuth.GetUser(context.Background(), sub)
		if err != nil {
			return "", err
		}
	}

	// Создаем custom token для этого пользователя
	customToken, err := h.FirebaseAuth.CustomToken(context.Background(), userRecord.UID)
	if err != nil {
		return "", err
	}

	return customToken, nil
}

// Helper method для создания сессионного токена
func (h *AuthHandler) createSessionToken(userID, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // 24 часа
	})

	return token.SignedString([]byte(h.JWTSecret))
}

// НОВЫЙ МЕТОД: Полный тестовый логин для тестера
// Требует ?secret=test123 в query (глупая защита).
// Создаёт/логинит пользователя, генерирует ID-токен, sync'ит профиль в БД.
func (h *AuthHandler) TestFullLogin(c *gin.Context) {
	if h.FirebaseAuth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Firebase auth is not configured"})
		return
	}

	// Глупая защита: Только тестер знает secret
	secret := c.Query("secret")
	if secret != "test123" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: invalid secret"})
		return
	}

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name,omitempty"` // Опционально
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	ctx := c.Request.Context()

	// 1. Создать пользователя в Firebase (если нет)
	user, err := h.FirebaseAuth.GetUserByEmail(ctx, req.Email)
	if err != nil || user == nil {
		// Создаём нового
		user, err = h.FirebaseAuth.CreateUser(ctx, (&auth.UserToCreate{}).
			Email(req.Email).
			Password(req.Password),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
			return
		}
	}

	// 2. Сгенерировать custom token
	customToken, err := h.FirebaseAuth.CustomToken(ctx, user.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate custom token", "details": err.Error()})
		return
	}

	// 3. Обменять custom token на ID token (используя API key)
	firebaseAPIKey := os.Getenv("FIREBASE_WEB_API_KEY")
	if firebaseAPIKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "FIREBASE_WEB_API_KEY is not configured"})
		return
	}
	apiURL := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=" + firebaseAPIKey

	payload := map[string]string{
		"token":             customToken,
		"returnSecureToken": "true",
	}

	resp, err := utils.MakeJSONRequest("POST", apiURL, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token", "details": err.Error()})
		return
	}

	idToken, ok := resp["idToken"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve ID token", "details": resp})
		return
	}

	// 4. Inline-sync пользователя в БД (простая логика без UserHandler/usecase)
	now := time.Now()
	testUser := &model.User{
		FirebaseUID:          user.UID,
		Email:                req.Email,
		Name:                 req.Name,
		LoginType:            "email",
		TimezoneOffset:       0, // Дефолт
		LastSeenAt:           &now,
		NotificationsEnabled: true, // Default
	}

	// Проверяем существующего пользователя
	existingUser, err := h.UserRepo.GetByFirebaseUID(ctx, testUser.FirebaseUID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Printf("[ERROR] Failed to check for existing user %s: %v\n", testUser.FirebaseUID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
		return
	}

	if existingUser != nil {
		// Update existing
		fmt.Printf("[INFO] Updating existing user %s\n", existingUser.FirebaseUID)
		existingUser.Email = testUser.Email
		existingUser.Name = testUser.Name
		existingUser.TimezoneOffset = testUser.TimezoneOffset
		existingUser.LoginType = testUser.LoginType
		existingUser.LastSeenAt = &now
		if err := h.UserRepo.Update(ctx, existingUser); err != nil {
			fmt.Printf("[ERROR] Failed to update user %s: %v\n", existingUser.FirebaseUID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		testUser = existingUser // Для ответа
	} else {
		// Create new
		fmt.Printf("[INFO] Creating new user %s\n", testUser.FirebaseUID)
		if err := h.UserRepo.Create(ctx, testUser); err != nil {
			fmt.Printf("[ERROR] Failed to create user %s: %v\n", testUser.FirebaseUID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	}

	fmt.Printf("[SUCCESS] User %s successfully created/updated during test login.\n", testUser.FirebaseUID)

	// 5. Вернуть готовый токен + UID
	c.JSON(http.StatusOK, gin.H{
		"message":      "Test full login successful",
		"id_token":     idToken, // Bearer для Postman
		"firebase_uid": user.UID,
		"role":         "user",   // Для Firebase-токена
		"user_data":    testUser, // Полный профиль для дебага
	})
}
