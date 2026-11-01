package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	AppPort            string
	AppName            string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	DBTimezone         string
	JWTSecret          string
	JWTExpirationHours int
}

func LoadConfig() (*Config, error) {
	// Try loading .env file if it exists (ignore error if not present in production/container)
	_ = godotenv.Load()

	jwtExpHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if jwtExpHours <= 0 {
		jwtExpHours = 24
	}

	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		AppPort:            getEnv("APP_PORT", "8080"),
		AppName:            getEnv("APP_NAME", "Vi-Pharma-Backend"),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		DBName:             getEnv("DB_NAME", "vipharma_db"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		DBTimezone:         getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret-vipharma-key-change-in-production"),
		JWTExpirationHours: jwtExpHours,
	}

	return cfg, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode, c.DBTimezone,
	)
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return fallback
}
