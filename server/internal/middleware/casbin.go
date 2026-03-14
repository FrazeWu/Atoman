package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

)

var Enforcer *casbin.Enforcer

// InitCasbin initializes the casbin enforcer with GORM adapter
func InitCasbin(db *gorm.DB) error {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return err
	}

	// Initialize the casbin model
	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
`)
	if err != nil {
		return err
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return err
	}

	Enforcer = enforcer

	// Load the policies from DB
	err = Enforcer.LoadPolicy()
	if err != nil {
		return err
	}

	// Add default policies if none exist
	initDefaultPolicies()

	return nil
}

func initDefaultPolicies() {
	// Anonymous policies
	Enforcer.AddPolicy("anonymous", "/api/auth/*", "POST")
	Enforcer.AddPolicy("anonymous", "/api/*", "GET")
	Enforcer.AddPolicy("anonymous", "/uploads/*", "(GET|HEAD)")

	// User policies (can do most things, ownership enforced by handlers)
	Enforcer.AddPolicy("user", "/api/*", "(GET|POST|PUT|DELETE)")
	Enforcer.AddPolicy("user", "/uploads/*", "(GET|HEAD)")

	// Admin policies
	Enforcer.AddPolicy("admin", "/api/*", "(GET|POST|PUT|DELETE)")
	Enforcer.AddPolicy("admin", "/uploads/*", "(GET|HEAD)")

	Enforcer.SavePolicy()
}

// CasbinMiddleware is the Gin middleware for Casbin authorization
func CasbinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := "anonymous"

		// Extract role from context if set by OptionalAuthMiddleware
		if roleVal, exists := c.Get("role"); exists {
			if r, ok := roleVal.(string); ok && r != "" {
				role = r
			}
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		// Special case: ignore OPTIONS
		if method == "OPTIONS" {
			c.Next()
			return
		}

		// Enforce
		allowed, err := Enforcer.Enforce(role, path, method)
		if err != nil {
			log.Printf("Casbin error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error checking permissions"})
			return
		}

		if !allowed {
			if role == "anonymous" && method != "GET" && !strings.HasPrefix(path, "/api/auth") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Login required to perform this action"})
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
			return
		}

		c.Next()
	}
}
