package config

import "os"

type Config struct {
	Address     string
	Environment string
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
	return Config{Address: address, Environment: env}
}
