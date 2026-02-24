package app

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func Load() Config {
	return Config{
		HTTPAddr:    getenv("RAGOPS_HTTP_ADDR", ":8080"),
		DatabaseURL: getenv("RAGOPS_DATABASE_URL", ""),
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
