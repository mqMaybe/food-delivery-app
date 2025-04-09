package delivery

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"food-delivery-app/internal/models"
)

type App struct {
	DB *sqlx.DB
}

// HandleLogin обрабатывает вход пользователя по email и паролю
func (app *App) HandleLogin(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	input.Email = strings.ToLower(input.Email)

	var user models.User
	err := app.DB.QueryRow("SELECT id, name, email, password, role, cuisine_type FROM users WHERE email = $1", input.Email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CuisineType)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)) != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
		return
	}

	c.SetCookie("session_id", sessionID, 3600*24, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Вход выполнен успешно", "role": user.Role})
}

// HandleRegister обрабатывает регистрацию пользователя
func (app *App) HandleRegister(c *gin.Context) {
	var input struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Password     string `json:"password"`
		Role         string `json:"role"`
		CuisineType  string `json:"cuisine_type"`
		DeliveryTime int    `json:"delivery_time"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	input.Email = strings.ToLower(input.Email)

	var existingUserID int
	err := app.DB.QueryRow("SELECT id FROM users WHERE email = $1", input.Email).Scan(&existingUserID)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email уже существует"})
		return
	}

	if input.Role != "customer" && input.Role != "restaurant" && input.Role != "rider" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверная роль"})
		return
	}

	if input.Role == "restaurant" && input.CuisineType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Тип кухни обязателен для ресторана"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось хешировать пароль"})
		return
	}

	var userID int
	err = app.DB.QueryRow("INSERT INTO users (name, email, password, role, cuisine_type) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		input.Name, input.Email, hashedPassword, input.Role, input.CuisineType).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось зарегистрировать пользователя"})
		return
	}

	if input.Role == "restaurant" {
		var restaurantID int
		err = app.DB.QueryRow("INSERT INTO restaurants (user_id, name, cuisine_type, delivery_time, rating) VALUES ($1, $2, $3, $4, $5) RETURNING id",
			userID, input.Name, input.CuisineType, input.DeliveryTime, 0.0).Scan(&restaurantID)
		if err != nil {
			app.DB.Exec("DELETE FROM users WHERE id = $1", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать ресторан"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Регистрация прошла успешно"})
}

// HandleLogout завершает сессию пользователя
func (app *App) HandleLogout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		app.DB.Exec("DELETE FROM sessions WHERE session_id = $1", sessionID)
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно"})
}

// HandleSession возвращает ID пользователя и его роль по session_id
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

// generateSessionID создаёт безопасный случайный идентификатор сессии
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (app *App) AddToCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Пользователь не авторизован"})
		return
	}
	log.Printf("UserID: %v", userID)

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{"error": "Не удалось прочитать тело запроса"})
		return
	}

	log.Printf("Полученное тело запроса: %s", string(body))

	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var input struct {
		MenuItemID int `json:"menu_item_id"`
		Quantity   int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Ошибка при парсинге JSON: %v", err)
		c.JSON(400, gin.H{"error": "Неверный формат данных"})
		return
	}

	log.Printf("Распарсенные данные: MenuItemID=%d, Quantity=%d", input.MenuItemID, input.Quantity)

	if input.MenuItemID <= 0 || input.Quantity <= 0 {
		c.JSON(400, gin.H{"error": "Неверный ID блюда или количество"})
		return
	}

	var menuItem models.MenuItem
	err = app.DB.Get(&menuItem, `
        SELECT id, restaurant_id, price
        FROM menu
        WHERE id = $1
    `, input.MenuItemID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Блюдо не найдено"})
		} else {
			log.Printf("Ошибка при проверке блюда: %v", err)
			c.JSON(500, gin.H{"error": "Ошибка при проверке блюда"})
		}
		return
	}
	log.Printf("MenuItem: ID=%d, RestaurantID=%d", menuItem.ID, menuItem.RestaurantID)

	// Проверяем, существует ли ресторан
	var restaurantID int
	err = app.DB.Get(&restaurantID, `
        SELECT id
        FROM restaurants
        WHERE id = $1
    `, menuItem.RestaurantID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Ресторан с ID=%d не найден", menuItem.RestaurantID)
			c.JSON(400, gin.H{"error": "Ресторан, связанный с блюдом, не найден"})
		} else {
			log.Printf("Ошибка при проверке ресторана: %v", err)
			c.JSON(500, gin.H{"error": "Ошибка при проверке ресторана"})
		}
		return
	}

	var existingItem struct {
		ID       int `db:"id"`
		Quantity int `db:"quantity"`
	}

	err = app.DB.Get(&existingItem, `
        SELECT id, quantity
        FROM cart
        WHERE user_id = $1 AND menu_item_id = $2
    `, userID, input.MenuItemID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Ошибка при проверке существующего товара в корзине: %v", err)
		c.JSON(500, gin.H{"error": "Ошибка при обновлении корзины"})
		return
	}

	if err == sql.ErrNoRows {
		_, err = app.DB.Exec(`
            INSERT INTO cart (user_id, menu_item_id, quantity, created_at)
            VALUES ($1, $2, $3, NOW())
        `, userID, input.MenuItemID, input.Quantity)
		if err != nil {
			log.Printf("Ошибка при добавлении товара в корзину (INSERT): %v", err)
			c.JSON(500, gin.H{"error": "Не удалось обновить корзину"})
			return
		}
	} else {
		newQuantity := existingItem.Quantity + input.Quantity
		if newQuantity <= 0 {
			_, err = app.DB.Exec(`DELETE FROM cart WHERE id = $1`, existingItem.ID)
			if err != nil {
				log.Printf("Ошибка при удалении товара из корзины (DELETE): %v", err)
				c.JSON(500, gin.H{"error": "Не удалось обновить корзину"})
				return
			}
		} else {
			_, err = app.DB.Exec(`UPDATE cart SET quantity = $1 WHERE id = $2`, newQuantity, existingItem.ID)
			if err != nil {
				log.Printf("Ошибка при обновлении количества в корзине (UPDATE): %v", err)
				c.JSON(500, gin.H{"error": "Не удалось обновить корзину"})
				return
			}
		}
	}

	c.JSON(200, gin.H{"message": "Товар добавлен в корзину"})
}

// UpdateCartItem изменяет количество определённого товара в корзине
func (app *App) UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Неверный ID товара"})
		return
	}

	var input struct {
		Change int `json:"change"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Неверный формат данных"})
		return
	}

	var cartItem struct {
		Quantity int `db:"quantity"`
	}
	err = app.DB.Get(&cartItem, `
        SELECT quantity
        FROM cart
        WHERE id = $1 AND user_id = $2
    `, itemID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Товар не найден в корзине"})
		} else {
			c.JSON(500, gin.H{"error": "Не удалось получить товар из корзины"})
		}
		return
	}

	newQuantity := cartItem.Quantity + input.Change
	if newQuantity <= 0 {
		c.JSON(400, gin.H{"error": "Количество должно быть больше 0"})
		return
	}

	_, err = app.DB.Exec(`UPDATE cart SET quantity = $1 WHERE id = $2`, newQuantity, itemID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Не удалось обновить корзину"})
		return
	}

	c.JSON(200, gin.H{"message": "Корзина обновлена"})
}

// UpdateCart устанавливает точное количество товара в корзине
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

	_, err := app.DB.Exec("UPDATE cart SET quantity = $1 WHERE id = $2 AND user_id = $3",
		input.Quantity, input.CartID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить корзину"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Корзина обновлена"})
}

// GetOrders возвращает список заказов пользователя
func (app *App) GetOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	rows, err := app.DB.Query(`
        SELECT o.id, o.user_id, o.restaurant_id, o.delivery_address, o.total_price, o.status
        FROM orders o
        WHERE o.user_id = $1
        ORDER BY o.created_at DESC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить заказы"})
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

		o.Items = []models.OrderItem{}
		itemRows, err := app.DB.Query(`
            SELECT id, menu_item_id, quantity, menu_name, menu_price
            FROM order_items
            WHERE order_id = $1`, o.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить элементы заказа"})
			return
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ID, &item.MenuItemID, &item.Quantity, &item.MenuName, &item.MenuPrice); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
				return
			}
			o.Items = append(o.Items, item)
		}

		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, orders)
}

// CreateOrder создаёт новый заказ на основе содержимого корзины
func (app *App) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input struct {
		DeliveryAddress string `json:"delivery_address"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	rows, err := app.DB.Query(`
		SELECT c.id, c.menu_item_id, c.quantity, m.price, m.restaurant_id
		FROM cart c
		JOIN menu m ON c.menu_item_id = m.id
		WHERE c.user_id = $1`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить корзину"})
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
		if err := rows.Scan(&item.ID, &item.MenuItemID, &item.Quantity, &price, &rID); err != nil {
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

	var orderID int
	err = app.DB.QueryRow("INSERT INTO orders (user_id, restaurant_id, delivery_address, total_price, status) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userID, restaurantID, input.DeliveryAddress, totalPrice, "preparing").Scan(&orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать заказ"})
		return
	}

	for _, item := range cartItems {
		_, err = app.DB.Exec(`
			INSERT INTO order_items (order_id, menu_item_id, quantity, menu_name, menu_price)
			SELECT $1, m.id, $2, m.name, m.price
			FROM menu m
			WHERE m.id = $3`,
			orderID, item.Quantity, item.MenuItemID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить элементы заказа"})
			return
		}
	}

	_, err = app.DB.Exec("DELETE FROM cart WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось очистить корзину"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заказ успешно создан", "order_id": orderID})
}

// GetOrder возвращает детали одного заказа пользователя
func (app *App) GetOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	orderID := c.Param("id")

	var order models.Order
	err := app.DB.QueryRow(`
        SELECT id, user_id, restaurant_id, delivery_address, total_price, status
        FROM orders
        WHERE id = $1 AND user_id = $2
    `, orderID, userID).
		Scan(&order.ID, &order.UserID, &order.RestaurantID, &order.DeliveryAddress, &order.TotalPrice, &order.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Заказ не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить заказ"})
		}
		return
	}

	order.Items = []models.OrderItem{}
	rows, err := app.DB.Query(`
        SELECT id, menu_item_id, quantity, menu_name, menu_price
        FROM order_items
        WHERE order_id = $1
    `, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить товары заказа"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.MenuItemID, &item.Quantity, &item.MenuName, &item.MenuPrice); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при чтении товара"})
			return
		}
		order.Items = append(order.Items, item)
	}

	c.JSON(http.StatusOK, order)
}

// GetRestaurants возвращает список ресторанов с фильтрами по кухне, рейтингу и времени доставки
func (app *App) GetRestaurants(c *gin.Context) {
	cuisineType := c.Query("cuisine_type")
	deliveryTime := c.Query("delivery_time")
	rating := c.Query("rating")

	query := "SELECT id, user_id, name, cuisine_type, address, delivery_time, rating, created_at FROM restaurants"
	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if cuisineType != "" && cuisineType != "all" {
		conditions = append(conditions, fmt.Sprintf("cuisine_type = $%d", argIndex))
		args = append(args, cuisineType)
		argIndex++
	}
	if deliveryTime != "" && deliveryTime != "all" {
		conditions = append(conditions, fmt.Sprintf("delivery_time <= $%d", argIndex))
		args = append(args, deliveryTime)
		argIndex++
	}
	if rating != "" && rating != "all" {
		conditions = append(conditions, fmt.Sprintf("rating >= $%d", argIndex))
		args = append(args, rating)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var restaurants []models.Restaurant
	err := app.DB.Select(&restaurants, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить рестораны"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"restaurants": restaurants,
	})
}

// AddMenuItem добавляет новое блюдо в меню ресторана
func (app *App) AddMenuItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	var input struct {
		RestaurantID int     `json:"restaurant_id"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		Description  string  `json:"description"`
		ImageURL     string  `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	var restaurantID int
	err := app.DB.QueryRow("SELECT id FROM restaurants WHERE id = $1 AND user_id = $2",
		input.RestaurantID, userID).Scan(&restaurantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ресторан не найден или не принадлежит этому пользователю"})
		return
	}

	var menuItemID int
	err = app.DB.QueryRow(`
		INSERT INTO menu (restaurant_id, name, price, description, image_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		input.RestaurantID, input.Name, input.Price, input.Description, input.ImageURL,
	).Scan(&menuItemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить блюдо"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Блюдо успешно добавлено", "menu_item_id": menuItemID})
}

// UpdateMenuItem обновляет информацию о блюде
func (app *App) UpdateMenuItem(c *gin.Context) {
	userID, _ := c.Get("user_id")
	log.Printf("UserID: %v", userID)

	var input struct {
		MenuID       int     `json:"menu_id"`
		RestaurantID int     `json:"restaurant_id"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		Description  string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Ошибка при парсинге JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}
	log.Printf("Input: MenuID=%d, RestaurantID=%d", input.MenuID, input.RestaurantID)

	var restaurantUserID int
	err := app.DB.QueryRow("SELECT id FROM restaurants WHERE id = $1 AND user_id = $2",
		input.RestaurantID, userID).Scan(&restaurantUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Ресторан с ID=%d не принадлежит пользователю с ID=%v", input.RestaurantID, userID)
			c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для редактирования этого блюда"})
		} else {
			log.Printf("Ошибка при проверке прав: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при проверке прав"})
		}
		return
	}

	_, err = app.DB.Exec(`
        UPDATE menu SET name = $1, price = $2, description = $3
        WHERE id = $4 AND restaurant_id = $5`,
		input.Name, input.Price, input.Description, input.MenuID, input.RestaurantID)
	if err != nil {
		log.Printf("Ошибка при обновлении блюда: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить блюдо"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Блюдо обновлено"})
}

// DeleteMenuItem удаляет блюдо из меню ресторана
func (app *App) DeleteMenuItem(c *gin.Context) {
	userID, _ := c.Get("user_id")
	log.Printf("UserID: %v", userID)

	var input struct {
		MenuID       int `json:"menu_id"`
		RestaurantID int `json:"restaurant_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Ошибка при парсинге JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}
	log.Printf("Input: MenuID=%d, RestaurantID=%d", input.MenuID, input.RestaurantID)

	var restaurantUserID int
	err := app.DB.QueryRow("SELECT id FROM restaurants WHERE id = $1 AND user_id = $2",
		input.RestaurantID, userID).Scan(&restaurantUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Ресторан с ID=%d не принадлежит пользователю с ID=%v", input.RestaurantID, userID)
			c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для удаления этого блюда"})
		} else {
			log.Printf("Ошибка при проверке прав: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при проверке прав"})
		}
		return
	}

	_, err = app.DB.Exec("DELETE FROM menu WHERE id = $1 AND restaurant_id = $2",
		input.MenuID, input.RestaurantID)
	if err != nil {
		log.Printf("Ошибка при удалении блюда: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить блюдо"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Блюдо удалено"})
}

// GetRestaurantOrders возвращает все заказы для конкретного ресторана, если он принадлежит текущему пользователю
func (app *App) GetRestaurantOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	restaurantID := c.Query("restaurant_id")
	if restaurantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ресторана обязателен"})
		return
	}

	restaurantIDInt, err := strconv.Atoi(restaurantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ресторана должен быть числом"})
		return
	}

	var restaurantUserID int
	err = app.DB.QueryRow("SELECT user_id FROM restaurants WHERE id = $1 AND user_id = $2",
		restaurantIDInt, userID).Scan(&restaurantUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для просмотра заказов этого ресторана"})
		return
	}

	rows, err := app.DB.Query(`
        SELECT o.id, o.user_id, o.restaurant_id, o.delivery_address, o.total_price, o.status
        FROM orders o
        WHERE o.restaurant_id = $1`, restaurantIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить заказы"})
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
            SELECT id, menu_item_id, quantity, menu_name, menu_price
            FROM order_items
            WHERE order_id = $1`, o.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить товары заказа"})
			return
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ID, &item.MenuItemID, &item.Quantity, &item.MenuName, &item.MenuPrice); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при чтении товара"})
				return
			}
			o.Items = append(o.Items, item)
		}

		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, orders)
}

// GetCart возвращает содержимое корзины текущего пользователя
func (app *App) GetCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	var cartItems []struct {
		ID       int            `db:"id" json:"id"`
		Name     string         `db:"name" json:"name"`
		Price    float64        `db:"price" json:"price"`
		Quantity int            `db:"quantity" json:"quantity"`
		ImageURL sql.NullString `db:"image_url" json:"image_url"`
	}

	err := app.DB.Select(&cartItems, `
        SELECT c.id, m.name, m.price, c.quantity, m.image_url
        FROM cart c
        JOIN menu m ON c.menu_item_id = m.id
        WHERE c.user_id = $1
    `, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить корзину"})
		return
	}

	c.JSON(http.StatusOK, cartItems)
}

// RemoveFromCart удаляет один элемент из корзины
func (app *App) RemoveFromCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	itemID := c.Param("id")
	result, err := app.DB.Exec(`
        DELETE FROM cart
        WHERE id = $1 AND user_id = $2
    `, itemID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить товар из корзины"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден в корзине"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Товар удалён из корзины"})
}

// Checkout оформляет заказ из корзины и применяет промокод, если есть
func (app *App) Checkout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	var input struct {
		PromoCode       string `json:"promoCode"`
		DeliveryAddress string `json:"deliveryAddress"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}
	if input.DeliveryAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Адрес доставки обязателен"})
		return
	}

	var cartItems []struct {
		MenuItemID   int     `db:"menu_item_id"`
		Quantity     int     `db:"quantity"`
		MenuName     string  `db:"menu_name"`
		Price        float64 `db:"price"`
		RestaurantID int     `db:"restaurant_id"`
	}
	err := app.DB.Select(&cartItems, `
        SELECT c.menu_item_id, c.quantity, m.name as menu_name, m.price, m.restaurant_id
        FROM cart c
        JOIN menu m ON c.menu_item_id = m.id
        WHERE c.user_id = $1
    `, userID)
	if err != nil || len(cartItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Корзина пуста или не удалось загрузить"})
		return
	}

	restaurantID := cartItems[0].RestaurantID
	for _, item := range cartItems {
		if item.RestaurantID != restaurantID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Все товары должны быть из одного ресторана"})
			return
		}
	}

	total := 0.0
	for _, item := range cartItems {
		total += float64(item.Quantity) * item.Price
	}

	if input.PromoCode != "" {
		var discount float64
		err := app.DB.Get(&discount, `
            SELECT discount FROM promo_codes
            WHERE code = $1 AND valid_until > NOW()
        `, input.PromoCode)
		if err == nil {
			total = total * (1 - discount/100)
		}
	}

	var orderID int
	err = app.DB.QueryRow(`
        INSERT INTO orders (user_id, restaurant_id, total_price, status, delivery_address, created_at)
        VALUES ($1, $2, $3, 'pending', $4, NOW())
        RETURNING id
    `, userID, restaurantID, total, input.DeliveryAddress).Scan(&orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать заказ"})
		return
	}

	for _, item := range cartItems {
		_, err = app.DB.Exec(`
            INSERT INTO order_items (order_id, menu_item_id, quantity, menu_name, menu_price)
            VALUES ($1, $2, $3, $4, $5)
        `, orderID, item.MenuItemID, item.Quantity, item.MenuName, item.Price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить элементы заказа"})
			return
		}
	}

	// Очистка корзины пользователя после успешного оформления заказа
	_, err = app.DB.Exec(`DELETE FROM cart WHERE user_id = $1`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось очистить корзину"})
		return
	}
}

// GetUserRestaurants возвращает рестораны, привязанные к текущему пользователю
func (app *App) GetUserRestaurants(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	var restaurants []models.Restaurant
	err := app.DB.Select(&restaurants, `
		SELECT id, user_id, name, cuisine_type, address, delivery_time, rating, created_at
		FROM restaurants
		WHERE user_id = $1
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить рестораны"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"restaurants": restaurants})
}

// GetMenuItems возвращает меню для заданного ресторана
func (app *App) GetMenuItems(c *gin.Context) {
	restaurantID := c.Query("restaurant_id")
	if restaurantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ресторана обязателен"})
		return
	}

	var menuItems []models.MenuItem
	err := app.DB.Select(&menuItems, `
		SELECT id, restaurant_id, name, price, description, image_url
		FROM menu
		WHERE restaurant_id = $1
	`, restaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить меню"})
		return
	}

	c.JSON(http.StatusOK, menuItems)
}

// UpdateOrderStatus позволяет владельцу ресторана изменить статус заказа
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

	validStatuses := map[string]bool{
		"pending":   true,
		"preparing": true,
		"en_route":  true,
		"delivered": true,
	}
	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый статус заказа"})
		return
	}

	var restaurantUserID int
	err := app.DB.QueryRow(`
		SELECT user_id FROM restaurants WHERE id = $1 AND user_id = $2
	`, input.RestaurantID, userID).Scan(&restaurantUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Вы не владелец этого ресторана"})
		return
	}

	_, err = app.DB.Exec(`
		UPDATE orders SET status = $1 WHERE id = $2 AND restaurant_id = $3
	`, input.Status, input.OrderID, input.RestaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить статус"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус обновлён"})
}
