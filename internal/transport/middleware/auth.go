// internal/transport/middleware/auth.go

package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// We will decode the JWT's "role" claim if found
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

// AuthMiddleware tries (in order):
//  1. Parse a JWT from the "Authorization: Bearer ..." header.
//  2. If JWT fails, parse a Firebase token using the same header.
//
// If a Firebase token for an email/password user is used, it now ENFORCES
// that the user's email has been verified.
//
// Sets c.Set("role", either from JWT or "user" if Firebase token).
func AuthMiddleware(jwtSecret string, fbClient *firebaseAuth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Println("[AUTH] Missing or invalid Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 1) Try JWT (for internal/admin tokens)
		roleExtracted, err := tryParseJWT(tokenString, jwtSecret)
		if err == nil {
			// success => roleExtracted is e.g. "admin" or "user"
			c.Set("role", roleExtracted)
			// Note: We are assuming JWTs are for trusted internal services and don't need UID.
			c.Next()
			return
		} else {
			log.Printf("[AUTH] JWT parse/verify failed: %v", err)
		}

		// 2) If JWT fails, try Firebase
		if fbClient != nil {
			token, errFB := fbClient.VerifyIDToken(context.Background(), tokenString)
			if errFB == nil && token != nil {
				// --- NEW: EMAIL VERIFICATION ENFORCEMENT ---
				// Check if the sign-in provider is email/password
				isPasswordProvider := false
				if signInProvider, ok := token.Claims["sign_in_provider"].(string); ok {
					isPasswordProvider = signInProvider == "password"
				}

				// If it's an email/password user, check if their email is verified.
				if isPasswordProvider {
					emailVerified, ok := token.Claims["email_verified"].(bool)
					if !ok || !emailVerified {
						log.Printf("[AUTH-DENIED] User %s attempted login with unverified email.", token.UID)
						c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "email_not_verified"})
						return
					}
				}
				// --- END OF NEW LOGIC ---

				// success => default role is "user"
				c.Set("role", "user")
				c.Set("firebase_uid", token.UID)
				c.Next()
				return
			}

			if errFB != nil {
				log.Printf("[AUTH] Firebase token verification failed: %v", errFB)
			}
		}

		// neither JWT nor Firebase worked
		log.Println("[AUTH] Authorization failed after checking JWT and Firebase")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}
}

// tryParseJWT tries to parse JWT. Returns role (string) or error
func tryParseJWT(tokenString, secret string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid JWT")
	}
	if claims.Role == "" {
		// default fallback
		return "user", nil
	}
	return claims.Role, nil
}

// RequireRoleMiddleware ensures the user's role is in the allowed list
// If not, returns 403 Forbidden
func RequireRoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: missing role"})
			return
		}
		role, _ := roleVal.(string)

		for _, r := range allowedRoles {
			if r == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	}
}
