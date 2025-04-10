package routes

import (
	"database/sql"
	"fmt"
	"food-delivery-app/internal/delivery"
	"food-delivery-app/internal/middleware"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
)

func SetupRoutes(r *gin.Engine, app *delivery.App) {
	// Публичные страницы
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
	})
	r.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
	})
	r.GET("/faq", func(c *gin.Context) {
		c.HTML(http.StatusOK, "faq.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
	})

	// Авторизация
	r.POST("/api/login", app.HandleLogin)
	r.POST("/api/register", app.HandleRegister)
	r.POST("/api/logout", app.HandleLogout)
	r.GET("/api/session", app.HandleSession)

	// Роуты клиента
	customerRoutes := r.Group("/").Use(middleware.AuthMiddleware(app.DB, "customer"))
	{
		customerRoutes.GET("/homepage", func(c *gin.Context) {
			c.HTML(http.StatusOK, "homepage.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
		})
		customerRoutes.GET("/restaurants", func(c *gin.Context) {
			c.HTML(http.StatusOK, "restaurants.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
		})
		customerRoutes.GET("/menu", func(c *gin.Context) {
			c.HTML(http.StatusOK, "menu.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
		})
		customerRoutes.GET("/cart", func(c *gin.Context) {
			c.HTML(http.StatusOK, "cart.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
		})
		customerRoutes.GET("/my-orders", func(c *gin.Context) {
			c.HTML(http.StatusOK, "my-orders.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
		})
		customerRoutes.GET("/order-status/:id", func(c *gin.Context) {
			c.HTML(http.StatusOK, "order-status.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
		})
		customerRoutes.GET("/product-details/:id", func(c *gin.Context) {
			itemID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"error": "Неверный ID блюда"})
				return
			}
			var menuItem struct {
				ID           int
				Name         string
				Description  sql.NullString
				Price        float64
				ImageURL     sql.NullString
				RestaurantID int
			}
			err = app.DB.QueryRow(`
				SELECT id, name, description, price, image_url, restaurant_id 
				FROM menu 
				WHERE id = $1`, itemID).
				Scan(
					&menuItem.ID,
					&menuItem.Name,
					&menuItem.Description,
					&menuItem.Price,
					&menuItem.ImageURL,
					&menuItem.RestaurantID, // ← добавляем это
				)
			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(404, gin.H{"error": "Блюдо не найдено"})
					return
				}
				c.JSON(500, gin.H{"error": "Не удалось загрузить блюдо"})
				return
			}
			c.HTML(http.StatusOK, "product-details.html", gin.H{
				"CSRFToken": csrf.Token(c.Request),
				"MenuItem":  menuItem,
			})
		})

		// API для получения меню
		customerRoutes.GET("/api/menu", func(c *gin.Context) {
			restaurantID := c.Query("restaurant_id")
			if restaurantID == "" {
				c.JSON(400, gin.H{"error": "ID ресторана не указан"})
				return
			}

			var restaurant struct {
				ID          int            `db:"id"`
				Name        string         `db:"name"`
				Address     sql.NullString `db:"address"`
				CuisineType sql.NullString `db:"cuisine_type"`
			}
			err := app.DB.QueryRow(`
				SELECT id, name, address, cuisine_type
				FROM restaurants
				WHERE id = $1
			`, restaurantID).Scan(&restaurant.ID, &restaurant.Name, &restaurant.Address, &restaurant.CuisineType)
			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(404, gin.H{"error": "Ресторан не найден"})
					return
				}
				log.Printf("Ошибка при получении ресторана: %v", err)
				c.JSON(500, gin.H{"error": fmt.Sprintf("Не удалось загрузить ресторан: %v", err)})
				return
			}

			var menuItems []struct {
				ID           int            `db:"id" json:"id"`
				RestaurantID int            `db:"restaurant_id" json:"restaurant_id"`
				Name         string         `db:"name" json:"name"`
				Description  sql.NullString `db:"description" json:"description"`
				Price        float64        `db:"price" json:"price"`
				ImageURL     sql.NullString `db:"image_url" json:"image_url"`
			}
			err = app.DB.Select(&menuItems, `
				SELECT id, restaurant_id, name, description, price, image_url
				FROM menu
				WHERE restaurant_id = $1
			`, restaurantID)
			if err != nil {
				log.Printf("Ошибка при получении меню: %v", err)
				c.JSON(500, gin.H{"error": fmt.Sprintf("Не удалось загрузить меню: %v", err)})
				return
			}

			c.JSON(200, gin.H{
				"restaurant": restaurant,
				"menu_items": menuItems,
			})
		})

		customerRoutes.GET("/api/recommended-dishes", func(c *gin.Context) {
			var dishes []struct {
				ID       int            `json:"id"`
				Name     string         `json:"name"`
				Price    float64        `json:"price"`
				ImageURL sql.NullString `json:"image_url"`
			}
			rows, err := app.DB.Query(`
				SELECT id, name, price, image_url
				FROM menu
				ORDER BY RANDOM()
				LIMIT 4
			`)
			if err != nil {
				log.Printf("Ошибка при получении рекомендованных блюд: %v", err)
				c.JSON(500, gin.H{"error": "Не удалось загрузить рекомендованные блюда"})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var dish struct {
					ID       int            `json:"id"`
					Name     string         `json:"name"`
					Price    float64        `json:"price"`
					ImageURL sql.NullString `json:"image_url"`
				}
				if err := rows.Scan(&dish.ID, &dish.Name, &dish.Price, &dish.ImageURL); err != nil {
					log.Printf("Ошибка при сканировании блюда: %v", err)
					c.JSON(500, gin.H{"error": "Не удалось загрузить блюда"})
					return
				}
				dishes = append(dishes, dish)
			}

			c.JSON(200, dishes)
		})

		// API клиента
		customerRoutes.GET("/api/restaurants", app.GetRestaurants)
		customerRoutes.POST("/api/checkout", app.Checkout)
		customerRoutes.PUT("/api/cart", app.UpdateCart)
		customerRoutes.GET("/api/cart", app.GetCart)
		customerRoutes.POST("/api/cart/add", app.AddToCart)
		customerRoutes.DELETE("/api/cart/:id", app.RemoveFromCart)
		customerRoutes.PATCH("/api/cart/:id", app.UpdateCartItem)
		customerRoutes.GET("/api/orders", app.GetOrders)
		customerRoutes.POST("/api/order", app.CreateOrder)
		customerRoutes.GET("/api/order/:id", app.GetOrder)
	}

	// Роуты ресторана
	restaurantRoutes := r.Group("/").Use(middleware.AuthMiddleware(app.DB, "restaurant"))
	{
		restaurantRoutes.GET("/restaurant-orders", func(c *gin.Context) {
			c.HTML(http.StatusOK, "restaurant-orders.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
		})
		restaurantRoutes.GET("/add-menu", func(c *gin.Context) {
			c.HTML(http.StatusOK, "add-menu.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
		})
		restaurantRoutes.GET("/manage-menu", func(c *gin.Context) {
			c.HTML(http.StatusOK, "manage-menu.html", gin.H{"CSRFToken": csrf.Token(c.Request)})
		})

		// API ресторана
		restaurantRoutes.GET("/api/restaurants/user", app.GetUserRestaurants)
		restaurantRoutes.POST("/api/menu", app.AddMenuItem)
		restaurantRoutes.PUT("/api/menu", app.UpdateMenuItem)
		restaurantRoutes.DELETE("/api/menu", app.DeleteMenuItem)
		restaurantRoutes.GET("/api/menu-restaurants", app.GetMenuItems)
		restaurantRoutes.GET("/api/restaurant/orders", app.GetRestaurantOrders)
		restaurantRoutes.PUT("/api/restaurant/orders", app.UpdateOrderStatus)
	}
}
