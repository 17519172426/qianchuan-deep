package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	ServerPort      string
	QianchuanAppID  string
	QianchuanSecret string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://qianchuan:qianchuan_dev@localhost:5432/qianchuan?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret"),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		QianchuanAppID:  getEnv("QIANCHUAN_APP_ID", ""),
		QianchuanSecret: getEnv("QIANCHUAN_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
