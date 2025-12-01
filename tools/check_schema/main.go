package main

import (
	"fmt"
	"log"

	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Check schema
	fmt.Println("📋 ReviewResult Schema:")
	fmt.Println("========================")
	
	var columns []struct {
		Name string
		Type string
	}
	
	db.Raw("SELECT name, type FROM pragma_table_info('review_results')").Scan(&columns)
	
	for _, col := range columns {
		fmt.Printf("  %-25s %s\n", col.Name, col.Type)
	}

	fmt.Println("\n✅ Schema check completed!")
}
