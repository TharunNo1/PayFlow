package config

import "os"

type Config struct {
	RedisAddr string
	DBURL     string
}

func Load() *Config {
	return &Config{
		// Provide defaults if the environment variable is missing
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		DBURL:     getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/db"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
