package config

import "os"

type Config struct {
	AppName     string
	AppVersion  string
	Env         string
	APIAddr     string
	DatabaseURL string
	CORSOrigins string
	JWTSecret   string
}

func Load() Config {
	return Config{
		AppName:     env("APP_NAME", "Grocery POS & Inventory System"),
		AppVersion:  env("APP_VERSION", "0.1.0"),
		Env:         env("APP_ENV", "development"),
		APIAddr:     env("API_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "grocery:password@tcp(127.0.0.1:3306)/grocery_pos?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"),
		CORSOrigins: env("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"),
		JWTSecret:   env("JWT_SECRET", "dev-change-me"),
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
