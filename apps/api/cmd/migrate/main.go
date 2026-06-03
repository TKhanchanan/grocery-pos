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

	for _, migration := range []string{"011_dynamic_rbac.sql", "012_profile_avatar.sql"} {
		if err := database.ApplyFirstExistingSQLFile(db, "migrations/"+migration, "apps/api/migrations/"+migration); err != nil {
			log.Fatalf("%s migration failed: %v", migration, err)
		}
	}
	log.Println("migrations applied")
}
