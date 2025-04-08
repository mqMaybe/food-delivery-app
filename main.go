package main

import (
	"food-delivery-app/internal/config"
	"food-delivery-app/internal/db"
	"food-delivery-app/internal/delivery"
	"food-delivery-app/internal/log"
	"food-delivery-app/internal/usecase"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logger := log.NewLogger()
		logger.Fatal("Failed to load config", err)
		return
	}

	// Инициализация логгера
	logger := log.NewLogger()

	// Подключение к базе данных
	database, err := db.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", err)
		return
	}

	// Инициализация use case и delivery
	uc := usecase.NewUseCase(database)
	d := delivery.NewDelivery(uc, logger)

	// Настройка Gin
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")

	// Группа маршрутов для страниц с проверкой авторизации
	r.GET("/", d.RenderPage("index.html"))
	r.GET("/login", d.RenderPage("login.html"))
	r.GET("/register", d.RenderPage("register.html"))
	r.GET("/faq", d.RenderPage("faq.html"))
	r.GET("/about", d.RenderPage("about.html"))
	r.GET("/restaurants", d.RenderPageWithAuth("restaurants.html", "customer"))
	r.GET("/cart", d.RenderPageWithAuth("cart.html", "customer"))
	r.GET("/my-orders", d.RenderPageWithAuth("my-orders.html", "customer"))
	r.GET("/restaurant-orders", d.RenderPageWithAuth("restaurant-orders.html", "restaurant"))
	r.GET("/add-menu", d.RenderPageWithAuth("add-menu.html", "restaurant"))
	r.GET("/manage-menu", d.RenderPageWithAuth("manage-menu.html", "restaurant"))
	r.GET("/menu", d.RenderPageWithAuth("menu.html", "customer"))
	r.GET("/order-status/:order_id", d.RenderPageWithAuth("order-status.html", "customer"))

	// API эндпоинты
	r.POST("/api/register", d.RegisterUser)
	r.POST("/api/login", d.LoginUser)
	r.POST("/api/logout", d.LogoutUser)
	r.GET("/api/restaurants", d.GetRestaurants)
	r.POST("/api/menu", d.AddMenuItem)
	r.PUT("/api/menu", d.UpdateMenuItem)
	r.DELETE("/api/menu", d.DeleteMenuItem)
	r.GET("/api/menu", d.GetMenu)
	r.POST("/api/cart", d.AddToCart)
	r.GET("/api/cart", d.GetCart)
	r.PUT("/api/cart", d.UpdateCartItem)
	r.DELETE("/api/cart", d.DeleteCartItem)
	r.POST("/api/order", d.CreateOrder)
	r.GET("/api/restaurant/orders", d.GetRestaurantOrders)
	r.PUT("/api/restaurant/orders", d.UpdateOrderStatus)
	r.GET("/api/orders", d.GetUserOrders)
	r.GET("/api/session", d.GetSession)
	r.GET("/api/order/:order_id", d.GetUserOrders)

	// Запуск сервера
	logger.Info("Server starting on port " + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		logger.Fatal("Failed to start server", err)
	}
}
