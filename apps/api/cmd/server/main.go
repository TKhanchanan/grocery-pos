package main

import (
	"log"
	"net/http"
	"strings"

	"grocery-pos/apps/api/internal/config"
	"grocery-pos/apps/api/internal/database"
	"grocery-pos/apps/api/internal/httpx"
)

func main() {
	cfg := config.Load()
	if err := cfg.PrepareUploadDir(); err != nil {
		log.Fatalf("prepare upload directory: %v", err)
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	server := httpx.NewServer(cfg, db)
	listenAddr := cfg.ListenAddr()

	log.Printf("APP_ENV=%s", cfg.Env)
	log.Printf("PORT set=%t", strings.TrimSpace(cfg.Port) != "")
	log.Printf("CORS_ORIGINS=%s", cfg.CORSOrigins)
	log.Printf("upload dir: %s", cfg.UploadDir)
	log.Printf("API base path=/api/v1")
	log.Printf("Grocery POS API %s listening on %s", cfg.AppVersion, listenAddr)
	if err := http.ListenAndServe(listenAddr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
