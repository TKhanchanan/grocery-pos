package main

import (
	"log"

	"grocery-pos/apps/api/internal/config"
	"grocery-pos/apps/api/internal/database"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := database.ApplyFirstExistingSQLFile(db, "migrations/011_dynamic_rbac.sql", "apps/api/migrations/011_dynamic_rbac.sql"); err != nil {
		log.Fatalf("dynamic RBAC migration failed: %v", err)
	}
	log.Println("dynamic RBAC migration applied")
}
