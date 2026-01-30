package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User definitions to match the main application
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;column:id"`
	Username  string    `json:"username" gorm:"unique;not null;column:username"`
	Email     string    `json:"email" gorm:"unique;not null;column:email"`
	Password  string    `json:"-" gorm:"not null;column:password"`
	Role      string    `json:"role" gorm:"default:'user';column:role"`
	CreatedAt time.Time `json:"created_at" gorm:"column:createdAt"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updatedAt"`
}

func (User) TableName() string {
	return "Users"
}

func main() {
	// Parse flags
	username := flag.String("username", "", "Admin username")
	email := flag.String("email", "", "Admin email")
	password := flag.String("password", "", "Admin password")
	promote := flag.String("promote", "", "Existing username to promote to admin")
	flag.Parse()

	// Load env
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: .env file not found, trying current directory")
		godotenv.Load()
	}

	// Connect to DB
	dbPath := "../../database.sqlite"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbPath = "database.sqlite"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate to ensure Role column exists
	db.AutoMigrate(&User{})

	if *promote != "" {
		// Promote existing user
		var user User
		if err := db.Where("username = ?", *promote).First(&user).Error; err != nil {
			log.Fatalf("User %s not found", *promote)
		}

		user.Role = "admin"
		if err := db.Save(&user).Error; err != nil {
			log.Fatal("Failed to update user role:", err)
		}
		fmt.Printf("Successfully promoted %s to admin\n", *promote)
		return
	}

	if *username == "" || *email == "" || *password == "" {
		fmt.Println("Usage:")
		fmt.Println("  Create new admin: go run . -username <name> -email <email> -password <pass>")
		fmt.Println("  Promote existing: go run . -promote <username>")
		os.Exit(1)
	}

	// Create new admin
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	user := User{
		Username: *username,
		Email:    *email,
		Password: string(hashedPassword),
		Role:     "admin",
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	fmt.Printf("Successfully created admin user %s\n", *username)
}
