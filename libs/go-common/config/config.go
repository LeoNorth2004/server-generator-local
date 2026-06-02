package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DBHost       string
	DBPort       int
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	JWTSecret    string
	GinMode      string
}

func LoadConfig() Config {
	port := getEnv("GENERATOR_SERVICE_PORT", "8084")
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))

	return Config{
		Port:       port,
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "123456"),
		DBName:     getEnv("DB_NAME", "generator_platform"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "low-code-platform-secret-key-2024"),
		GinMode:    getEnv("GIN_MODE", "debug"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
