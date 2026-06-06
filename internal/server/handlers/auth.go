package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"seasons/backend/gen/dbstore"
	"seasons/backend/internal/server/middleware"
)

const sessionDuration = 30 * 24 * time.Hour

type AuthHandler struct {
	queries dbstore.Querier
}

func NewAuthHandler(queries dbstore.Querier) *AuthHandler {
	return &AuthHandler{queries: queries}
}

type loginRequest struct {
	Login    string `json:"login"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

type userResponse struct {
	ID         string           `json:"id"`
	Login      string           `json:"login"`
	FirstName  string           `json:"first_name"`
	LastName   string           `json:"last_name"`
	MiddleName string           `json:"middle_name"`
	Role       dbstore.UserRole `json:"role"`
	Language   string           `json:"language"`
}

type loginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      userResponse `json:"user"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.queries.GetUserByLogin(c.Request.Context(), sql.NullString{
		String: req.Login,
		Valid:  true,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !user.Password.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !user.Active {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is inactive"})
		return
	}

	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	expiresAt := time.Now().Add(sessionDuration)
	session, err := h.queries.CreateSession(c.Request.Context(), dbstore.CreateSessionParams{
		UserID:    user.ID,
		Token:     token,
		IpAddress: sql.NullString{String: c.ClientIP(), Valid: true},
		UserAgent: sql.NullString{String: c.Request.UserAgent(), Valid: true},
		ExpiresAt: expiresAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	_ = h.queries.UpdateUserLastOnline(c.Request.Context(), user.ID)

	c.JSON(http.StatusOK, loginResponse{
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
		User:      toUserResponse(user),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := middleware.CurrentToken(c)
	if err := h.queries.DeleteSession(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	c.JSON(http.StatusOK, toUserResponse(user))
}

func toUserResponse(u dbstore.User) userResponse {
	login := ""
	if u.Login.Valid {
		login = u.Login.String
	}
	return userResponse{
		ID:         u.ID.String(),
		Login:      login,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: u.MiddleName,
		Role:       u.Role,
		Language:   u.Language,
	}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
