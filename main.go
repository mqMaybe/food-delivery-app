package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"food-delivery-app/internal/db"
	"food-delivery-app/internal/delivery"
	"food-delivery-app/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	err := godotenv.Load()
	if err != nil {
		slog.Error("Ошибка загрузки .env файла", "error", err)
		log.Fatal("Ошибка загрузки .env файла")
	}

	database, err := db.InitDB()
	if err != nil {
		slog.Error("Не удалось подключиться к базе данных", "error", err)
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}
	defer database.Close()

	app := &delivery.App{DB: database}
	r := gin.Default()

	csrfMiddleware := csrf.Protect(
		[]byte(os.Getenv("CSRF_SECRET")),
		csrf.Secure(false),
		csrf.Path("/"),
	)

	r.Use(func(c *gin.Context) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Writer = &csrfResponseWriter{c.Writer, w}
			c.Request = r
			c.Next()
		})
		csrfMiddleware(handler).ServeHTTP(c.Writer, c.Request)
	})

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	routes.SetupRoutes(r, app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("Сервер запущен", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Не удалось запустить сервер", "error", err)
		log.Fatal("Не удалось запустить сервер:", err)
	}
}

type csrfResponseWriter struct {
	gin.ResponseWriter
	w http.ResponseWriter
}

func (crw *csrfResponseWriter) WriteHeader(code int) {
	crw.w.WriteHeader(code)
	crw.ResponseWriter.WriteHeader(code)
}

func (crw *csrfResponseWriter) Write(b []byte) (int, error) {
	return crw.w.Write(b)
}
