#!/usr/bin/env bash
set -euo pipefail
# Loopback-only disposable cache for Windows/WSL review. Never flush another Redis.
command -v redis-server >/dev/null || { echo 'Redis is unavailable in WSL.' >&2; exit 1; }
command -v redis-cli >/dev/null || exit 1
cache_runtime="${XDG_RUNTIME_DIR:-/tmp}/cherriva-review-redis"
mkdir -p "$cache_runtime"
if redis-cli -h 127.0.0.1 -p 56379 ping >/dev/null 2>&1; then
    if ! redis-cli -p 56379 INFO server | tr -d '\r' | grep -Fxq "config_file:$cache_runtime/redis.conf"; then
        echo 'Port 56379 belongs to another Redis; configure REDIS_URL explicitly.' >&2
        exit 1
    fi
    exit 0
fi
cat > "$cache_runtime/redis.conf" <<CONFIG
bind 127.0.0.1
protected-mode yes
port 56379
maxmemory 128mb
maxmemory-policy allkeys-lru
save ""
appendonly no
daemonize yes
pidfile $cache_runtime/redis.pid
logfile $cache_runtime/redis.log
CONFIG
redis-server "$cache_runtime/redis.conf"
redis-cli -p 56379 ping >/dev/null
