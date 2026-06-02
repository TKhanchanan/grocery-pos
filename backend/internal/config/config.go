package config

import "os"

type Config struct {
	Addr        string
	DatabaseURL string
	JWTSecret   string
	LineToken   string
}

func Load() Config {
	return Config{
		Addr:        env("API_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "grocery:password@tcp(127.0.0.1:3306)/grocery_pos?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"),
		JWTSecret:   env("JWT_SECRET", "dev-change-me"),
		LineToken:   os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
