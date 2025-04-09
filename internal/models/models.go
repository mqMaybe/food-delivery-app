package models

import (
	"database/sql"
	"encoding/json"
)

type User struct {
	ID          int
	Name        string
	Email       string
	Password    string
	Role        string
	CuisineType sql.NullString
}

type MenuItem struct {
	ID           int            `json:"id" db:"id"`
	RestaurantID int            `json:"restaurant_id" db:"restaurant_id"`
	Name         string         `json:"name" db:"name"`
	Description  sql.NullString `json:"description" db:"description"`
	Price        float64        `json:"price" db:"price"`
	ImageURL     sql.NullString `json:"image_url" db:"image_url"`
	CreatedAt    string         `json:"created_at" db:"created_at"`
}

type CartItem struct {
	ID              int            `json:"id" db:"id"`
	UserID          int            `json:"user_id" db:"user_id"`
	MenuItemID      int            `json:"menu_item_id" db:"menu_item_id"`
	Quantity        int            `json:"quantity" db:"quantity"`
	ImageURL        sql.NullString `db:"image_url" json:"image_url"`
	MenuName        string         `json:"menu_name" db:"menu_name"`
	MenuPrice       float64        `json:"menu_price" db:"menu_price"`
	MenuDescription string         `json:"menu_description" db:"menu_description"`
	RestaurantID    int            `json:"restaurant_id" db:"restaurant_id"`
}

type Order struct {
	ID              int         `json:"id"`
	UserID          int         `json:"user_id"`
	RestaurantID    int         `json:"restaurant_id"`
	DeliveryAddress string      `json:"delivery_address"`
	TotalPrice      float64     `json:"total_price"`
	Status          string      `json:"status"`
	Items           []OrderItem `json:"items"`
}

type Restaurant struct {
	ID           int            `json:"id" db:"id"`
	UserID       int            `json:"user_id" db:"user_id"`
	Name         string         `json:"name" db:"name"`
	CuisineType  sql.NullString `json:"cuisine_type" db:"cuisine_type"`
	Address      sql.NullString `json:"address" db:"address"`
	DeliveryTime sql.NullInt32  `json:"delivery_time" db:"delivery_time"`
	Rating       float64        `json:"rating" db:"rating"`
	CreatedAt    string         `json:"created_at" db:"created_at"`
}

// MarshalJSON сериализует DeliveryTime как null или число
func (r Restaurant) MarshalJSON() ([]byte, error) {
	type Alias Restaurant
	return json.Marshal(&struct {
		*Alias
		DeliveryTime interface{} `json:"delivery_time"`
	}{
		Alias:        (*Alias)(&r),
		DeliveryTime: r.DeliveryTime.Int32,
	})
}

type OrderItem struct {
	ID         int     `json:"id"`
	MenuItemID int     `json:"menu_item_id"`
	Quantity   int     `json:"quantity"`
	MenuName   string  `json:"menu_name"`
	MenuPrice  float64 `json:"menu_price"`
}
