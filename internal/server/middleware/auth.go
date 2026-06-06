package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"seasons/backend/gen/dbstore"
)

type contextKey string

const (
	userContextKey  contextKey = "user"
	tokenContextKey contextKey = "session_token"
)

func RequireAuth(queries dbstore.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ExtractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		session, err := queries.GetSessionByToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
			return
		}

		user, err := queries.GetUserByID(c.Request.Context(), session.UserID)
		if err != nil || !user.Active {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		go func() {
			_ = queries.UpdateSessionLastUsed(context.Background(), token)
		}()

		c.Set(string(userContextKey), user)
		c.Set(string(tokenContextKey), token)
		c.Next()
	}
}

func ExtractToken(c *gin.Context) string {
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	token, _ := c.Cookie("session_token")
	return token
}

func CurrentUser(c *gin.Context) (dbstore.User, bool) {
	val, exists := c.Get(string(userContextKey))
	if !exists {
		return dbstore.User{}, false
	}
	user, ok := val.(dbstore.User)
	return user, ok
}

func CurrentToken(c *gin.Context) string {
	token, _ := c.Get(string(tokenContextKey))
	s, _ := token.(string)
	return s
}
