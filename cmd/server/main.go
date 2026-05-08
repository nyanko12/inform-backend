package main

import (
	"log"
	"path/filepath"
	"runtime"
	"time"
	_ "time/tzdata"

	"github.com/nyanko/inform-backend/internal/config"
	"github.com/nyanko/inform-backend/internal/db"
	"github.com/nyanko/inform-backend/internal/handlers"
	"github.com/nyanko/inform-backend/internal/repository"
	"github.com/nyanko/inform-backend/internal/router"
	"github.com/nyanko/inform-backend/internal/services"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

database, err := db.Connect(cfg.DBDSN)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	migrationsDir := filepath.Join(projectRoot, "internal", "db", "migrations")

	if err := db.RunMigrations(database, migrationsDir); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations applied")

	userRepo := repository.NewUserRepository(database)
	productRepo := repository.NewProductRepository(database)
	calendarRepo := repository.NewCalendarRepository(database)

	rakutenSvc := services.NewRakutenService(cfg.RakutenApplicationID)
	calcSvc := services.NewCalculatorService(database)
	notifSvc := services.NewNotificationService(productRepo, cfg.FirebaseServiceAccountJSON, cfg.FirebaseProjectID)

	authH := handlers.NewAuthHandler(userRepo)
	itemsH := handlers.NewItemsHandler(rakutenSvc)
	productsH := handlers.NewProductsHandler(productRepo, userRepo, calcSvc)
	calendarH := handlers.NewCalendarHandler(calendarRepo, userRepo)
	settingsH := handlers.NewSettingsHandler(userRepo, productRepo)

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("[cron] Asia/Tokyo not found, falling back to UTC: %v", err)
		jst = time.UTC
	} else {
		log.Println("[cron] timezone: Asia/Tokyo")
	}
	c := cron.New(cron.WithLocation(jst))
	c.AddFunc("0 8 * * *", notifSvc.SendDailyNotifications)
	c.Start()
	log.Printf("[cron] daily notification scheduled at 08:00 %s", jst)
	defer c.Stop()

	r := router.Setup(authH, itemsH, productsH, calendarH, settingsH, userRepo, cfg.FirebaseProjectID, cfg.FirebaseServiceAccountJSON, notifSvc.SendDailyNotifications)

	log.Printf("server listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
