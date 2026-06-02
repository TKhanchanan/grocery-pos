package main

import (
	"log"
	"net/http"

	"grocery-pos/backend/internal/config"
	"grocery-pos/backend/internal/db"
	api "grocery-pos/backend/internal/http"
	"grocery-pos/backend/internal/repo"
	"grocery-pos/backend/internal/service"
)

func main() {
	cfg := config.Load()
	conn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer conn.Close()

	repository := repo.New(conn)
	services := service.New(repository, cfg)
	server := api.NewServer(services, cfg)

	log.Printf("Grocery POS API listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
