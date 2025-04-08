package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/food-delivery-app/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	DB *sql.DB
}

func (app *App) HandleLogin(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	var user models.User
	err := app.DB.QueryRow("SELECT id, name, email, password, role, cuisine_type FROM users WHERE email = $1", input.Email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CuisineType)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
		return
	}

	sessionID, err := generateSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать сессию"})
		return
	}

	_, err = app.DB.Exec("INSERT INTO sessions (session_id, user_id, role, expires_at) VALUES ($1, $2, $3, $4)",
		sessionID, user.ID, user.Role, time.Now().Add(24*time.Hour))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось создать сессию: %v", err)})
		return
	}

	c.SetCookie("session_id", sessionID, 3600*24, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Вход выполнен успешно", "role": user.Role})
}

func (app *App) HandleRegister(c *gin.Context) {
	var input struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		Role        string `json:"role"`
		CuisineType string `json:"cuisine_type"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось хешировать пароль"})
		return
	}

	var userID int
	if input.Role == "restaurant" {
		err = app.DB.QueryRow("INSERT INTO users (name, email, password, role, cuisine_type) VALUES ($1, $2, $3, $4, $5) RETURNING id",
			input.Name, input.Email, hashedPassword, input.Role, input.CuisineType).Scan(&userID)
	} else {
		err = app.DB.QueryRow("INSERT INTO users (name, email, password, role) VALUES ($1, $2, $3, $4) RETURNING id",
			input.Name, input.Email, hashedPassword, input.Role).Scan(&userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось зарегистрировать пользователя: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Регистрация прошла успешно"})
}

func (app *App) HandleLogout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен"})
		return
	}

	_, err = app.DB.Exec("DELETE FROM sessions WHERE session_id = $1", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось удалить сессию: %v", err)})
		return
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен"})
}

func (app *App) HandleSession(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	var userID int
	var role string
	err = app.DB.QueryRow("SELECT user_id, role FROM sessions WHERE session_id = $1", sessionID).Scan(&userID, &role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверная сессия"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role})
}

func (app *App) GetRestaurants(c *gin.Context) {
	cuisineType := c.Query("cuisine_type")
	deliveryTime := c.Query("delivery_time")
	rating := c.Query("rating")

	query := "SELECT id, name, cuisine_type, delivery_time, rating FROM users WHERE role = 'restaurant'"
	conditions := []string{}
	args := []interface{}{}
	if cuisineType != "" && cuisineType != "all" {
		conditions = append(conditions, fmt.Sprintf("cuisine_type = $%d", len(args)+1))
		args = append(args, cuisineType)
	}
	if deliveryTime != "" && deliveryTime != "all" {
		conditions = append(conditions, fmt.Sprintf("delivery_time <= $%d", len(args)+1))
		args = append(args, deliveryTime)
	}
	if rating != "" && rating != "all" {
		conditions = append(conditions, fmt.Sprintf("rating >= $%d", len(args)+1))
		args = append(args, rating)
	}
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	rows, err := app.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось загрузить рестораны: %v", err)})
		return
	}
	defer rows.Close()

	var restaurants []models.Restaurant
	for rows.Next() {
		var r models.Restaurant
		if err := rows.Scan(&r.ID, &r.Name, &r.CuisineType, &r.DeliveryTime, &r.Rating); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		restaurants = append(restaurants, r)
	}

	c.JSON(http.StatusOK, restaurants)
}

func (app *App) GetMenu(c *gin.Context) {
	restaurantID := c.Query("restaurant_id")
	if restaurantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ресторана обязателен"})
		return
	}

	rows, err := app.DB.Query("SELECT id, restaurant_id, name, price, description FROM menu WHERE restaurant_id = $1", restaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось загрузить меню: %v", err)})
		return
	}
	defer rows.Close()

	var menuItems []models.MenuItem
	for rows.Next() {
		var item models.MenuItem
		if err := rows.Scan(&item.ID, &item.RestaurantID, &item.Name, &item.Price, &item.Description); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		menuItems = append(menuItems, item)
	}

	c.JSON(http.StatusOK, menuItems)
}

func (app *App) AddToCart(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		MenuID   int `json:"menu_id"`
		Quantity int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	_, err := app.DB.Exec("INSERT INTO cart (user_id, menu_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (user_id, menu_id) DO UPDATE SET quantity = cart.quantity + $3",
		userID, input.MenuID, input.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось добавить в корзину: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Добавлено в корзину"})
}

func (app *App) UpdateCart(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		CartID   int `json:"cart_id"`
		Quantity int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	_, err := app.DB.Exec("UPDATE cart SET quantity = $1 WHERE id = $2 AND user_id = $3", input.Quantity, input.CartID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось обновить корзину: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Корзина обновлена"})
}

func (app *App) RemoveFromCart(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		CartID int `json:"cart_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	_, err := app.DB.Exec("DELETE FROM cart WHERE id = $1 AND user_id = $2", input.CartID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось удалить из корзины: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Удалено из корзины"})
}

func (app *App) GetOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	rows, err := app.DB.Query(`
		SELECT o.id, o.user_id, o.restaurant_id, o.delivery_address, o.total_price, o.status
		FROM orders o
		WHERE o.user_id = $1`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось загрузить заказы: %v", err)})
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.RestaurantID, &o.DeliveryAddress, &o.TotalPrice, &o.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
			return
		}

		itemRows, err := app.DB.Query(`
			SELECT c.id, c.user_id, c.menu_id, c.quantity, m.name, m.price, m.description
			FROM order_items oi
			JOIN cart c ON oi.cart_id = c.id
			JOIN menu m ON c.menu_id = m.id
			WHERE oi.order_id = $1`, o.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось загрузить элементы заказа: %v", err)})
			return
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item models.CartItem
			if err := itemRows.Scan(&item.ID, &item.UserID, &item.MenuID, &item.Quantity, &item.MenuName, &item.MenuPrice, &item.MenuDescription); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
				return
			}
			o.Items = append(o.Items, item)
		}

		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, orders)
}

func (app *App) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		DeliveryAddress string `json:"delivery_address"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Получаем товары из корзины
	rows, err := app.DB.Query(`
		SELECT c.id, c.menu_id, c.quantity, m.price, m.restaurant_id
		FROM cart c
		JOIN menu m ON c.menu_id = m.id
		WHERE c.user_id = $1`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось загрузить корзину: %v", err)})
		return
	}
	defer rows.Close()

	var cartItems []models.CartItem
	var totalPrice float64
	var restaurantID int
	first := true
	for rows.Next() {
		var item models.CartItem
		var price float64
		var rID int
		if err := rows.Scan(&item.ID, &item.MenuID, &item.Quantity, &price, &rID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		if first {
			restaurantID = rID
			first = false
		} else if restaurantID != rID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Все товары должны быть из одного ресторана"})
			return
		}
		totalPrice += price * float64(item.Quantity)
		cartItems = append(cartItems, item)
	}

	if len(cartItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Корзина пуста"})
		return
	}

	// Создаём заказ
	var orderID int
	err = app.DB.QueryRow("INSERT INTO orders (user_id, restaurant_id, delivery_address, total_price, status) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userID, restaurantID, input.DeliveryAddress, totalPrice, "preparing").Scan(&orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось создать заказ: %v", err)})
		return
	}

	// Сохраняем элементы заказа
	for _, item := range cartItems {
		_, err = app.DB.Exec("INSERT INTO order_items (order_id, cart_id) VALUES ($1, $2)", orderID, item.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось сохранить элементы заказа: %v", err)})
			return
		}
	}

	// Очищаем корзину
	_, err = app.DB.Exec("DELETE FROM cart WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось очистить корзину: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заказ успешно создан", "order_id": orderID})
}

func (app *App) GetOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orderID := c.Param("id")

	var order models.Order
	err := app.DB.QueryRow(`
		SELECT id, user_id, restaurant_id, delivery_address, total_price, status
		FROM orders
		WHERE id = $1 AND user_id = $2`, orderID, userID).
		Scan(&order.ID, &order.UserID, &order.RestaurantID, &order.DeliveryAddress, &order.TotalPrice, &order.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Заказ не найден"})
		return
	}

	itemRows, err := app.DB.Query(`
		SELECT c.id, c.user_id, c.menu_id, c.quantity, m.name, m.price, m.description
		FROM order_items oi
		JOIN cart c ON oi.cart_id = c.id
		JOIN menu m ON c.menu_id = m.id
		WHERE oi.order_id = $1`, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось загрузить элементы заказа: %v", err)})
		return
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item models.CartItem
		if err := itemRows.Scan(&item.ID, &item.UserID, &item.MenuID, &item.Quantity, &item.MenuName, &item.MenuPrice, &item.MenuDescription); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		order.Items = append(order.Items, item)
	}

	c.JSON(http.StatusOK, order)
}

func (app *App) AddMenuItem(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		RestaurantID int     `json:"restaurant_id"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		Description  string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Проверяем, что ресторан принадлежит пользователю
	var restaurantUserID int
	err := app.DB.QueryRow("SELECT id FROM users WHERE id = $1 AND role = 'restaurant' AND id = $2", input.RestaurantID, userID).
		Scan(&restaurantUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для добавления блюда в этот ресторан"})
		return
	}

	_, err = app.DB.Exec("INSERT INTO menu (restaurant_id, name, price, description) VALUES ($1, $2, $3, $4)",
		input.RestaurantID, input.Name, input.Price, input.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось добавить блюдо: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Блюдо добавлено"})
}

func (app *App) UpdateMenuItem(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		MenuID       int     `json:"menu_id"`
		RestaurantID int     `json:"restaurant_id"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		Description  string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Проверяем, что ресторан принадлежит пользователю
	var restaurantUserID int
	err := app.DB.QueryRow("SELECT id FROM users WHERE id = $1 AND role = 'restaurant' AND id = $2", input.RestaurantID, userID).
		Scan(&restaurantUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для редактирования этого блюда"})
		return
	}

	_, err = app.DB.Exec("UPDATE menu SET name = $1, price = $2, description = $3 WHERE id = $4 AND restaurant_id = $5",
		input.Name, input.Price, input.Description, input.MenuID, input.RestaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось обновить блюдо: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Блюдо обновлено"})
}

func (app *App) DeleteMenuItem(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		MenuID       int `json:"menu_id"`
		RestaurantID int `json:"restaurant_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Проверяем, что ресторан принадлежит пользователю
	var restaurantUserID int
	err := app.DB.QueryRow("SELECT id FROM users WHERE id = $1 AND role = 'restaurant' AND id = $2", input.RestaurantID, userID).
		Scan(&restaurantUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для удаления этого блюда"})
		return
	}

	_, err = app.DB.Exec("DELETE FROM menu WHERE id = $1 AND restaurant_id = $2", input.MenuID, input.RestaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось удалить блюдо: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Блюдо удалено"})
}

func (app *App) GetRestaurantOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	restaurantID := c.Query("restaurant_id")
	if restaurantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ресторана обязателен"})
		return
	}

	// Проверяем, что ресторан принадлежит пользователю
	var restaurantUserID int
	err := app.DB.QueryRow("SELECT id FROM users WHERE id = $1 AND role = 'restaurant' AND id = $2", restaurantID, userID).
		Scan(&restaurantUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для просмотра заказов этого ресторана"})
		return
	}

	rows, err := app.DB.Query(`
		SELECT o.id, o.user_id, o.restaurant_id, o.delivery_address, o.total_price, o.status
		FROM orders o
		WHERE o.restaurant_id = $1`, restaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось загрузить заказы: %v", err)})
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.RestaurantID, &o.DeliveryAddress, &o.TotalPrice, &o.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
			return
		}

		itemRows, err := app.DB.Query(`
			SELECT c.id, c.user_id, c.menu_id, c.quantity, m.name, m.price, m.description
			FROM order_items oi
			JOIN cart c ON oi.cart_id = c.id
			JOIN menu m ON c.menu_id = m.id
			WHERE oi.order_id = $1`, o.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось загрузить элементы заказа: %v", err)})
			return
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item models.CartItem
			if err := itemRows.Scan(&item.ID, &item.UserID, &item.MenuID, &item.Quantity, &item.MenuName, &item.MenuPrice, &item.MenuDescription); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
				return
			}
			o.Items = append(o.Items, item)
		}

		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, orders)
}

func (app *App) UpdateOrderStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		OrderID      int    `json:"order_id"`
		RestaurantID int    `json:"restaurant_id"`
		Status       string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Проверяем, что ресторан принадлежит пользователю
	var restaurantUserID int
	err := app.DB.QueryRow("SELECT id FROM users WHERE id = $1 AND role = 'restaurant' AND id = $2", input.RestaurantID, userID).
		Scan(&restaurantUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для обновления статуса этого заказа"})
		return
	}

	// Валидация статуса
	validStatuses := map[string]bool{
		"preparing": true,
		"en_route":  true,
		"delivered": true,
	}
	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный статус заказа"})
		return
	}

	_, err = app.DB.Exec("UPDATE orders SET status = $1 WHERE id = $2 AND restaurant_id = $3",
		input.Status, input.OrderID, input.RestaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось обновить статус заказа: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус заказа обновлён"})
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("не удалось сгенерировать сессионный ID: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
