package routes

import (
	"net/http"

	"github.com/food-delivery-app/handlers"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
)

func SetupRoutes(r *gin.Engine, app *handlers.App) {
	// Главная страница
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"CSRFToken": csrf.Token(c),
		})
	})

	// Страницы без авторизации
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"CSRFToken": csrf.Token(c),
		})
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"CSRFToken": csrf.Token(c),
		})
	})
	r.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{
			"CSRFToken": csrf.Token(c),
		})
	})
	r.GET("/faq", func(c *gin.Context) {
		c.HTML(http.StatusOK, "faq.html", gin.H{
			"CSRFToken": csrf.Token(c),
		})
	})

	// API для авторизации
	r.POST("/api/login", app.HandleLogin)
	r.POST("/api/register", app.HandleRegister)
	r.POST("/api/logout", app.HandleLogout)
	r.GET("/api/session", app.HandleSession)

	// Страницы и API для клиентов
	customerRoutes := r.Group("/").Use(middleware.AuthMiddleware(app.DB, "customer"))
	{
		customerRoutes.GET("/restaurants", func(c *gin.Context) {
			c.HTML(http.StatusOK, "restaurants.html", gin.H{
				"CSRFToken": csrf.Token(c),
			})
		})
		customerRoutes.GET("/menu", func(c *gin.Context) {
			c.HTML(http.StatusOK, "menu.html", gin.H{
				"CSRFToken": csrf.Token(c),
			})
		})
		customerRoutes.GET("/cart", func(c *gin.Context) {
			c.HTML(http.StatusOK, "cart.html", gin.H{
				"CSRFToken": csrf.Token(c),
			})
		})
		customerRoutes.GET("/my-orders", func(c *gin.Context) {
			c.HTML(http.StatusOK, "my-orders.html", gin.H{
				"CSRFToken": csrf.Token(c),
			})
		})
		customerRoutes.GET("/order-status/:id", func(c *gin.Context) {
			c.HTML(http.StatusOK, "order-status.html", gin.H{
				"CSRFToken": csrf.Token(c),
			})
		})

		// API
		customerRoutes.GET("/api/restaurants", app.GetRestaurants)
		customerRoutes.GET("/api/menu", app.GetMenu)
		customerRoutes.POST("/api/cart", app.AddToCart)
		customerRoutes.PUT("/api/cart", app.UpdateCart)
		customerRoutes.DELETE("/api/cart", app.RemoveFromCart)
		customerRoutes.GET("/api/orders", app.GetOrders)
		customerRoutes.POST("/api/order", app.CreateOrder)
		customerRoutes.GET("/api/order/:id", app.GetOrder)
	}

	// Страницы и API для ресторанов
	restaurantRoutes := r.Group("/").Use(middleware.AuthMiddleware(app.DB, "restaurant"))
	{
		restaurantRoutes.GET("/restaurant-orders", func(c *gin.Context) {
			c.HTML(http.StatusOK, "restaurant-orders.html", gin.H{
				"CSRFToken": csrf.Token(c),
			})
		})
		restaurantRoutes.GET("/add-menu", func(c *gin.Context) {
			c.HTML(http.StatusOK, "add-menu.html", gin.H{
				"CSRFToken": csrf.Token(c),
			})
		})
		restaurantRoutes.GET("/manage-menu", func(c *gin.Context) {
			c.HTML(http.StatusOK, "manage-menu.html", gin.H{
				"CSRFToken": csrf.Token(c),
			})
		})

		// API
		restaurantRoutes.POST("/api/menu", app.AddMenuItem)
		restaurantRoutes.PUT("/api/menu", app.UpdateMenuItem)
		restaurantRoutes.DELETE("/api/menu", app.DeleteMenuItem)
		restaurantRoutes.GET("/api/restaurant/orders", app.GetRestaurantOrders)
		restaurantRoutes.PUT("/api/restaurant/orders", app.UpdateOrderStatus)
	}
}
