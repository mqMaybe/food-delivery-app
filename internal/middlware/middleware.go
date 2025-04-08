package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func AuthMiddleware(db *sql.DB, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
			c.Abort()
			return
		}

		var userID int
		var role string
		err = db.QueryRow("SELECT user_id, role FROM sessions WHERE session_id = $1", sessionID).Scan(&userID, &role)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверная сессия"})
			c.Abort()
			return
		}

		if requiredRole != "" && role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}
