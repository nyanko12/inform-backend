package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nyanko/inform-backend/internal/middleware"
	"github.com/nyanko/inform-backend/internal/repository"
)

type SettingsHandler struct {
	userRepo    *repository.UserRepository
	productRepo *repository.ProductRepository
}

func NewSettingsHandler(userRepo *repository.UserRepository, productRepo *repository.ProductRepository) *SettingsHandler {
	return &SettingsHandler{userRepo: userRepo, productRepo: productRepo}
}

func (h *SettingsHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		NotificationDaysBefore int `json:"notification_days_before" binding:"required,min=1,max=14"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userRepo.UpdateSettings(userID, req.NotificationDaysBefore); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *SettingsHandler) DeleteExpired(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.productRepo.DeleteExpiredByUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
