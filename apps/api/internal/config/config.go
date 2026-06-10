package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const defaultUploadDir = "storage/uploads"

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
	appEnv := env("APP_ENV", "development")
	return Config{
		AppName:     env("APP_NAME", "Grocery POS & Inventory System"),
		AppVersion:  env("APP_VERSION", "0.1.0"),
		Env:         appEnv,
		Port:        os.Getenv("PORT"),
		APIAddr:     os.Getenv("API_ADDR"),
		DatabaseURL: env("DATABASE_URL", "grocery:password@tcp(127.0.0.1:3306)/grocery_pos?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"),
		CORSOrigins: env("CORS_ORIGINS", "https://grocery-pos-front-production.up.railway.app,http://localhost:5173"),
		JWTSecret:   env("JWT_SECRET", "dev-change-me"),
		UploadDir:   uploadDir(appEnv, os.Getenv("UPLOAD_DIR")),
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

func (c Config) PrepareUploadDir() error {
	if strings.TrimSpace(c.UploadDir) == "" {
		return errors.New("UPLOAD_DIR is required in production")
	}
	for _, dir := range []string{
		c.UploadDir,
		filepath.Join(c.UploadDir, "products"),
		filepath.Join(c.UploadDir, "avatars"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func uploadDir(appEnv, value string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return filepath.Clean(value)
	}
	if strings.EqualFold(strings.TrimSpace(appEnv), "production") {
		return ""
	}
	return defaultUploadDir
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
