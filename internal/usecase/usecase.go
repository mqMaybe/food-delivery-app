package usecase

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Database интерфейс для работы с базой данных
type Database interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Begin() (*sql.Tx, error) // Добавляем метод Begin
}

// UseCase структура для бизнес-логики
type UseCase struct {
	db Database
}

// NewUseCase создаёт новый экземпляр UseCase
func NewUseCase(db Database) *UseCase {
	return &UseCase{db: db}
}

// CartItem структура для товара в корзине
type CartItem struct {
	ID              int     `json:"id"`
	UserID          int     `json:"user_id"`
	MenuID          int     `json:"menu_id"`
	MenuName        string  `json:"menu_name"`
	MenuPrice       float64 `json:"menu_price"`
	MenuDescription string  `json:"menu_description"`
	Quantity        int     `json:"quantity"`
	RestaurantID    int     `json:"restaurant_id"`
}

// Order структура для заказа
type Order struct {
	ID              int         `json:"id"`
	UserID          int         `json:"user_id"`
	RestaurantID    int         `json:"restaurant_id"`
	TotalPrice      float64     `json:"total_price"`
	DeliveryAddress string      `json:"delivery_address"`
	Status          string      `json:"status"`
	Items           []OrderItem `json:"items"`
}

// OrderItem структура для товара в заказе
type OrderItem struct {
	MenuID    int     `json:"menu_id"`
	MenuName  string  `json:"menu_name"`
	MenuPrice float64 `json:"menu_price"`
	Quantity  int     `json:"quantity"`
}

// RegisterUser регистрирует нового пользователя
// Хеширует пароль и создаёт запись в таблице users
// Если роль "restaurant", создаёт запись в таблице restaurants
// Возвращает ошибку в случае неудачи
func (uc *UseCase) RegisterUser(name, email, password, role, cuisineType string) error {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Начинаем транзакцию
	tx, err := uc.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Создаём пользователя
	var userID int
	err = tx.QueryRow(
		"INSERT INTO users (name, email, password, role) VALUES ($1, $2, $3, $4) RETURNING id",
		name, email, hashedPassword, role,
	).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to register user: %v", err)
	}

	// Если роль "restaurant", создаём запись в таблице restaurants
	if role == "restaurant" {
		_, err = tx.Exec(
			"INSERT INTO restaurants (user_id, name, cuisine_type) VALUES ($1, $2, $3)",
			userID, name, cuisineType,
		)
		if err != nil {
			return fmt.Errorf("failed to create restaurant: %v", err)
		}
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// LoginUser выполняет вход пользователя
// Возвращает ID пользователя, его роль и ошибку
func (uc *UseCase) LoginUser(email, password string) (int, string, error) {
	var userID int
	var storedPassword, role string
	// Ищем пользователя по email
	err := uc.db.QueryRow(
		"SELECT id, password, role FROM users WHERE email = $1",
		email,
	).Scan(&userID, &storedPassword, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", errors.New("user not found")
		}
		return 0, "", fmt.Errorf("failed to query user: %v", err)
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		return 0, "", errors.New("invalid password")
	}

	return userID, role, nil
}

// GetRestaurants возвращает список ресторанов
func (uc *UseCase) GetRestaurants(cuisineType, deliveryTime, rating string) ([]map[string]interface{}, error) {
	query := "SELECT id, name, rating, cuisine_type, delivery_time FROM restaurants"
	var conditions []string
	var args []interface{}
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

	rows, err := uc.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query restaurants: %v", err)
	}
	defer rows.Close()

	var restaurants []map[string]interface{}
	for rows.Next() {
		var id int
		var name, cuisineType string
		var rating float64
		var deliveryTime sql.NullInt64 // Используем sql.NullInt64, так как delivery_time может быть NULL
		if err := rows.Scan(&id, &name, &rating, &cuisineType, &deliveryTime); err != nil {
			return nil, fmt.Errorf("failed to scan restaurant: %v", err)
		}
		restaurant := map[string]interface{}{
			"id":           id,
			"name":         name,
			"rating":       rating,
			"cuisine_type": cuisineType,
		}
		if deliveryTime.Valid {
			restaurant["delivery_time"] = deliveryTime.Int64
		} else {
			restaurant["delivery_time"] = nil
		}
		restaurants = append(restaurants, restaurant)
	}

	return restaurants, nil
}

// AddMenuItem добавляет блюдо в меню
// Проверяет, что ресторан принадлежит пользователю, и добавляет блюдо
// Возвращает ID нового блюда и ошибку
func (uc *UseCase) AddMenuItem(userID, restaurantID int, name string, price float64, description string) (int, error) {
	// Проверяем, что ресторан принадлежит пользователю
	var dbRestaurantID int
	err := uc.db.QueryRow(
		"SELECT id FROM restaurants WHERE user_id = $1 AND id = $2",
		userID, restaurantID,
	).Scan(&dbRestaurantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("restaurant not found or not owned by user")
		}
		return 0, fmt.Errorf("failed to query restaurant: %v", err)
	}

	// Добавляем блюдо в меню и получаем ID
	var menuID int
	err = uc.db.QueryRow(
		"INSERT INTO menu (restaurant_id, name, price, description) VALUES ($1, $2, $3, $4) RETURNING id",
		restaurantID, name, price, description,
	).Scan(&menuID)
	if err != nil {
		return 0, fmt.Errorf("failed to add menu item: %v", err)
	}

	return menuID, nil
}

// UpdateMenuItem обновляет блюдо в меню
// Проверяет права доступа и обновляет блюдо
// Возвращает ID обновлённого блюда и ошибку
func (uc *UseCase) UpdateMenuItem(userID, menuID, restaurantID int, name string, price float64, description string) (int, error) {
	// Проверяем, что блюдо существует
	var dbRestaurantID int
	err := uc.db.QueryRow(
		"SELECT restaurant_id FROM menu WHERE id = $1",
		menuID,
	).Scan(&dbRestaurantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("menu item not found")
		}
		return 0, fmt.Errorf("failed to query menu item: %v", err)
	}

	// Проверяем, что ресторан принадлежит пользователю
	err = uc.db.QueryRow(
		"SELECT id FROM restaurants WHERE user_id = $1 AND id = $2",
		userID, restaurantID,
	).Scan(&dbRestaurantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("restaurant not found or not owned by user")
		}
		return 0, fmt.Errorf("failed to query restaurant: %v", err)
	}

	// Обновляем блюдо
	_, err = uc.db.Exec(
		"UPDATE menu SET name = $1, price = $2, description = $3 WHERE id = $4 AND restaurant_id = $5",
		name, price, description, menuID, restaurantID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update menu item: %v", err)
	}

	return menuID, nil
}

// DeleteMenuItem удаляет блюдо из меню
// Проверяет права доступа и удаляет блюдо
// Возвращает ошибку, если операция не удалась
func (uc *UseCase) DeleteMenuItem(menuID, restaurantID int) error {
	// Сначала удаляем связанные записи из cart
	_, err := uc.db.Exec(`DELETE FROM cart WHERE menu_id = $1`, menuID)
	if err != nil {
		return fmt.Errorf("failed to delete related cart items: %v", err)
	}

	// Затем удаляем запись из menu
	result, err := uc.db.Exec(`DELETE FROM menu WHERE id = $1 AND restaurant_id = $2`, menuID, restaurantID)
	if err != nil {
		return fmt.Errorf("failed to delete menu item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("menu item with id %d not found or does not belong to restaurant %d", menuID, restaurantID)
	}

	return nil
}

// GetMenu возвращает меню ресторана
// Возвращает список блюд в формате []map[string]interface{} и ошибку
func (uc *UseCase) GetMenu(restaurantID int) ([]map[string]interface{}, error) {
	rows, err := uc.db.Query(
		"SELECT id, restaurant_id, name, price, description FROM menu WHERE restaurant_id = $1",
		restaurantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu: %v", err)
	}
	defer rows.Close()

	var menu []map[string]interface{}
	for rows.Next() {
		var id, restaurantID int
		var name, description string
		var price float64
		if err := rows.Scan(&id, &restaurantID, &name, &price, &description); err != nil {
			return nil, fmt.Errorf("failed to scan menu item: %v", err)
		}
		menu = append(menu, map[string]interface{}{
			"id":            id,
			"restaurant_id": restaurantID,
			"name":          name,
			"price":         price,
			"description":   description,
		})
	}

	return menu, nil
}

// AddToCart добавляет товар в корзину
// Проверяет, что блюдо существует, и добавляет его в корзину
// Возвращает ID записи в корзине и ошибку
func (uc *UseCase) AddToCart(userID, menuID, quantity int) (int, error) {
	// Проверяем, что блюдо существует
	var restaurantID int
	var name string
	var price float64
	var description string
	err := uc.db.QueryRow(
		"SELECT restaurant_id, name, price, description FROM menu WHERE id = $1",
		menuID,
	).Scan(&restaurantID, &name, &price, &description)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("menu item not found")
		}
		return 0, fmt.Errorf("failed to query menu item: %v", err)
	}

	// Добавляем товар в корзину и получаем ID
	var cartID int
	err = uc.db.QueryRow(
		"INSERT INTO cart (user_id, menu_id, quantity) VALUES ($1, $2, $3) RETURNING id",
		userID, menuID, quantity,
	).Scan(&cartID)
	if err != nil {
		return 0, fmt.Errorf("failed to add to cart: %v", err)
	}

	return cartID, nil
}

// GetCart возвращает содержимое корзины
// Возвращает список товаров в корзине в формате []CartItem и ошибку
func (uc *UseCase) GetCart(userID int) ([]CartItem, error) {
	rows, err := uc.db.Query(`
		SELECT c.id, c.user_id, c.menu_id, m.name, m.price, m.description, c.quantity, m.restaurant_id
		FROM cart c
		JOIN menu m ON c.menu_id = m.id
		WHERE c.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart: %v", err)
	}
	defer rows.Close()

	var cart []CartItem
	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.MenuID, &item.MenuName, &item.MenuPrice, &item.MenuDescription, &item.Quantity, &item.RestaurantID); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %v", err)
		}
		cart = append(cart, item)
	}

	return cart, nil
}

// UpdateCartItem обновляет количество товара в корзине
// Проверяет права доступа и обновляет количество
// Если quantity <= 0, удаляет товар из корзины
// Возвращает ошибку, если операция не удалась
func (uc *UseCase) UpdateCartItem(userID, cartID, quantity int) error {
	// Проверяем, что товар в корзине принадлежит пользователю
	var dbUserID int
	err := uc.db.QueryRow(
		"SELECT user_id FROM cart WHERE id = $1",
		cartID,
	).Scan(&dbUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("cart item not found")
		}
		return fmt.Errorf("failed to query cart item: %v", err)
	}

	if dbUserID != userID {
		return errors.New("cart item does not belong to user")
	}

	// Если количество <= 0, удаляем товар из корзины
	if quantity <= 0 {
		_, err = uc.db.Exec("DELETE FROM cart WHERE id = $1", cartID)
		if err != nil {
			return fmt.Errorf("failed to delete cart item: %v", err)
		}
		return nil
	}

	// Обновляем количество
	_, err = uc.db.Exec(
		"UPDATE cart SET quantity = $1 WHERE id = $2 AND user_id = $3",
		quantity, cartID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update cart item: %v", err)
	}

	return nil
}

// DeleteCartItem удаляет товар из корзины
// Проверяет права доступа и удаляет товар
// Возвращает ошибку, если операция не удалась
func (uc *UseCase) DeleteCartItem(userID, cartID int) error {
	// Проверяем, что товар в корзине принадлежит пользователю
	var dbUserID int
	err := uc.db.QueryRow(
		"SELECT user_id FROM cart WHERE id = $1",
		cartID,
	).Scan(&dbUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("cart item not found")
		}
		return fmt.Errorf("failed to query cart item: %v", err)
	}

	if dbUserID != userID {
		return errors.New("cart item does not belong to user")
	}

	// Удаляем товар из корзины
	_, err = uc.db.Exec("DELETE FROM cart WHERE id = $1 AND user_id = $2", cartID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete cart item: %v", err)
	}

	return nil
}

// CreateOrder создаёт заказ
// Использует транзакцию для создания заказа, добавления товаров и очистки корзины
// Возвращает ID нового заказа и ошибку
func (uc *UseCase) CreateOrder(userID, restaurantID int, cartItems []CartItem, totalPrice float64, deliveryAddress string) (int, error) {
	// Начинаем транзакцию
	tx, err := uc.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Создаём заказ и получаем ID
	var orderID int
	err = tx.QueryRow(
		"INSERT INTO orders (user_id, restaurant_id, total_price, delivery_address, status) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userID, restaurantID, totalPrice, deliveryAddress, "pending",
	).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %v", err)
	}

	// Добавляем товары в заказ
	for _, item := range cartItems {
		_, err = tx.Exec(
			"INSERT INTO order_items (order_id, menu_id, quantity, menu_name, menu_price) VALUES ($1, $2, $3, $4, $5)",
			orderID, item.MenuID, item.Quantity, item.MenuName, item.MenuPrice,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to add order item: %v", err)
		}
	}

	// Очищаем корзину
	_, err = tx.Exec("DELETE FROM cart WHERE user_id = $1", userID)
	if err != nil {
		return 0, fmt.Errorf("failed to clear cart: %v", err)
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return orderID, nil
}

// GetRestaurantOrders возвращает заказы ресторана
// Проверяет права доступа и возвращает заказы
// Возвращает список заказов в формате []Order и ошибку
func (uc *UseCase) GetRestaurantOrders(userID, restaurantID int) ([]Order, error) {
	// Проверяем, что ресторан принадлежит пользователю
	var dbRestaurantID int
	err := uc.db.QueryRow(
		"SELECT id FROM restaurants WHERE user_id = $1 AND id = $2",
		userID, restaurantID,
	).Scan(&dbRestaurantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("restaurant not found or not owned by user")
		}
		return nil, fmt.Errorf("failed to query restaurant: %v", err)
	}

	// Получаем заказы ресторана
	rows, err := uc.db.Query(`
		SELECT o.id, o.user_id, o.restaurant_id, o.total_price, o.delivery_address, o.status
		FROM orders o
		WHERE o.restaurant_id = $1`,
		restaurantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %v", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.RestaurantID, &order.TotalPrice, &order.DeliveryAddress, &order.Status); err != nil {
			return nil, fmt.Errorf("failed to scan order: %v", err)
		}

		// Получаем товары в заказе
		itemRows, err := uc.db.Query(`
			SELECT oi.menu_id, oi.menu_name, oi.menu_price, oi.quantity
			FROM order_items oi
			WHERE oi.order_id = $1`,
			order.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query order items: %v", err)
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item OrderItem
			if err := itemRows.Scan(&item.MenuID, &item.MenuName, &item.MenuPrice, &item.Quantity); err != nil {
				return nil, fmt.Errorf("failed to scan order item: %v", err)
			}
			order.Items = append(order.Items, item)
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// UpdateOrderStatus обновляет статус заказа
// Проверяет права доступа и обновляет статус
// Возвращает ID обновлённого заказа и ошибку
func (uc *UseCase) UpdateOrderStatus(userID, orderID, restaurantID int, status string) (int, error) {
	// Проверяем, что заказ существует
	var dbRestaurantID int
	err := uc.db.QueryRow(
		"SELECT restaurant_id FROM orders WHERE id = $1",
		orderID,
	).Scan(&dbRestaurantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("order not found")
		}
		return 0, fmt.Errorf("failed to query order: %v", err)
	}

	// Проверяем, что ресторан принадлежит пользователю
	err = uc.db.QueryRow(
		"SELECT id FROM restaurants WHERE user_id = $1 AND id = $2",
		userID, restaurantID,
	).Scan(&dbRestaurantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("restaurant not found or not owned by user")
		}
		return 0, fmt.Errorf("failed to query restaurant: %v", err)
	}

	// Обновляем статус заказа
	_, err = uc.db.Exec(
		"UPDATE orders SET status = $1 WHERE id = $2 AND restaurant_id = $3",
		status, orderID, restaurantID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update order status: %v", err)
	}

	return orderID, nil
}

// GetUserOrders возвращает заказы пользователя
// Возвращает список заказов в формате []Order и ошибку
// GetOrders возвращает список заказов пользователя
func (uc *UseCase) GetUserOrders(userID int, page, limit int) ([]Order, error) {
	offset := (page - 1) * limit
	query := "SELECT id, user_id, restaurant_id, total_price, delivery_address, status FROM orders WHERE user_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3"
	rows, err := uc.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %v", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.RestaurantID, &order.TotalPrice, &order.DeliveryAddress, &order.Status); err != nil {
			return nil, fmt.Errorf("failed to scan order: %v", err)
		}

		// Загружаем товары в заказе
		itemsQuery := "SELECT menu_id, menu_name, menu_price, quantity FROM order_items WHERE order_id = $1"
		itemsRows, err := uc.db.Query(itemsQuery, order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query order items: %v", err)
		}
		defer itemsRows.Close()

		for itemsRows.Next() {
			var item OrderItem
			if err := itemsRows.Scan(&item.MenuID, &item.MenuName, &item.MenuPrice, &item.Quantity); err != nil {
				return nil, fmt.Errorf("failed to scan order item: %v", err)
			}
			order.Items = append(order.Items, item)
		}

		orders = append(orders, order)
	}

	return orders, nil
}
