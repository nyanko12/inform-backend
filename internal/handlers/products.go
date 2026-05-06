package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nyanko/inform-backend/internal/middleware"
	"github.com/nyanko/inform-backend/internal/models"
	"github.com/nyanko/inform-backend/internal/repository"
	"github.com/nyanko/inform-backend/internal/services"
)

type ProductsHandler struct {
	productRepo *repository.ProductRepository
	userRepo    *repository.UserRepository
	calculator  *services.CalculatorService
}

func NewProductsHandler(productRepo *repository.ProductRepository, userRepo *repository.UserRepository, calc *services.CalculatorService) *ProductsHandler {
	return &ProductsHandler{productRepo: productRepo, userRepo: userRepo, calculator: calc}
}

func (h *ProductsHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	keyword := c.Query("keyword")

	notificationDaysBefore := 7
	if user, err := h.userRepo.FindByID(userID); err == nil && user != nil {
		notificationDaysBefore = user.NotificationDaysBefore
	}

	products, err := h.productRepo.ListByUser(userID, keyword, notificationDaysBefore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if products == nil {
		products = []*models.ProductWithStatus{}
	}
	c.JSON(http.StatusOK, gin.H{"products": products})
}

func (h *ProductsHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		ItemCode      string   `json:"item_code"`
		Name          string   `json:"name" binding:"required"`
		Genre         string   `json:"genre"`
		ImageURL      string   `json:"image_url"`
		ContentVolume *float64 `json:"content_volume"`
		ContentUnit   string   `json:"content_unit"`
		DaysToConsume int      `json:"days_to_consume" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := &models.Product{
		Name:          req.Name,
		DaysToConsume: req.DaysToConsume,
	}
	if req.ItemCode != "" {
		p.ItemCode = &req.ItemCode
	}
	if req.Genre != "" {
		p.Genre = &req.Genre
	}
	if req.ImageURL != "" {
		p.ImageURL = &req.ImageURL
	}
	if req.ContentVolume != nil {
		p.ContentVolume = req.ContentVolume
	}
	if req.ContentUnit != "" {
		p.ContentUnit = &req.ContentUnit
	}

	created, err := h.productRepo.Create(userID, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":            created.ID,
		"next_due_date": created.NextDueDate.Format("2006-01-02"),
	})
}

func (h *ProductsHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	productID := c.Param("id")

	if err := h.productRepo.SoftDelete(productID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProductsHandler) Purchase(c *gin.Context) {
	userID := middleware.GetUserID(c)
	productID := c.Param("id")

	newDueDate, err := h.productRepo.Purchase(productID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"new_due_date": newDueDate.Format("2006-01-02")})
}

func (h *ProductsHandler) UpdateDays(c *gin.Context) {
	userID := middleware.GetUserID(c)
	productID := c.Param("id")
	var req struct {
		DaysToConsume int `json:"days_to_consume" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.productRepo.UpdateDaysToConsume(productID, userID, req.DaysToConsume); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *ProductsHandler) CalculateDays(c *gin.Context) {
	var req struct {
		Genre               string   `json:"genre" binding:"required"`
		ContentVolume       float64  `json:"content_volume" binding:"required"`
		ContentUnit         string   `json:"content_unit"`
		NumPeople           int      `json:"num_people" binding:"required,min=1"`
		DailyUsagePerPerson *float64 `json:"daily_usage_per_person"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usage, days, err := h.calculator.Calculate(req.Genre, req.ContentVolume, req.NumPeople, req.DailyUsagePerPerson, req.ContentUnit)
	if err != nil {
		if strings.HasPrefix(err.Error(), "genre not found") {
			c.JSON(http.StatusOK, gin.H{"genre_not_found": true})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"days_to_consume":        days,
		"daily_usage_per_person": usage.DailyUsagePerPerson,
	})
}
