// internal/transport/middleware/firebase_auth.go

package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type FirebaseAuthMiddleware struct {
	authClient *auth.Client
}

// NewFirebaseAuthMiddleware loads your Firebase service account JSON for user auth
func NewFirebaseAuthMiddleware(credsFile string) (*FirebaseAuthMiddleware, error) {
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(credsFile))
	if err != nil {
		return nil, err
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}
	return &FirebaseAuthMiddleware{authClient: client}, nil
}

func (f *FirebaseAuthMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("Missing Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Println("Invalid Authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
			return
		}

		tokenStr := parts[1]
		token, err := f.authClient.VerifyIDToken(context.Background(), tokenStr)
		if err != nil {
			log.Printf("Token verification failed: %v\n", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("firebase_uid", token.UID)

		c.Next()
	}
}
