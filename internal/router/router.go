package router

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nyanko/inform-backend/internal/handlers"
	"github.com/nyanko/inform-backend/internal/middleware"
	"github.com/nyanko/inform-backend/internal/repository"
)

func Setup(
	authHandler *handlers.AuthHandler,
	itemsHandler *handlers.ItemsHandler,
	productsHandler *handlers.ProductsHandler,
	calendarHandler *handlers.CalendarHandler,
	settingsHandler *handlers.SettingsHandler,
	userRepo *repository.UserRepository,
	firebaseProjectID, firebaseServiceAccountJSON string,
) *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Dev-User-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})


	r.GET("/myip", func(c *gin.Context) {
		resp, err := http.Get("https://api.ipify.org?format=json")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		c.JSON(200, result)
	})

	r.GET("/api/v1/image-proxy", func(c *gin.Context) {
		imgURL := c.Query("url")
		if imgURL == "" {
			c.Status(400)
			return
		}
		resp, err := http.Get(imgURL)
		if err != nil || resp.StatusCode != 200 {
			c.Status(404)
			return
		}
		defer resp.Body.Close()
		c.Header("Cache-Control", "public, max-age=86400")
		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	v1 := r.Group("/api/v1")

	// Public
	v1.POST("/auth/register", authHandler.Register)

	// Protected
	authMW := middleware.FirebaseAuth(firebaseProjectID, firebaseServiceAccountJSON, userRepo)
	protected := v1.Group("/", authMW)
	{
		protected.GET("/items/search", itemsHandler.Search)
		protected.GET("/items/barcode", itemsHandler.SearchByBarcode)

		protected.GET("/products", productsHandler.List)
		protected.POST("/products", productsHandler.Create)
		protected.POST("/products/calculate-days", productsHandler.CalculateDays)
		protected.DELETE("/products/:id", productsHandler.Delete)
		protected.PATCH("/products/:id/days", productsHandler.UpdateDays)
		protected.POST("/products/:id/purchase", productsHandler.Purchase)

		protected.GET("/calendar", calendarHandler.GetMonth)

		protected.PUT("/settings", settingsHandler.Update)
		protected.PUT("/fcm-token", settingsHandler.UpdateFCMToken)
		protected.DELETE("/settings/products/expired", settingsHandler.DeleteExpired)
	}

	return r
}
