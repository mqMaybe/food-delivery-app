package delivery

import (
	"food-delivery-app/internal/log"
	"food-delivery-app/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Delivery структура для обработки HTTP-запросов
type Delivery struct {
	uc     *usecase.UseCase
	logger *log.Logger
}

// NewDelivery создаёт новый экземпляр Delivery
func NewDelivery(uc *usecase.UseCase, logger *log.Logger) *Delivery {
	return &Delivery{
		uc:     uc,
		logger: logger,
	}
}

// RenderPage рендерит страницу без проверки авторизации
func (d *Delivery) RenderPage(page string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, page, nil)
	}
}

// RenderPageWithAuth рендерит страницу с проверкой авторизации и роли
func (d *Delivery) RenderPageWithAuth(page, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		role := session.Get("role")

		// Проверяем, авторизован ли пользователь
		if userID == nil {
			d.logger.Info("Unauthorized access attempt to " + page)
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		// Проверяем роль пользователя
		if role != requiredRole {
			d.logger.Info("Access denied to " + page + " for role " + role.(string))
			if role == "customer" {
				c.Redirect(http.StatusSeeOther, "/restaurants")
			} else if role == "restaurant" {
				c.Redirect(http.StatusSeeOther, "/restaurant-orders")
			}
			return
		}

		// Рендерим страницу
		c.HTML(http.StatusOK, page, nil)
	}
}

// RegisterUser обрабатывает регистрацию нового пользователя
func (d *Delivery) RegisterUser(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		Role        string `json:"role"`
		CuisineType string `json:"cuisine_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		d.logger.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	if req.Role != "customer" && req.Role != "restaurant" {
		d.logger.Info("Invalid role: " + req.Role)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверная роль"})
		return
	}

	err := d.uc.RegisterUser(req.Name, req.Email, req.Password, req.Role, req.CuisineType)
	if err != nil {
		d.logger.Error("Failed to register user", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d.logger.Info("User registered successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Регистрация прошла успешно!"})
}

// LoginUser обрабатывает вход пользователя
func (d *Delivery) LoginUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		d.logger.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	userID, role, err := d.uc.LoginUser(req.Email, req.Password)
	if err != nil {
		d.logger.Error("Failed to login user", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", userID)
	session.Set("role", role)
	if err := session.Save(); err != nil {
		d.logger.Error("Failed to save session", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить сессию"})
		return
	}

	d.logger.Info("User logged in successfully: " + req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "Вход выполнен успешно!", "user_id": userID, "role": role})
}

// LogoutUser обрабатывает выход пользователя
func (d *Delivery) LogoutUser(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		d.logger.Error("Failed to clear session", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось очистить сессию"})
		return
	}

	d.logger.Info("User logged out successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно!"})
}

// GetSession возвращает данные текущей сессии пользователя
func (h *Delivery) GetSession(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil {
		c.JSON(http.StatusOK, gin.H{
			"user_id": nil,
			"role":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"role":    role,
	})
}

// / GetRestaurants возвращает список ресторанов
func (d *Delivery) GetRestaurants(c *gin.Context) {
	cuisineType := c.Query("cuisine_type")
	deliveryTime := c.Query("delivery_time")
	rating := c.Query("rating")

	restaurants, err := d.uc.GetRestaurants(cuisineType, deliveryTime, rating)
	if err != nil {
		d.logger.Error("Failed to get restaurants", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, restaurants)
}

// AddMenuItem добавляет блюдо в меню
func (d *Delivery) AddMenuItem(c *gin.Context) {
	var req struct {
		RestaurantID int     `json:"restaurant_id"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		Description  string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		d.logger.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil || role != "restaurant" {
		d.logger.Info("Unauthorized attempt to add menu item")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация как ресторан"})
		return
	}

	menuID, err := d.uc.AddMenuItem(userID.(int), req.RestaurantID, req.Name, req.Price, req.Description)
	if err != nil {
		d.logger.Error("Failed to add menu item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.logger.Info("Menu item added successfully: " + req.Name)
	c.JSON(http.StatusOK, gin.H{"message": "Блюдо успешно добавлено!", "menu_id": menuID})
}

// UpdateMenuItem обновляет блюдо в меню
func (d *Delivery) UpdateMenuItem(c *gin.Context) {
	var req struct {
		MenuID       int     `json:"menu_id"`
		RestaurantID int     `json:"restaurant_id"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		Description  string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		d.logger.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil || role != "restaurant" {
		d.logger.Info("Unauthorized attempt to update menu item")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация как ресторан"})
		return
	}

	updatedMenuID, err := d.uc.UpdateMenuItem(userID.(int), req.MenuID, req.RestaurantID, req.Name, req.Price, req.Description)
	if err != nil {
		d.logger.Error("Failed to update menu item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.logger.Info("Menu item updated successfully: " + req.Name)
	c.JSON(http.StatusOK, gin.H{"message": "Блюдо успешно обновлено!", "menu_id": updatedMenuID})
}

// DeleteMenuItem удаляет блюдо из меню
func (d *Delivery) DeleteMenuItem(c *gin.Context) {
	var input struct {
		MenuID       int `json:"menu_id"`
		RestaurantID int `json:"restaurant_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	if input.MenuID == 0 || input.RestaurantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "menu_id и restaurant_id обязательны"})
		return
	}

	err := d.uc.DeleteMenuItem(input.MenuID, input.RestaurantID)
	if err != nil {
		d.logger.Error("[ERROR] Failed to delete menu item: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Menu item deleted successfully"})
}

// GetMenu возвращает меню ресторана
func (d *Delivery) GetMenu(c *gin.Context) {
	restaurantIDStr := c.Query("restaurant_id")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		d.logger.Error("Invalid restaurant_id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID ресторана"})
		return
	}

	menu, err := d.uc.GetMenu(restaurantID)
	if err != nil {
		d.logger.Error("Failed to get menu", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, menu)
}

// AddToCart добавляет товар в корзину
func (d *Delivery) AddToCart(c *gin.Context) {
	var req struct {
		MenuID   int `json:"menu_id"`
		Quantity int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		d.logger.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil || role != "customer" {
		d.logger.Info("Unauthorized attempt to add to cart")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация как клиент"})
		return
	}

	cartID, err := d.uc.AddToCart(userID.(int), req.MenuID, req.Quantity)
	if err != nil {
		d.logger.Error("Failed to add to cart", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.logger.Info("Item added to cart successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Товар добавлен в корзину!", "cart_id": cartID})
}

// GetCart возвращает содержимое корзины
func (d *Delivery) GetCart(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil || role != "customer" {
		d.logger.Info("Unauthorized attempt to get cart")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация как клиент"})
		return
	}

	cart, err := d.uc.GetCart(userID.(int))
	if err != nil {
		d.logger.Error("Failed to get cart", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// UpdateCartItem обновляет количество товара в корзине
func (d *Delivery) UpdateCartItem(c *gin.Context) {
	var req struct {
		CartID   int `json:"cart_id"`
		Quantity int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		d.logger.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil || role != "customer" {
		d.logger.Info("Unauthorized attempt to update cart")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация как клиент"})
		return
	}

	err := d.uc.UpdateCartItem(userID.(int), req.CartID, req.Quantity)
	if err != nil {
		d.logger.Error("Failed to update cart item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.logger.Info("Cart item updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Количество обновлено!"})
}

// DeleteCartItem удаляет товар из корзины
func (d *Delivery) DeleteCartItem(c *gin.Context) {
	var req struct {
		CartID int `json:"cart_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		d.logger.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil || role != "customer" {
		d.logger.Info("Unauthorized attempt to delete cart item")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация как клиент"})
		return
	}

	err := d.uc.DeleteCartItem(userID.(int), req.CartID)
	if err != nil {
		d.logger.Error("Failed to delete cart item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.logger.Info("Cart item deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Товар удалён из корзины!"})
}

// CreateOrder создаёт заказ
func (d *Delivery) CreateOrder(c *gin.Context) {
	var req struct {
		TotalPrice      float64 `json:"total_price"`
		DeliveryAddress string  `json:"delivery_address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		d.logger.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil || role != "customer" {
		d.logger.Info("Unauthorized attempt to create order")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация как клиент"})
		return
	}

	// Получаем содержимое корзины
	cartItems, err := d.uc.GetCart(userID.(int))
	if err != nil {
		d.logger.Error("Failed to get cart for order", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить корзину"})
		return
	}

	if len(cartItems) == 0 {
		d.logger.Info("Cart is empty for user: " + strconv.Itoa(userID.(int)))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Корзина пуста"})
		return
	}

	// Определяем restaurant_id из первого товара в корзине
	restaurantID := cartItems[0].RestaurantID

	// Проверяем, что все товары в корзине принадлежат одному ресторану
	for _, item := range cartItems {
		if item.RestaurantID != restaurantID {
			d.logger.Info("Cart contains items from different restaurants")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Корзина содержит товары из разных ресторанов"})
			return
		}
	}

	orderID, err := d.uc.CreateOrder(userID.(int), restaurantID, cartItems, req.TotalPrice, req.DeliveryAddress)
	if err != nil {
		d.logger.Error("Failed to create order", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.logger.Info("Order created successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Заказ успешно оформлен!", "order_id": orderID})
}

// GetRestaurantOrders возвращает заказы ресторана
func (d *Delivery) GetRestaurantOrders(c *gin.Context) {
	restaurantIDStr := c.Query("restaurant_id")
	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		d.logger.Error("Invalid restaurant_id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID ресторана"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil || role != "restaurant" {
		d.logger.Info("Unauthorized attempt to get restaurant orders")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация как ресторан"})
		return
	}

	orders, err := d.uc.GetRestaurantOrders(userID.(int), restaurantID)
	if err != nil {
		d.logger.Error("Failed to get restaurant orders", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus обновляет статус заказа
func (d *Delivery) UpdateOrderStatus(c *gin.Context) {
	var req struct {
		OrderID      int    `json:"order_id"`
		RestaurantID int    `json:"restaurant_id"`
		Status       string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		d.logger.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	if userID == nil || role != "restaurant" {
		d.logger.Info("Unauthorized attempt to update order status")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация как ресторан"})
		return
	}

	updatedOrderID, err := d.uc.UpdateOrderStatus(userID.(int), req.OrderID, req.RestaurantID, req.Status)
	if err != nil {
		d.logger.Error("Failed to update order status", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.logger.Info("Order status updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Статус заказа обновлён!", "order_id": updatedOrderID})
}

// GetUserOrders возвращает список заказов пользователя
func (d *Delivery) GetUserOrders(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		d.logger.Info("Unauthorized access attempt to get orders")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неавторизованный доступ"})
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	limit := 10 // Количество заказов на странице

	orders, err := d.uc.GetUserOrders(userID.(int), page, limit)
	if err != nil {
		d.logger.Error("Failed to get orders", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}
