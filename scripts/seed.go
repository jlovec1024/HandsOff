package main

import (
	"fmt"
	"log"

	"github.com/handsoff/handsoff/internal/model"
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

	// Create default admin user
	admin := model.User{
		Username: "admin",
		Password: "admin123", // Will be hashed by BeforeCreate hook
		Email:    "admin@handsoff.local",
		IsActive: true,
	}

	// Check if admin already exists
	var existing model.User
	if err := db.Where("username = ?", admin.Username).First(&existing).Error; err == nil {
		fmt.Println("Admin user already exists, skipping...")
		return
	}

	// Create admin user
	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("✅ Created default admin user:\n")
	fmt.Printf("   Username: %s\n", admin.Username)
	fmt.Printf("   Password: admin123\n")
	fmt.Printf("   Email: %s\n", admin.Email)
	fmt.Println("\n⚠️  Please change the default password after first login!")
}
