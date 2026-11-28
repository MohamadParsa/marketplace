package config

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	OperationLogDir                                                            string
	LogRetentionDays, BehaviorRetentionDays, LogMaxMB, LogFileMB, LogQueueSize int
	PaymentGatewayKey                                                          string
	PaymentPublicBaseURL                                                       string
	NeshanWebKey                                                               string
	NeshanServiceKey                                                           string
	AddressMapDailyLimit                                                       int
	CacheHomeTTL, CacheProductTTL, CacheListingTTL, CacheSuggestionsTTL        time.Duration
	RedisURL                                                                   string
	PublicCacheEnabled                                                         bool
	CacheWarmupEnabled                                                         bool

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
	logDir := os.Getenv("OPERATION_LOG_DIR")
	if logDir == "" {
		if base, err := os.UserCacheDir(); err == nil {
			logDir = filepath.Join(base, "Cherriva", "operation-logs")
		}
	}
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
	dailyLimit, _ := strconv.Atoi(os.Getenv("ADDRESS_MAP_DAILY_LIMIT"))
	if dailyLimit < 1 || dailyLimit > 100000 {
		dailyLimit = 1000
	}
	return Config{OperationLogDir: logDir, LogRetentionDays: logNumber("LOG_RETENTION_DAYS", 7, 3650), BehaviorRetentionDays: logNumber("LOG_BEHAVIOR_RETENTION_DAYS", 30, 3650), LogMaxMB: logNumber("LOG_MAX_MB", 500, 100000), LogFileMB: logNumber("LOG_FILE_MB", 8, 1000), LogQueueSize: logNumber("LOG_QUEUE_SIZE", 1024, 65536), PaymentPublicBaseURL: os.Getenv("PAYMENT_PUBLIC_BASE_URL"), PaymentGatewayKey: os.Getenv("PAYMENT_GATEWAY_ENCRYPTION_KEY"), NeshanWebKey: os.Getenv("NESHAN_WEB_KEY"), NeshanServiceKey: os.Getenv("NESHAN_SERVICE_KEY"), AddressMapDailyLimit: dailyLimit, CacheHomeTTL: cacheTTL("CACHE_HOME_TTL_SECONDS", 300), CacheProductTTL: cacheTTL("CACHE_PRODUCT_TTL_SECONDS", 600), CacheListingTTL: cacheTTL("CACHE_LISTING_TTL_SECONDS", 60), CacheSuggestionsTTL: cacheTTL("CACHE_SUGGESTIONS_TTL_SECONDS", 30), RedisURL: os.Getenv("REDIS_URL"), PublicCacheEnabled: os.Getenv("PUBLIC_CACHE_ENABLED") == "true", CacheWarmupEnabled: os.Getenv("CACHE_WARMUP_ENABLED") != "false", DatabaseURL: os.Getenv("DATABASE_URL"), WebRoot: os.Getenv("WEB_ROOT"), WebSurface: os.Getenv("WEB_SURFACE"), Address: address, Environment: env, ReviewDatabaseURL: os.Getenv("REVIEW_DATABASE_URL"), KavenegarAPIKey: os.Getenv("KAVENEGAR_API_KEY"), KavenegarTemplate: os.Getenv("KAVENEGAR_TEMPLATE"), KavenegarBaseURL: baseURL}
}

func cacheTTL(key string, maximum int) time.Duration {
	n, e := strconv.Atoi(os.Getenv(key))
	if e != nil || n < 1 || n > maximum {
		n = maximum
	}
	return time.Duration(n) * time.Second
}

func logNumber(key string, fallback, maximum int) int {
	n, err := strconv.Atoi(os.Getenv(key))
	if err != nil || n < 1 || n > maximum {
		return fallback
	}
	return n
}
