package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nyanko/inform-backend/internal/repository"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
}

func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		FirebaseUID string `json:"firebase_uid" binding:"required"`
		Email       string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.userRepo.FindByFirebaseUID(req.FirebaseUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing != nil {
		c.JSON(http.StatusOK, gin.H{"user_id": existing.ID})
		return
	}

	user, err := h.userRepo.Create(req.FirebaseUID, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user_id": user.ID})
}
