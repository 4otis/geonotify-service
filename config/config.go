package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort               string
	DBURL                  string
	RedisURL               string
	WebhookURL             string
	APIKey                 string
	LogLevel               string
	StatsTimeWindowMinutes int
	MaxRetries             int
	RetryDelaySeconds      int
	CacheTTLMinutes        int
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Print("No .env file found, using environment variables")
	}

	return &Config{
		HTTPPort:               getEnv("HTTP_PORT", "8080"),
		DBURL:                  getDBURL(),
		RedisURL:               getEnv("REDIS_URL", "redis://localhost:6379/0"),
		WebhookURL:             getEnv("WEBHOOK_URL", ""),
		APIKey:                 getEnv("SECRET_API_KEY", ""),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		StatsTimeWindowMinutes: getEnvAsInt("STATS_TIME_WINDOWS_MINUTES", 30),
		MaxRetries:             getEnvAsInt("WEBHOOK_MAX_RETRIES", 3),
		RetryDelaySeconds:      getEnvAsInt("WEBHOOK_RETRY_DELAY_SECONDS", 60),
		CacheTTLMinutes:        getEnvAsInt("CACHE_TTL_MINUTES", 10),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		log.Printf("Invalid integer value for %s: %s, using default: %d", key, strValue, defaultValue)
		return defaultValue
	}

	return value
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

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}
