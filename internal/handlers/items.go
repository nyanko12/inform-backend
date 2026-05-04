package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nyanko/inform-backend/internal/services"
)

type ItemsHandler struct {
	rakuten *services.RakutenService
}

func NewItemsHandler(rakuten *services.RakutenService) *ItemsHandler {
	return &ItemsHandler{rakuten: rakuten}
}

func (h *ItemsHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keyword is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	hits, _ := strconv.Atoi(c.DefaultQuery("hits", "10"))
	if hits != 10 && hits != 20 && hits != 30 {
		hits = 10
	}

	result, err := h.rakuten.Search(keyword, page, hits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ItemsHandler) SearchByBarcode(c *gin.Context) {
	janCode := c.Query("jan_code")
	if janCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "jan_code is required"})
		return
	}
	result, err := h.rakuten.SearchByJAN(janCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
