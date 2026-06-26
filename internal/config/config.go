package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string 
	JWTSecret string 
}

func Load() *Config {
	// Best-effort load of .env. Equivalent to `require('dotenv').config()`.
	_ = godotenv.Load()

	return &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-ro20"),
	}
}


func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
