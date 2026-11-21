package config

import "os"

type Config struct {
	Address           string
	Environment       string
	ReviewDatabaseURL string
	KavenegarAPIKey   string
	KavenegarTemplate string
	KavenegarBaseURL  string
}

func Load() Config {
	address := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		address = ":" + port
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}
	baseURL := os.Getenv("KAVENEGAR_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.kavenegar.com"
	}
	return Config{Address: address, Environment: env, ReviewDatabaseURL: os.Getenv("REVIEW_DATABASE_URL"), KavenegarAPIKey: os.Getenv("KAVENEGAR_API_KEY"), KavenegarTemplate: os.Getenv("KAVENEGAR_TEMPLATE"), KavenegarBaseURL: baseURL}
}
