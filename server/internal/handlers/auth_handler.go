package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"atoman/internal/model"
	"atoman/internal/service"
)

const authTokenTTL = 30 * 24 * time.Hour

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func generateAuthToken(user model.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is not configured")
	}

	role := user.Role
	if role == "" {
		role = "user"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.UUID.String(),
		"username": user.Username,
		"role":     role,
		"exp":      time.Now().Add(authTokenTTL).Unix(),
	})

	return token.SignedString([]byte(secret))
}

// RegisterInput represents user registration request
type RegisterInput struct {
	Username         string `json:"username" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=6"`
	PasswordConfirm  string `json:"password_confirm" binding:"required,eqfield=Password"`
	VerificationCode string `json:"verification_code" binding:"required,len=6"`
}

// LoginInput represents user login request
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SendVerificationInput represents email verification code request
type SendVerificationInput struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyEmailInput represents email verification request
type VerifyEmailInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// SetupAuthRoutes configures authentication routes
func SetupAuthRoutes(router *gin.Engine, db *gorm.DB, emailService *service.EmailService) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", RegisterHandler(db, emailService))
		auth.POST("/login", LoginHandler(db))
		auth.POST("/send-verification", SendVerificationHandler(emailService))
		auth.POST("/verify-email", VerifyEmailHandler(emailService))
	}
}

// RegisterHandler handles user registration
func RegisterHandler(db *gorm.DB, emailService *service.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RegisterInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify email verification code first
		valid, err := emailService.VerifyCode(input.Email, input.VerificationCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email code"})
			return
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification code"})
			return
		}

		// Check if user exists
		var existingUser model.User
		if err := db.Where("username = ? OR email = ?", input.Username, input.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		user := model.User{
			Username: input.Username,
			Email:    input.Email,
			Password: string(hashedPassword),
			Role:     "user",
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		if err := db.Create(&model.UserSettings{UserID: user.UUID}).Error; err != nil {
			c.Error(err)
		}

		if _, err := EnsureDefaultChannelForUser(db, user.UUID, user.Username); err != nil {
			c.Error(err)
		}

		// Auto-subscribe to user's own channel feed
		go func() {
			// Give it a moment for the channel to be fully created
			time.Sleep(100 * time.Millisecond)

			var channel model.Channel
			if err := db.Where("user_id = ?", user.UUID).First(&channel).Error; err == nil {
				// Create subscription to own channel
				sourceHash := fmt.Sprintf("internal_channel:%s", channel.ID.String())
				h := sha256.New()
				h.Write([]byte(sourceHash))
				hash := hex.EncodeToString(h.Sum(nil))

				var source model.FeedSource
				if err := db.Where("hash = ?", hash).First(&source).Error; err != nil {
					source = model.FeedSource{
						SourceType: "internal_channel",
						SourceID:   &channel.ID,
						Title:      channel.Name,
						Hash:       hash,
					}
					db.Create(&source)
				}

				// Get or create default subscription group
				var group model.SubscriptionGroup
				if err := db.Where("user_id = ? AND name = ?", user.UUID, "默认分组").First(&group).Error; err != nil {
					group = model.SubscriptionGroup{
						UserID: user.UUID,
						Name:   "默认分组",
					}
					db.Create(&group)
				}

				// Create subscription
				sub := model.Subscription{
					UserID:              user.UUID,
					FeedSourceID:        source.ID,
					Title:               channel.Name,
					SubscriptionGroupID: &group.ID,
				}
				db.Create(&sub)
			}
		}()

		tokenString, err := generateAuthToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"token": tokenString,
			"user": gin.H{
				"uuid":         user.UUID,
				"id":           user.ID,
				"username":     user.Username,
				"email":        user.Email,
				"role":         user.Role,
				"display_name": user.DisplayName,
				"avatar_url":   user.AvatarURL,
				"is_active":    user.IsActive,
			},
		})
	}
}

// LoginHandler handles user login
func LoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user model.User
		if err := db.Where("username = ? OR email = ?", input.Username, input.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		if user.Role == "" {
			user.Role = "user"
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		tokenString, err := generateAuthToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"user": gin.H{
				"uuid":         user.UUID,
				"id":           user.ID,
				"username":     user.Username,
				"email":        user.Email,
				"role":         user.Role,
				"display_name": user.DisplayName,
				"avatar_url":   user.AvatarURL,
				"is_active":    user.IsActive,
			},
		})
	}
}

// SendVerificationHandler handles sending verification code
func SendVerificationHandler(emailService *service.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input SendVerificationInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Send verification code
		code, err := emailService.SendVerificationCode(input.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification code", "details": err.Error()})
			return
		}

		// Log the code in development mode for debugging
		if os.Getenv("GIN_MODE") != "release" {
			println("[DEBUG] Verification code for", input.Email, ":", code)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Verification code sent"})
	}
}

// VerifyEmailHandler handles email verification
func VerifyEmailHandler(emailService *service.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input VerifyEmailInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify the code
		valid, err := emailService.VerifyCode(input.Email, input.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify code"})
			return
		}

		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification code"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
	}
}
