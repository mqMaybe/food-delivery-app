package models

type User struct {
	ID          int
	Name        string
	Email       string
	Password    string
	Role        string
	CuisineType string
}

type MenuItem struct {
	ID           int
	RestaurantID int
	Name         string
	Price        float64
	Description  string
}

type CartItem struct {
	ID              int
	UserID          int
	MenuID          int
	Quantity        int
	MenuName        string
	MenuPrice       float64
	MenuDescription string
}

type Order struct {
	ID              int
	UserID          int
	RestaurantID    int
	DeliveryAddress string
	TotalPrice      float64
	Status          string
	Items           []CartItem
}

type Restaurant struct {
	ID           int
	Name         string
	CuisineType  string
	DeliveryTime int
	Rating       float64
}
