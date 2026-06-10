package config

import (
	"os"
	"strings"
)

type Config struct {
	AppName     string
	AppVersion  string
	Env         string
	Port        string
	APIAddr     string
	DatabaseURL string
	CORSOrigins string
	JWTSecret   string
	UploadDir   string
}

func Load() Config {
	return Config{
		AppName:     env("APP_NAME", "Grocery POS & Inventory System"),
		AppVersion:  env("APP_VERSION", "0.1.0"),
		Env:         env("APP_ENV", "development"),
		Port:        os.Getenv("PORT"),
		APIAddr:     os.Getenv("API_ADDR"),
		DatabaseURL: env("DATABASE_URL", "grocery:password@tcp(127.0.0.1:3306)/grocery_pos?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"),
		CORSOrigins: env("CORS_ORIGINS", "https://grocery-pos-front-production.up.railway.app,http://localhost:5173"),
		JWTSecret:   env("JWT_SECRET", "dev-change-me"),
		UploadDir:   env("UPLOAD_DIR", "storage/uploads"),
	}
}

func (c Config) ListenAddr() string {
	if port := strings.TrimSpace(c.Port); port != "" {
		return ":" + strings.TrimPrefix(port, ":")
	}
	if addr := strings.TrimSpace(c.APIAddr); addr != "" {
		return addr
	}
	return ":8080"
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
