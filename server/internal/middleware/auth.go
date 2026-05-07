package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func jwtSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func parseAuthToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret(), nil
	})
}

func setAuthContext(c *gin.Context, claims jwt.MapClaims) bool {
	var userIDStr string
	switch v := claims["user_id"].(type) {
	case string:
		userIDStr = v
	case float64:
		userIDStr = fmt.Sprintf("%v", v)
	default:
		return false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return false
	}

	role, _ := claims["role"].(string)
	if role == "" {
		role = "user"
	}

	c.Set("user_id", userID)
	c.Set("username", claims["username"])
	c.Set("role", role)
	return true
}

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := parseAuthToken(tokenString)

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !setAuthContext(c, claims) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT if present, but does not block if missing
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.Next()
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := parseAuthToken(tokenString)

		if err != nil {
			log.Printf("[Auth] JWT parse error: %v", err)
			c.Next()
			return
		}

		if !token.Valid {
			c.Next()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && !setAuthContext(c, claims) {
			log.Printf("[Auth] Invalid token claims")
		}

		c.Next()
	}
}
