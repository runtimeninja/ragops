package app

import "os"

type Config struct {
	Env             string
	HTTPAddr        string
	DatabaseURL     string
	RedisAddr       string
	OpenAIAPIKey    string
	OpenAIEmbModel  string
	OpenAIChatModel string
}

func Load() Config {
	return Config{
		Env:             getenv("RAGOPS_ENV", "dev"),
		HTTPAddr:        getenv("RAGOPS_HTTP_ADDR", ":8080"),
		DatabaseURL:     getenv("RAGOPS_DATABASE_URL", ""),
		RedisAddr:       getenv("RAGOPS_REDIS_ADDR", "localhost:6379"),
		OpenAIAPIKey:    getenv("OPENAI_API_KEY", ""),
		OpenAIEmbModel:  getenv("RAGOPS_OPENAI_EMBED_MODEL", "text-embedding-3-small"),
		OpenAIChatModel: getenv("RAGOPS_OPENAI_CHAT_MODEL", "gpt-4o-mini"),
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
