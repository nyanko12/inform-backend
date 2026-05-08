package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBDSN      string

	RakutenApplicationID string

	FirebaseProjectID          string
	FirebaseServiceAccountJSON string

	Port string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "inform_user"),
		DBPassword: getEnv("DB_PASSWORD", "inform_pass"),
		DBName:     getEnv("DB_NAME", "inform_db"),

		RakutenApplicationID: getEnv("YAHOO_APPLICATION_ID", getEnv("YAHOO_APP_ID", getEnv("RAKUTEN_APPLICATION_ID", ""))),

		FirebaseProjectID:          getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseServiceAccountJSON: getEnv("FIREBASE_SERVICE_ACCOUNT_JSON", ""),

		Port: getEnv("PORT", "8080"),
	}

	// RailwayはDATABASE_URLで接続情報を渡す
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DBDSN = dbURL + "?sslmode=require"
	} else {
		cfg.DBDSN = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
		)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
