package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort     string
	BaseURL        string
	AllowedOrigins []string
	RedisURL       string
	RedisStartOffset int64
	DatabaseURL    string
	HashSalt       string
	HashPepper     string
	HashMinLength  int
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		ServerPort:       getEnv("SERVER_PORT", ":8080"),
		BaseURL:          getEnv("BASE_URL", "http://localhost:8080"),
		AllowedOrigins:   strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001"), ","),
		RedisURL:         getEnv("REDIS_URL", "redis://127.0.0.1:6379"),
		RedisStartOffset: getEnvInt64("REDIS_START_OFFSET", 14_000_000),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5432/encurtador?sslmode=disable"),
		HashSalt:         getEnv("HASH_SALT", ""),
		HashPepper:       getEnv("HASH_PEPPER", ""),
		HashMinLength:    getEnvInt("HASH_MIN_LENGTH", 6),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}
