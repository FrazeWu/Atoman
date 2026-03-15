package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"atoman/internal/model"
)

func main() {
	// Load env: prefer .env, fallback to .env.dev
	if err := godotenv.Load(".env"); err != nil {
		if err2 := godotenv.Load(".env.dev"); err2 != nil {
			log.Println("No .env file found, using system environment variables")
		} else {
			log.Println("Loaded .env.dev")
		}
	} else {
		log.Println("Loaded .env")
	}

	dbType := os.Getenv("DATABASE_TYPE")
	dbURL := os.Getenv("DATABASE_URL")
	if dbType == "" || dbURL == "" {
		log.Fatal("DATABASE_TYPE and DATABASE_URL must be set")
	}

	var dialector gorm.Dialector
	switch dbType {
	case "postgres", "postgresql":
		dialector = postgres.Open(dbURL)
	case "sqlite":
		dialector = sqlite.Open(dbURL)
	default:
		log.Fatalf("Unsupported DATABASE_TYPE: %s (expected: postgres or sqlite)", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	username := prompt(scanner, "Username: ")
	email := prompt(scanner, "Email: ")
	password := prompt(scanner, "Password: ")

	if len(password) < 6 {
		log.Fatal("Password must be at least 6 characters")
	}

	// Check if user already exists
	var existing model.User
	if db.Where("username = ? OR email = ?", username, email).First(&existing).Error == nil {
		// User exists — update role to admin
		if err := db.Model(&existing).Update("role", "admin").Error; err != nil {
			log.Fatalf("Failed to update user role: %v", err)
		}
		fmt.Printf("User '%s' already exists. Role updated to admin.\n", existing.Username)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	user := model.User{
		Username: username,
		Email:    email,
		Password: string(hashed),
		Role:     "admin",
		IsActive: true,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	// Create default user settings
	_ = db.Create(&model.UserSettings{UserID: user.UUID})

	fmt.Printf("Admin user '%s' created successfully.\n", username)
}

func prompt(scanner *bufio.Scanner, label string) string {
	fmt.Print(label)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
