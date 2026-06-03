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

	for _, migration := range []string{"011_dynamic_rbac.sql", "012_profile_avatar.sql"} {
		if err := database.ApplyFirstExistingSQLFile(db, "migrations/"+migration, "apps/api/migrations/"+migration); err != nil {
			log.Printf("%s migration skipped: %v", migration, err)
		}
	}

	server := httpx.NewServer(cfg, db)

	log.Printf("Grocery POS API %s listening on %s", cfg.AppVersion, cfg.APIAddr)
	if err := http.ListenAndServe(cfg.APIAddr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
