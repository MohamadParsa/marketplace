package config

import (
	"net"
	"os"
)

type Config struct {
	WebRoot           string
	WebSurface        string
	Address           string
	Environment       string
	ReviewDatabaseURL string
	KavenegarAPIKey   string
	KavenegarTemplate string
	KavenegarBaseURL  string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := net.JoinHostPort(os.Getenv("BIND_HOST"), port)

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}
	baseURL := os.Getenv("KAVENEGAR_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.kavenegar.com"
	}
	return Config{WebRoot: os.Getenv("WEB_ROOT"), WebSurface: os.Getenv("WEB_SURFACE"), Address: address, Environment: env, ReviewDatabaseURL: os.Getenv("REVIEW_DATABASE_URL"), KavenegarAPIKey: os.Getenv("KAVENEGAR_API_KEY"), KavenegarTemplate: os.Getenv("KAVENEGAR_TEMPLATE"), KavenegarBaseURL: baseURL}
}
