package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/nyanko/inform-backend/internal/repository"
	"google.golang.org/api/option"
)

const userIDKey = "userID"

func FirebaseAuth(projectID, serviceAccountJSON string, userRepo *repository.UserRepository) gin.HandlerFunc {
	if projectID == "" {
		log.Println("[auth] Firebase not configured, using dev bypass mode")
		return devBypassAuth(userRepo)
	}

	// serviceAccountJSONがJSONの中身ならWithCredentialsJSON、ファイルパスならWithCredentialsFile
	var opt option.ClientOption
	if len(serviceAccountJSON) > 0 && serviceAccountJSON[0] == '{' {
		opt = option.WithCredentialsJSON([]byte(serviceAccountJSON))
	} else {
		opt = option.WithCredentialsFile(serviceAccountJSON)
	}

	app, err := firebase.NewApp(context.Background(), &firebase.Config{ProjectID: projectID}, opt)
	if err != nil {
		log.Printf("[auth] Firebase init error: %v — falling back to dev bypass", err)
		return devBypassAuth(userRepo)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Printf("[auth] Auth client error: %v — falling back to dev bypass", err)
		return devBypassAuth(userRepo)
	}

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		idToken := strings.TrimPrefix(header, "Bearer ")

		token, err := authClient.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		user, err := userRepo.FindByFirebaseUID(token.UID)
		if err != nil || user == nil {
			email, _ := token.Claims["email"].(string)
			user, err = userRepo.Create(token.UID, email)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user creation failed"})
				return
			}
		}
		c.Set(userIDKey, user.ID)
		c.Next()
	}
}

func devBypassAuth(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("X-Dev-User-ID")
		if header != "" {
			c.Set(userIDKey, header)
			c.Next()
			return
		}
		// Create/use a default dev user
		user, _ := userRepo.FindByFirebaseUID("dev-uid")
		if user == nil {
			user, _ = userRepo.Create("dev-uid", "dev@example.com")
		}
		if user != nil {
			c.Set(userIDKey, user.ID)
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	id, _ := c.Get(userIDKey)
	if s, ok := id.(string); ok {
		return s
	}
	return ""
}
