// src/middleware/jwtAuth.go
package middleware

import (
	"net/http"
	"publicPost/src/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func JWTAuthMiddleware(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		userID, err := ValidateToken(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)

		c.Next()
	}
}

func ValidateToken(token string, secret string) (string, error) {

	return "1", nil
}
