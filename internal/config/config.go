package config

type Config struct {
	Port   string
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func LoadConfig() (*Config, error) {
	return &Config{
		Port:   "8081",
		DBHost: "localhost",
		DBPort: "5432",
		DBUser: "postgres",
		DBPass: "root",
		DBName: "food_delivery",
	}, nil
}
