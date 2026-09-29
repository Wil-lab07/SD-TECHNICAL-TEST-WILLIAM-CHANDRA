package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret string

	AdminSeedEmail    string
	AdminSeedPassword string
	AdminSeedName     string

	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
}

func Load() Config {
	cfg := Config{
		ServerPort:          getEnvOrDefault("SERVER_PORT", "8080"),
		DBHost:              requireEnv("DB_HOST"),
		DBPort:              getEnvOrDefault("DB_PORT", "5432"),
		DBUser:              requireEnv("DB_USER"),
		DBPassword:          requireEnv("DB_PASSWORD"),
		DBName:              requireEnv("DB_NAME"),
		DBSSLMode:           getEnvOrDefault("DB_SSLMODE", "disable"),
		JWTSecret:           requireEnv("JWT_SECRET"),
		AdminSeedEmail:      getEnvOrDefault("ADMIN_SEED_EMAIL", "admin@football.local"),
		AdminSeedPassword:   getEnvOrDefault("ADMIN_SEED_PASSWORD", ""),
		AdminSeedName:       getEnvOrDefault("ADMIN_SEED_NAME", "Administrator"),
		CloudinaryCloudName: getEnvOrDefault("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnvOrDefault("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnvOrDefault("CLOUDINARY_API_SECRET", ""),
	}
	return cfg
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func (c Config) CloudinaryURL() string {
	if c.CloudinaryCloudName == "" || c.CloudinaryAPIKey == "" || c.CloudinaryAPISecret == "" {
		return os.Getenv("CLOUDINARY_URL")
	}
	return fmt.Sprintf("cloudinary://%s:%s@%s", c.CloudinaryAPIKey, c.CloudinaryAPISecret, c.CloudinaryCloudName)
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("FATAL: required environment variable %s is not set", key))
	}
	return val
}

func getEnvOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
