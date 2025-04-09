package middleware

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// CORSMiddleware разрешает запросы с любых источников (CORS) — для dev-режима
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

// AuthMiddleware проверяет наличие и валидность сессии пользователя, а также его роль
func AuthMiddleware(db *sqlx.DB, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			if c.Request.URL.Path[:4] == "/api/" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		var userID int
		var role string
		err = db.QueryRow("SELECT user_id, role FROM sessions WHERE session_id = $1", sessionID).Scan(&userID, &role)
		if err != nil {
			if err == sql.ErrNoRows {
				if c.Request.URL.Path[:4] == "/api/" {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Сессия недействительна"})
				} else {
					c.Redirect(http.StatusFound, "/login")
				}
			} else {
				log.Printf("Ошибка при проверке сессии: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
			}
			c.Abort()
			return
		}

		if role != requiredRole {
			if c.Request.URL.Path[:4] == "/api/" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
