package config

import (
	"testing"
	"time"
)

func TestPublicCacheConfigurationBounds(t *testing.T) {
	t.Setenv("PUBLIC_CACHE_ENABLED", "true")
	t.Setenv("CACHE_WARMUP_ENABLED", "")
	t.Setenv("CACHE_HOME_TTL_SECONDS", "301")
	t.Setenv("CACHE_PRODUCT_TTL_SECONDS", "45")
	t.Setenv("CACHE_LISTING_TTL_SECONDS", "-1")
	t.Setenv("CACHE_SUGGESTIONS_TTL_SECONDS", "bad")
	cfg := Load()
	if !cfg.PublicCacheEnabled || !cfg.CacheWarmupEnabled || cfg.CacheHomeTTL != 5*time.Minute || cfg.CacheProductTTL != 45*time.Second || cfg.CacheListingTTL != time.Minute || cfg.CacheSuggestionsTTL != 30*time.Second {
		t.Fatal("unexpected cache configuration")
	}
	t.Setenv("PUBLIC_CACHE_ENABLED", "false")
	t.Setenv("CACHE_WARMUP_ENABLED", "false")
	cfg = Load()
	if cfg.PublicCacheEnabled || cfg.CacheWarmupEnabled {
		t.Fatal("cache disable ignored")
	}
}
