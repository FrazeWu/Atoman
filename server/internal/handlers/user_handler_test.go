package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"atoman/internal/model"
)

func newUserHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Follow{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func createActiveUser(t *testing.T, db *gorm.DB, username string) model.User {
	t.Helper()
	user := model.User{
		Username: username,
		Email:    username + "@example.com",
		Password: "secret",
		IsActive: true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return user
}

func authToken(t *testing.T, user model.User) string {
	t.Helper()
	secret := "test-secret"
	t.Setenv("JWT_SECRET", secret)
	claims := jwt.MapClaims{
		"user_id":  user.UUID.String(),
		"username": user.Username,
		"role":     "user",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func setupSearchUsersRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	users := r.Group("/api/users")
	users.GET("/search", func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			})
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if userIDStr, ok := claims["user_id"].(string); ok {
						if parsed, err := uuid.Parse(userIDStr); err == nil {
							c.Set("user_id", parsed)
						}
					}
				}
			}
		}
		SearchUsers(db)(c)
	})
	return r
}

func TestSearchUsersMentionScopeReturnsFollowersNotFollowing(t *testing.T) {
	db := newUserHandlerTestDB(t)
	current := createActiveUser(t, db, "current")
	follower := createActiveUser(t, db, "ingfollower")
	following := createActiveUser(t, db, "following")
	_ = createActiveUser(t, db, "outsider")

	if err := db.Create(&model.Follow{FollowerID: follower.UUID, FollowingID: current.UUID}).Error; err != nil {
		t.Fatalf("create follower edge: %v", err)
	}
	if err := db.Create(&model.Follow{FollowerID: current.UUID, FollowingID: following.UUID}).Error; err != nil {
		t.Fatalf("create following edge: %v", err)
	}

	router := setupSearchUsersRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/api/users/search?scope=mention&q=ing&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+authToken(t, current))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data []struct {
			Username string `json:"username"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].Username != "ingfollower" {
		t.Fatalf("expected only follower, got %+v", resp.Data)
	}
}

func TestSearchUsersMentionScopeUnauthenticatedReturnsEmptyArray(t *testing.T) {
	db := newUserHandlerTestDB(t)
	current := createActiveUser(t, db, "current")
	follower := createActiveUser(t, db, "follower")
	if err := db.Create(&model.Follow{FollowerID: follower.UUID, FollowingID: current.UUID}).Error; err != nil {
		t.Fatalf("create follower edge: %v", err)
	}

	router := setupSearchUsersRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/api/users/search?scope=mention&q=fol&limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected empty array, got %+v", resp.Data)
	}
}
