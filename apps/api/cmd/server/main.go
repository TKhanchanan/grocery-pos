package main

import (
	"log"
	"net/http"

	"grocery-pos/apps/api/internal/config"
	"grocery-pos/apps/api/internal/database"
	"grocery-pos/apps/api/internal/httpx"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	server := httpx.NewServer(cfg, db)

	log.Printf("Grocery POS API %s listening on %s", cfg.AppVersion, cfg.APIAddr)
	if err := http.ListenAndServe(cfg.APIAddr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
