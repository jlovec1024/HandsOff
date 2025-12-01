package main

import (
	"fmt"
	"log"

	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/pkg/config"
	"github.com/handsoff/handsoff/pkg/database"
)

func main() {
	fmt.Println("🔧 Initializing database...")

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

	fmt.Println("✅ Connected to database")

	// Auto migrate
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("✅ Database tables created")

	// Create default admin user
	var existingUser model.User
	if err := db.Where("username = ?", "admin").First(&existingUser).Error; err == nil {
		fmt.Println("ℹ️  Admin user already exists, skipping...")
		fmt.Printf("   Username: %s\n", existingUser.Username)
		fmt.Printf("   Email: %s\n", existingUser.Email)
		return
	}

	admin := model.User{
		Username: "admin",
		Password: "admin123", // Will be hashed by BeforeCreate hook
		Email:    "admin@handsoff.local",
		IsActive: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Println("✅ Created default admin user:")
	fmt.Printf("   Username: %s\n", admin.Username)
	fmt.Printf("   Password: admin123\n")
	fmt.Printf("   Email: %s\n", admin.Email)
	fmt.Println("\n⚠️  Please change the default password after first login!")
	fmt.Println("\n🎉 Database initialization complete!")
}
