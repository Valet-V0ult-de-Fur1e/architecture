package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr string
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TodoTTL  time.Duration
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "app"),
			Password: getEnv("DB_PASSWORD", "app"),
			Name:     getEnv("DB_NAME", "app"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "redis:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			TodoTTL:  getEnvAsDuration("REDIS_TODO_TTL", 2*time.Minute),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "dev-secret-change-me"),
			ExpiresIn: getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour),
		},
	}

	if cfg.DB.Host == "" || cfg.DB.Port == "" || cfg.DB.User == "" || cfg.DB.Name == "" {
		return Config{}, fmt.Errorf("db env vars are required")
	}

	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("jwt secret is required")
	}

	return cfg, nil
}

func (c Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host,
		c.DB.Port,
		c.DB.User,
		c.DB.Password,
		c.DB.Name,
		c.DB.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return defaultValue
	}

	return parsed
}
