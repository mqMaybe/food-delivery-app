package db

import (
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// InitDB инициализирует соединение с базой данных, настраивает пул соединений и запускает миграции
func InitDB() (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к базе данных: %v", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("ошибка выполнения миграций: %v", err)
	}

	fmt.Println("Успешно подключено к базе данных")
	return db, nil
}

// migrate создаёт все необходимые таблицы, если они не существуют
func migrate(db *sqlx.DB) error {
	fmt.Println("Создание таблиц, если они не существуют...")

	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            role TEXT NOT NULL DEFAULT 'customer',
            cuisine_type TEXT,
            delivery_time INT,
            rating DECIMAL DEFAULT 0.0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы users: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS restaurants (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            name VARCHAR(255) NOT NULL,
            cuisine_type VARCHAR(50),
            address TEXT,
            delivery_time INT,
            rating FLOAT DEFAULT 0.0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы restaurants: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS menu (
            id SERIAL PRIMARY KEY,
            restaurant_id INT NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
            name TEXT NOT NULL,
            price DECIMAL NOT NULL,
            description TEXT,
            image_url TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы menu: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS cart (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            menu_item_id INT NOT NULL REFERENCES menu(id) ON DELETE CASCADE,
            quantity INT NOT NULL DEFAULT 1,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            restaurant_id INT NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
            UNIQUE (user_id, menu_item_id)
        );
    `)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы cart: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS orders (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            restaurant_id INT NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
            total_price DECIMAL NOT NULL,
            delivery_address TEXT NOT NULL,
            status TEXT DEFAULT 'pending',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы orders: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS order_items (
            id SERIAL PRIMARY KEY,
            order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
            menu_item_id INT NOT NULL REFERENCES menu(id) ON DELETE CASCADE,
            quantity INT NOT NULL,
            menu_name TEXT NOT NULL,
            menu_price DECIMAL NOT NULL
        );
    `)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы order_items: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS promo_codes (
			id SERIAL PRIMARY KEY,
			code VARCHAR(50) UNIQUE NOT NULL,
			discount DECIMAL(5,2) NOT NULL,
			valid_until TIMESTAMP NOT NULL,
			CONSTRAINT valid_discount CHECK (discount >= 0 AND discount <= 100)
    	);
	`)
	if err != nil {
		return fmt.Errorf("ошибка при создании таблицы promo_codes: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS sessions (
            session_id TEXT PRIMARY KEY,
            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            role TEXT NOT NULL,
            expires_at TIMESTAMP NOT NULL
        );
    `)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы sessions: %v", err)
	}

	return nil
}
