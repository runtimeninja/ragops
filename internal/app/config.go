package app

import "os"

type Config struct {
	Env         string
	HTTPAddr    string
	DatabaseURL string
	RedisAddr   string
}

func Load() Config {
	return Config{
		Env:         getenv("RAGOPS_ENV", "dev"),
		HTTPAddr:    getenv("RAGOPS_HTTP_ADDR", ":8080"),
		DatabaseURL: getenv("RAGOPS_DATABASE_URL", ""),
		RedisAddr:   getenv("RAGOPS_REDIS_ADDR", "localhost:6379"),
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
