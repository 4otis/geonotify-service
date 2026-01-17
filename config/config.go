package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string
	DBURL    string
	LogLevel string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Print("failed to found .env file")
	}

	return &Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		DBURL:    getDBURL(),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDBURL() string {
	if dbURL := os.Getenv("PG_DB_URL"); dbURL != "" {
		return dbURL
	}

	host := getEnv("PG_DB_HOST", "localhost")
	port := getEnv("PG_DB_PORT", "5434")
	user := getEnv("PG_DB_USER", "postgres")
	password := getEnv("PG_DB_PASSWORD", "password")
	dbname := getEnv("PG_DB_NAME", "geonotify_db")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}
