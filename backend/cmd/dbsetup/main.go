package main

import (
	"log"
	"os"

	"mendian-backend/internal/config"
	"mendian-backend/internal/database"
	"mendian-backend/internal/seed"
)

func main() {
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}
	if err := database.CreateDatabase(cfg.Database); err != nil {
		log.Fatalf("create database: %v", err)
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		log.Fatalf("migrate database tables: %v", err)
	}
	if err := seed.EnsureDefaultMenu(db); err != nil {
		log.Fatalf("seed default merchant menu: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get database connection: %v", err)
	}
	for _, table := range []string{"merchants", "menus", "menu_series", "products", "product_options", "product_option_values", "cart_items", "payment_orders"} {
		if !db.Migrator().HasTable(table) {
			log.Fatalf("expected table %q was not created", table)
		}
		log.Printf("verified table: %s", table)
	}
	if err := sqlDB.Close(); err != nil {
		log.Fatalf("close database connection: %v", err)
	}

	log.Println("database and tables initialized successfully")
}
