package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/food-delivery-app/db"
	"github.com/food-delivery-app/handlers"
	"github.com/food-delivery-app/routes"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

func main() {
	// Логирование
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Загрузка переменных окружения
	err := godotenv.Load()
	if err != nil {
		slog.Error("Ошибка загрузки .env файла", "error", err)
		log.Fatal("Ошибка загрузки .env файла")
	}

	// Инициализация базы данных
	database, err := db.InitDB()
	if err != nil {
		slog.Error("Не удалось подключиться к базе данных", "error", err)
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}
	defer database.Close()

	// Инициализация приложения
	app := &handlers.App{DB: database}

	// Инициализация роутера Gin
	r := gin.Default()

	// Подключение middleware
	r.Use(csrf.Protect([]byte(os.Getenv("CSRF_SECRET")), csrf.Secure(false))) // Secure(false) для разработки

	// Загрузка HTML-шаблонов
	r.LoadHTMLGlob("templates/*")

	// Подключение статических файлов
	r.Static("/static", "./static")

	// Инициализация маршрутов
	routes.SetupRoutes(r, app)

	// Запуск сервера
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
