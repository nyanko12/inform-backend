package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nyanko/inform-backend/internal/middleware"
	"github.com/nyanko/inform-backend/internal/models"
	"github.com/nyanko/inform-backend/internal/repository"
)

type CalendarHandler struct {
	calendarRepo *repository.CalendarRepository
	userRepo     *repository.UserRepository
}

func NewCalendarHandler(calendarRepo *repository.CalendarRepository, userRepo *repository.UserRepository) *CalendarHandler {
	return &CalendarHandler{calendarRepo: calendarRepo, userRepo: userRepo}
}

func (h *CalendarHandler) GetMonth(c *gin.Context) {
	userID := middleware.GetUserID(c)

	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}
	month, err := strconv.Atoi(c.Query("month"))
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
		return
	}

	notificationDaysBefore := 7
	if user, err := h.userRepo.FindByID(userID); err == nil && user != nil {
		notificationDaysBefore = user.NotificationDaysBefore
	}

	dates, err := h.calendarRepo.GetMonthData(userID, year, month, notificationDaysBefore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dates == nil {
		dates = []models.CalendarDate{}
	}
	c.JSON(http.StatusOK, gin.H{"dates": dates})
}
