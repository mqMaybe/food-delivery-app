package db

import (
	"database/sql"
	"fmt"
	"food-delivery-app/internal/config"

	_ "github.com/lib/pq"
)

type Database struct {
	*sql.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	fmt.Printf("Connecting to database: host=%s, port=%s, user=%s, dbname=%s\n", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName)
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Failed to open database connection: %v\n", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		fmt.Printf("Failed to ping database: %v\n", err)
		return nil, err
	}

	fmt.Println("Running migrations...")
	if err := migrate(db); err != nil {
		fmt.Printf("Migration failed: %v\n", err)
		return nil, err
	}
	fmt.Println("Database migrations completed successfully")

	return &Database{db}, nil
}

func (d *Database) Begin() (*sql.Tx, error) {
	return d.DB.Begin()
}

func migrate(db *sql.DB) error {
	fmt.Println("Creating tables if not exists...")
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            role TEXT NOT NULL DEFAULT 'customer',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS restaurants (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id),
			name TEXT NOT NULL,
			rating DECIMAL DEFAULT 0.0,
			cuisine_type TEXT, -- Добавляем тип кухни
			delivery_time INT, -- Добавляем время доставки в минутах
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
        CREATE TABLE IF NOT EXISTS menu (
            id SERIAL PRIMARY KEY,
            restaurant_id INT NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
            name TEXT NOT NULL,
            price DECIMAL NOT NULL,
            description TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS cart (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            menu_id INT NOT NULL REFERENCES menu(id) ON DELETE CASCADE,
            quantity INT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS orders (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            restaurant_id INT NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
            total_price DECIMAL NOT NULL,
            delivery_address TEXT NOT NULL,
            status TEXT DEFAULT 'pending',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS order_items (
            id SERIAL PRIMARY KEY,
            order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
            menu_id INT NOT NULL REFERENCES menu(id) ON DELETE CASCADE,
            quantity INT NOT NULL,
            menu_name TEXT NOT NULL,
            menu_price DECIMAL NOT NULL
        );
    `)
	if err != nil {
		fmt.Printf("Error executing migration: %v\n", err)
	}
	return err
}
