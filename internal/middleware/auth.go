package middleware

import (
	"backend/pkg/auth"
	"backend/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthRequired validates the JWT token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.JSONError(c, http.StatusUnauthorized, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.JSONError(c, http.StatusUnauthorized, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			response.JSONError(c, http.StatusUnauthorized, "invalid or expired token", err.Error())
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

// RoleRequired checks if the authenticated user has one of the allowed roles
func RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			response.JSONError(c, http.StatusUnauthorized, "user role not found in context")
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		response.JSONError(c, http.StatusForbidden, "you don't have permission to access this resource")
		c.Abort()
	}
}
