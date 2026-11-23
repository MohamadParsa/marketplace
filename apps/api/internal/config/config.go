package config

import (
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	CacheHomeTTL, CacheProductTTL, CacheListingTTL, CacheSuggestionsTTL time.Duration
	RedisURL                                                            string
	PublicCacheEnabled                                                  bool
	CacheWarmupEnabled                                                  bool

	DatabaseURL       string
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
	return Config{CacheHomeTTL: cacheTTL("CACHE_HOME_TTL_SECONDS", 300), CacheProductTTL: cacheTTL("CACHE_PRODUCT_TTL_SECONDS", 600), CacheListingTTL: cacheTTL("CACHE_LISTING_TTL_SECONDS", 60), CacheSuggestionsTTL: cacheTTL("CACHE_SUGGESTIONS_TTL_SECONDS", 30), RedisURL: os.Getenv("REDIS_URL"), PublicCacheEnabled: os.Getenv("PUBLIC_CACHE_ENABLED") == "true", CacheWarmupEnabled: os.Getenv("CACHE_WARMUP_ENABLED") != "false", DatabaseURL: os.Getenv("DATABASE_URL"), WebRoot: os.Getenv("WEB_ROOT"), WebSurface: os.Getenv("WEB_SURFACE"), Address: address, Environment: env, ReviewDatabaseURL: os.Getenv("REVIEW_DATABASE_URL"), KavenegarAPIKey: os.Getenv("KAVENEGAR_API_KEY"), KavenegarTemplate: os.Getenv("KAVENEGAR_TEMPLATE"), KavenegarBaseURL: baseURL}
}

func cacheTTL(key string, maximum int) time.Duration {
	n, e := strconv.Atoi(os.Getenv(key))
	if e != nil || n < 1 || n > maximum {
		n = maximum
	}
	return time.Duration(n) * time.Second
}
