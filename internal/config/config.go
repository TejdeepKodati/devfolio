package config

import "os"

type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	AdminEmail  string // seeded admin account
	AdminPass   string // seeded admin password
}

func Load() *Config {
	return &Config{
		Env:         getEnv("ENV", "development"),
		Port:        getEnv("PORT", "8070"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/devfolio?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", ""),  // Cleaned: No hardcoded fallbacks
		AdminEmail:  getEnv("ADMIN_EMAIL", ""), // Cleaned: No hardcoded fallbacks
		AdminPass:   getEnv("ADMIN_PASS", ""),  // Cleaned: No hardcoded fallbacks
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
