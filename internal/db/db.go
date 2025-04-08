package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	// Формирование строки подключения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Подключение к базе данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Hour)

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к базе данных: %v", err)
	}

	// Выполняем миграции
	err = migrate(db)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения миграций: %v", err)
	}

	fmt.Println("Успешно подключено к базе данных")
	return db, nil
}

func migrate(db *sql.DB) error {
	fmt.Println("Создание таблиц, если они не существуют...")

	// Таблица users (добавляем cuisine_type и delivery_time, так как они используются в коде)
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'customer',
			cuisine_type TEXT, -- Добавляем тип кухни для ресторанов
			delivery_time INT, -- Добавляем время доставки в минутах для ресторанов
			rating DECIMAL DEFAULT 0.0, -- Добавляем рейтинг для ресторанов
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы users: %v", err)
	}

	// Таблица restaurants (уже есть в исходной миграции)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS restaurants (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id),
			name TEXT NOT NULL,
			rating DECIMAL DEFAULT 0.0,
			cuisine_type TEXT,
			delivery_time INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы restaurants: %v", err)
	}

	// Таблица menu
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS menu (
			id SERIAL PRIMARY KEY,
			restaurant_id INT NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			price DECIMAL NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы menu: %v", err)
	}

	// Таблица cart
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cart (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			menu_id INT NOT NULL REFERENCES menu(id) ON DELETE CASCADE,
			quantity INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы cart: %v", err)
	}

	// Таблица orders
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

	// Таблица order_items
	_, err = db.Exec(`
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
		return fmt.Errorf("ошибка создания таблицы order_items: %v", err)
	}

	// Таблица sessions (добавляем, так как она используется в коде для управления сессиями)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы sessions: %v", err)
	}

	fmt.Println("Миграции успешно выполнены")
	return nil
}
