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

	for _, migration := range []string{
		"001_init.sql",
		"002_catalog_compat.sql",
		"003_stock_movements_before_after.sql",
		"004_sales_pos_compat.sql",
		"005_sales_cancel_compat.sql",
		"006_alerts_read_compat.sql",
		"007_purchase_orders_sent_compat.sql",
		"008_purchasing_compat.sql",
		"009_purchase_orders_compat.sql",
		"010_settings_line_defaults.sql",
		"011_dynamic_rbac.sql",
		"012_profile_avatar.sql",
		"013_product_images.sql",
	} {
		if err := database.ApplyFirstExistingSQLFile(db, "migrations/"+migration, "apps/api/migrations/"+migration); err != nil {
			log.Fatalf("%s migration failed: %v", migration, err)
		}
	}
	log.Println("migrations applied")
}
